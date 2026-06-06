package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/password"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) CheckPassword(plain string) bool {
	ok, err := password.Verify(plain, u.PasswordHash)
	return err == nil && ok
}

type TokenClaims struct {
	UserID uuid.UUID
	Email  string
}

type TokenService interface {
	GeneratePair(ctx context.Context, userID uuid.UUID, email string) (access, refresh string, accessExp, refreshExp int64, err error)
	ParseAccess(ctx context.Context, token string) (*TokenClaims, error)
	ParseRefresh(ctx context.Context, token string) (*TokenClaims, error)
}

type TodoService interface {
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// TaskDistributor is the interface the auth usecase uses to enqueue background work.
// The concrete implementation lives in auth_distributor.go and wraps asynqtask.Distributor.
type TaskDistributor interface {
	DistributeTaskSendWelcomeEmail(ctx context.Context, payload *TaskPayloadSendWelcomeEmail) error
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
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
}

// AuthService is the interface that RegisterAuthRoutes consumes.
// *Service satisfies this interface implicitly — it exists to enable handler
// unit-testing with humatest mocks without a real database or JWT infrastructure.
type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (TokenPair, error)
	Login(ctx context.Context, in LoginInput) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
}
