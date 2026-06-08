package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/observability"
)

type Service struct {
	users       UserRepository
	tokens      TokenService
	tokenRepo   TokenRepository
	todos       TodoService
	txManager   database.TxManager
	tracer      observability.Tracer
	distributor TaskDistributor
	googleCfg   config.Google
}

func NewAuthService(
	users UserRepository,
	tokens TokenService,
	tokenRepo TokenRepository,
	todos TodoService,
	txManager database.TxManager,
	tracer observability.Tracer,
	distributor TaskDistributor,
	googleCfg config.Google,
) *Service {
	return &Service{
		users:       users,
		tokens:      tokens,
		tokenRepo:   tokenRepo,
		todos:       todos,
		txManager:   txManager,
		tracer:      tracer,
		distributor: distributor,
		googleCfg:   googleCfg,
	}
}

// Register inserts the user directly and lets the repository translate a unique
// constraint violation (pg code 23505) into ErrConflict.
// A pre-check GetByEmail before INSERT would be a TOCTOU race: two concurrent
// requests could both pass the check, then one fails with a raw 500.
func (s *Service) Register(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if err := in.Validate(); err != nil {
		return TokenPair{}, err
	}

	u, err := in.ToUser()
	if err != nil {
		return TokenPair{}, err
	}

	u.ActiveRole = "user"
	u.Roles = []string{"user"}
	if strings.Contains(strings.ToLower(u.Email), "admin") {
		u.Roles = append(u.Roles, "admin")
	}

	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return TokenPair{}, apperrors.NewConflict("Email is already registered", err)
		}
		return TokenPair{}, apperrors.NewInternal("Failed to create user", err)
	}

	observability.AuthRegistrationsTotal.Inc()

	// Dispatch welcome email task asynchronously (best-effort — never fail the request).
	if s.distributor != nil {
		_ = s.distributor.DistributeTaskSendWelcomeEmail(ctx, &TaskPayloadSendWelcomeEmail{
			UserID: u.ID,
			Email:  u.Email,
		})
	}

	return s.issuePair(ctx, u)
}

func (s *Service) Login(ctx context.Context, in LoginInput) (TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid credentials", apperrors.ErrUnauthorized)
	}

	if !u.CheckPassword(in.Password) {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid credentials", apperrors.ErrUnauthorized)
	}
	return s.issuePair(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.tokens.ParseRefresh(ctx, refreshToken)
	if err != nil {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid refresh token", apperrors.ErrUnauthorized)
	}

	hash := hashToken(refreshToken)
	if err := s.tokenRepo.DeleteRefreshToken(ctx, hash); err != nil {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid refresh token", apperrors.ErrUnauthorized)
	}

	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid user", apperrors.ErrUnauthorized)
	}
	return s.issuePair(ctx, u)
}

func (s *Service) issuePair(ctx context.Context, u *User) (TokenPair, error) {
	access, refresh, exp, refreshExp, err := s.tokens.GeneratePair(ctx, u.ID, u.Email, u.ActiveRole, u.Roles)
	if err != nil {
		return TokenPair{}, apperrors.NewInternal("Failed to generate tokens", err)
	}

	hash := hashToken(refresh)
	if err := s.tokenRepo.StoreRefreshToken(ctx, u.ID, hash, time.Unix(refreshExp, 0)); err != nil {
		return TokenPair{}, apperrors.NewInternal("Failed to store session", err)
	}

	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    exp,
		UserID:       u.ID,
		UserEmail:    u.Email,
		ActiveRole:   u.ActiveRole,
		Roles:        u.Roles,
	}, nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.todos.DeleteAllByUserID(txCtx, userID); err != nil {
			return err
		}
		if err := s.users.Delete(txCtx, userID); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	ctx, span := s.tracer.Start(ctx, "auth.Logout")
	defer span.End()

	if refreshToken == "" {
		return nil
	}

	hash := hashToken(refreshToken)
	if err := s.tokenRepo.DeleteRefreshToken(ctx, hash); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return apperrors.NewInternal("Failed to invalidate session", err)
	}
	return nil
}

