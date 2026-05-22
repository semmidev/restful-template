package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TxManager handles executing operations within a database transaction context.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, tokenHash string, userID uuid.UUID, expiresAt time.Time) error
	// DeleteRefreshToken returns ErrNotFound if the token was not found
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

// TodoRepository is the driven port for todo persistence.
type TodoRepository interface {
	Create(ctx context.Context, t *Todo) error
	FindByID(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	// ListByUser returns paginated todos for a user.
	//   - status:  empty string = all statuses
	//   - keyword: empty string = no text filter; otherwise case-insensitive
	//              substring match against title and description
	ListByUser(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Update(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
}
