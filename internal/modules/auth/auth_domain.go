package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	jwtpkg "github.com/semmidev/restful-template/internal/shared/jwt"
	"github.com/semmidev/restful-template/internal/shared/password"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash *string   `json:"password_hash,omitempty"`
	GoogleID     *string   `json:"google_id,omitempty"`
	ActiveRole   string    `json:"active_role"`
	Roles        []string  `json:"roles"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) CheckPassword(plain string) bool {
	if u.PasswordHash == nil {
		return false
	}
	ok, err := password.Verify(plain, *u.PasswordHash)
	return err == nil && ok
}

type TokenService interface {
	GeneratePair(ctx context.Context, userID uuid.UUID, email string, activeRole string, roles []string) (access, refresh string, accessExp, refreshExp int64, err error)
	ParseAccess(ctx context.Context, token string) (*jwtpkg.TokenClaims, error)
	ParseRefresh(ctx context.Context, token string) (*jwtpkg.TokenClaims, error)
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
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	Update(ctx context.Context, u *User) error
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
	GoogleLogin(ctx context.Context, code string, codeVerifier string) (TokenPair, error)
	GoogleConfig() (clientID, redirectURI string)
	Logout(ctx context.Context, refreshToken string) error
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	SwitchRole(ctx context.Context, userID uuid.UUID, role string) (TokenPair, error)
}