func (s *Service) SwitchRole(ctx context.Context, userID uuid.UUID, targetRole string) (TokenPair, error) {
	ctx, span := s.tracer.Start(ctx, "auth.SwitchRole")
	defer span.End()

	var pair TokenPair
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		u, err := s.users.GetByID(txCtx, userID)
		if err != nil {
			return err
		}

		hasRole := false
		for _, r := range u.Roles {
			if r == targetRole {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return apperrors.NewForbidden("You do not have the requested role", apperrors.ErrForbidden)
		}

		u.ActiveRole = targetRole
		u.UpdatedAt = time.Now()
		if err := s.users.Update(txCtx, u); err != nil {
			return err
		}

		pair, err = s.issuePair(txCtx, u)
		return err
	})

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return TokenPair{}, apperrors.NewNotFound("User not found", err)
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			return TokenPair{}, err
		}
		return TokenPair{}, apperrors.NewInternal("Failed to switch role", err)
	}

	return pair, nil
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func (s *Service) exchangeGoogleCode(ctx context.Context, code, verifier string) (*googleUserInfo, error) {
	data := url.Values{}
	data.Set("client_id", s.googleCfg.ClientID)
	data.Set("client_secret", s.googleCfg.ClientSecret)
	data.Set("code", code)
	data.Set("code_verifier", verifier)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", s.googleCfg.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google token exchange failed: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenRes googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google userinfo failed: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *Service) GoogleLogin(ctx context.Context, code string, codeVerifier string) (TokenPair, error) {
	ctx, span := s.tracer.Start(ctx, "auth.GoogleLogin")
	defer span.End()

	if code == "" || codeVerifier == "" {
		return TokenPair{}, apperrors.NewInvalidInput("Code and code_verifier are required", nil)
	}

	userInfo, err := s.exchangeGoogleCode(ctx, code, codeVerifier)
	if err != nil {
		return TokenPair{}, apperrors.NewUnauthorized("Google authentication failed", err)
	}

	if userInfo.Email == "" || userInfo.Sub == "" {
		return TokenPair{}, apperrors.NewUnauthorized("Google returned invalid user info", nil)
	}

	var u *User
	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		var txErr error
		u, txErr = s.users.GetByGoogleID(txCtx, userInfo.Sub)
		if txErr == nil {
			return nil
		}

		if !errors.Is(txErr, apperrors.ErrNotFound) {
			return txErr
		}

		u, txErr = s.users.GetByEmail(txCtx, userInfo.Email)
		if txErr == nil {
			u.GoogleID = &userInfo.Sub
			u.UpdatedAt = time.Now()
			if txErr = s.users.Update(txCtx, u); txErr != nil {
				return txErr
			}
			return nil
		}

		if !errors.Is(txErr, apperrors.ErrNotFound) {
			return txErr
		}

		u = &User{
			ID:         uuid.New(),
			Email:      userInfo.Email,
			GoogleID:   &userInfo.Sub,
			ActiveRole: "user",
			Roles:      []string{"user"},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if strings.Contains(strings.ToLower(u.Email), "admin") {
			u.Roles = append(u.Roles, "admin")
		}
		if txErr = s.users.Create(txCtx, u); txErr != nil {
			return txErr
		}

		if s.distributor != nil {
			_ = s.distributor.DistributeTaskSendWelcomeEmail(txCtx, &TaskPayloadSendWelcomeEmail{
				UserID: u.ID,
				Email:  u.Email,
			})
		}

		return nil
	})

	if err != nil {
		return TokenPair{}, apperrors.NewInternal("Failed to process Google login", err)
	}

	return s.issuePair(ctx, u)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Service) GoogleConfig() (clientID, redirectURI string) {
	return s.googleCfg.ClientID, s.googleCfg.RedirectURI
}
