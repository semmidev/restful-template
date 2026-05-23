package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TokenClaims struct {
	UserID uuid.UUID
	Email  string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RegisterInput struct {
	Email    string `json:"email" format:"email" required:"true"`
	Password string `json:"password" minLength:"8" required:"true"`
}

type LoginInput struct {
	Email    string `json:"email" format:"email" required:"true"`
	Password string `json:"password" required:"true"`
}

type TokenService interface {
	GeneratePair(ctx context.Context, userID uuid.UUID, email string) (access, refresh string, accessExp, refreshExp int64, err error)
	ParseAccess(ctx context.Context, token string) (*TokenClaims, error)
	ParseRefresh(ctx context.Context, token string) (*TokenClaims, error)
}

type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type TodoService interface {
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}
