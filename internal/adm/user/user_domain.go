package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash *string   `json:"-" db:"password_hash"`
	ActiveRole   string    `json:"active_role" db:"active_role"`
	Roles        []string  `json:"roles" db:"-"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int, search string, sortBy, sortDir string) ([]*User, int, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	Create(ctx context.Context, in CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, page, perPage int, search string, sortBy, sortDir string) ([]*User, int, error)
	Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
