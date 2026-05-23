package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/semmidev/restful-template/internal/shared/password"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
)

type Usecase struct {
	users     UserRepository
	tokens    TokenService
	tokenRepo TokenRepository
	todos     TodoService
	txManager TxManager
	tracer    observability.Tracer
}

func NewAuth(users UserRepository, tokens TokenService, tokenRepo TokenRepository, todos TodoService, txManager TxManager, tracer observability.Tracer) *Usecase {
	return &Usecase{users: users, tokens: tokens, tokenRepo: tokenRepo, todos: todos, txManager: txManager, tracer: tracer}
}

func (s *Usecase) Register(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if in.Email == "" || in.Password == "" {
		return TokenPair{}, errors.ErrInvalidInput
	}

	if _, err := s.users.GetByEmail(ctx, in.Email); err == nil {
		return TokenPair{}, errors.ErrConflict
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return TokenPair{}, err
	}

	u := &User{
		ID:           uuidgen.New(),
		Email:        in.Email,
		PasswordHash: hash,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return TokenPair{}, err
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) Login(ctx context.Context, in LoginInput) (TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return TokenPair{}, errors.ErrUnauthorized
	}

	ok, err := password.Verify(in.Password, u.PasswordHash)
	if err != nil || !ok {
		return TokenPair{}, errors.ErrUnauthorized
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.tokens.ParseRefresh(ctx, refreshToken)
	if err != nil {
		return TokenPair{}, errors.ErrUnauthorized
	}

	hash := hashToken(refreshToken)
	if err := s.tokenRepo.DeleteRefreshToken(ctx, hash); err != nil {
		return TokenPair{}, errors.ErrUnauthorized
	}

	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, errors.ErrUnauthorized
	}
	return s.issuePair(ctx, u)
}

func (s *Usecase) issuePair(ctx context.Context, u *User) (TokenPair, error) {
	access, refresh, exp, refreshExp, err := s.tokens.GeneratePair(ctx, u.ID, u.Email)
	if err != nil {
		return TokenPair{}, err
	}

	hash := hashToken(refresh)
	if err := s.tokenRepo.StoreRefreshToken(ctx, u.ID, hash, time.Unix(refreshExp, 0)); err != nil {
		return TokenPair{}, err
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
