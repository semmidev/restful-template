package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/database"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/observability"
)

type Usecase struct {
	users     UserRepository
	tokens    TokenService
	tokenRepo TokenRepository
	todos     TodoService
	txManager database.TxManager
	tracer    observability.Tracer
}

func NewAuth(users UserRepository, tokens TokenService, tokenRepo TokenRepository, todos TodoService, txManager database.TxManager, tracer observability.Tracer) *Usecase {
	return &Usecase{users: users, tokens: tokens, tokenRepo: tokenRepo, todos: todos, txManager: txManager, tracer: tracer}
}

// Register creates a new user and returns a token pair.
//
// point 9: The previous implementation did a GetByEmail pre-check before Insert,
// which was a TOCTOU race: two concurrent requests could both pass the check and
// then one fails with a raw DB unique-violation (returning 500). Now we attempt
// the INSERT directly and let the repository translate the unique constraint
// violation (pgErrCode 23505) into apperrors.ErrConflict → proper 409.
func (s *Usecase) Register(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if err := in.Validate(); err != nil {
		return TokenPair{}, err
	}

	u, err := in.ToUser()
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return TokenPair{}, apperrors.NewConflict("Email is already registered", err)
		}
		return TokenPair{}, apperrors.NewInternal("Failed to create user", err)
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) Login(ctx context.Context, in LoginInput) (TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid credentials", apperrors.ErrUnauthorized)
	}

	if !u.CheckPassword(in.Password) {
		return TokenPair{}, apperrors.NewUnauthorized("Invalid credentials", apperrors.ErrUnauthorized)
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
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

func (s *Usecase) issuePair(ctx context.Context, u *User) (TokenPair, error) {
	access, refresh, exp, refreshExp, err := s.tokens.GeneratePair(ctx, u.ID, u.Email)
	if err != nil {
		return TokenPair{}, apperrors.NewInternal("Failed to generate tokens", err)
	}

	hash := hashToken(refresh)
	if err := s.tokenRepo.StoreRefreshToken(ctx, u.ID, hash, time.Unix(refreshExp, 0)); err != nil {
		return TokenPair{}, apperrors.NewInternal("Failed to store session", err)
	}

	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: exp}, nil
}

func (s *Usecase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
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

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
