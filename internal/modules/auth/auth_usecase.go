package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/database"
	"github.com/semmidev/restful-template/internal/shared/errors"
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

func (s *Usecase) Register(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if err := in.Validate(); err != nil {
		return TokenPair{}, err
	}

	if _, err := s.users.GetByEmail(ctx, in.Email); err == nil {
		return TokenPair{}, errors.NewConflict("Email is already registered", errors.ErrConflict)
	}

	u, err := in.ToUser()
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.users.Create(ctx, u); err != nil {
		return TokenPair{}, errors.NewInternal("Failed to create user", err)
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) Login(ctx context.Context, in LoginInput) (TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return TokenPair{}, errors.NewUnauthorized("Invalid credentials", errors.ErrUnauthorized)
	}

	if !u.CheckPassword(in.Password) {
		return TokenPair{}, errors.NewUnauthorized("Invalid credentials", errors.ErrUnauthorized)
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.tokens.ParseRefresh(ctx, refreshToken)
	if err != nil {
		return TokenPair{}, errors.NewUnauthorized("Invalid refresh token", errors.ErrUnauthorized)
	}

	hash := hashToken(refreshToken)
	if err := s.tokenRepo.DeleteRefreshToken(ctx, hash); err != nil {
		return TokenPair{}, errors.NewUnauthorized("Invalid refresh token", errors.ErrUnauthorized)
	}

	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, errors.NewUnauthorized("Invalid user", errors.ErrUnauthorized)
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) issuePair(ctx context.Context, u *User) (TokenPair, error) {
	access, refresh, exp, refreshExp, err := s.tokens.GeneratePair(ctx, u.ID, u.Email)
	if err != nil {
		return TokenPair{}, errors.NewInternal("Failed to generate tokens", err)
	}

	hash := hashToken(refresh)
	if err := s.tokenRepo.StoreRefreshToken(ctx, u.ID, hash, time.Unix(refreshExp, 0)); err != nil {
		return TokenPair{}, errors.NewInternal("Failed to store session", err)
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
