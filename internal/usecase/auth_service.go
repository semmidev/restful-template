package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/semmidev/restful-template/internal/domain"
	"github.com/semmidev/restful-template/internal/shared/password"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
)

// AuthService implements domain.AuthUsecase.
type AuthService struct {
	users     domain.UserRepository
	tokens    domain.TokenService
	tokenRepo domain.TokenRepository
}

func NewAuthService(users domain.UserRepository, tokens domain.TokenService, tokenRepo domain.TokenRepository) *AuthService {
	return &AuthService{users: users, tokens: tokens, tokenRepo: tokenRepo}
}

func (s *AuthService) Register(ctx context.Context, in domain.RegisterInput) (domain.TokenPair, error) {
	if in.Email == "" || in.Password == "" {
		return domain.TokenPair{}, domain.ErrInvalidInput
	}

	// Conflict check
	if _, err := s.users.FindByEmail(ctx, in.Email); err == nil {
		return domain.TokenPair{}, domain.ErrConflict
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return domain.TokenPair{}, err
	}

	u := &domain.User{
		ID:           uuidgen.New(),
		Email:        in.Email,
		PasswordHash: hash,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return domain.TokenPair{}, err
	}
	return s.issuePair(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, in domain.LoginInput) (domain.TokenPair, error) {
	u, err := s.users.FindByEmail(ctx, in.Email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	ok, err := password.Verify(in.Password, u.PasswordHash)
	if err != nil || !ok {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	return s.issuePair(ctx, u)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	claims, err := s.tokens.ParseRefresh(ctx, refreshToken)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	hash := hashToken(refreshToken)
	if err := s.tokenRepo.DeleteRefreshToken(ctx, hash); err != nil {
		// If it's not in the DB, it was already used or revoked -> treat as unauthorized
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	// Re-fetch user to ensure the account still exists
	u, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	return s.issuePair(ctx, u)
}

func (s *AuthService) issuePair(ctx context.Context, u *domain.User) (domain.TokenPair, error) {
	access, refresh, exp, refreshExp, err := s.tokens.GeneratePair(ctx, u.ID, u.Email)
	if err != nil {
		return domain.TokenPair{}, err
	}

	hash := hashToken(refresh)
	if err := s.tokenRepo.StoreRefreshToken(ctx, hash, u.ID, time.Unix(refreshExp, 0)); err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: exp}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
