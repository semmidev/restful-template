package domain

import (
	"context"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type AuthUsecase interface {
	Register(ctx context.Context, in RegisterInput) (TokenPair, error)
	Login(ctx context.Context, in LoginInput) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
}

type CreateTodoInput struct {
	UserID      uuid.UUID
	Title       string
	Description *string
	Cover       *string
}

type UpdateTodoInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       *string
	Description *string
	Cover       *string
	Status      *TodoStatus
}

type ListTodosQuery struct {
	UserID  uuid.UUID
	Limit   int
	Offset  int
	Status  string // optional; empty means all statuses
	Keyword string // optional; searches title and description (case-insensitive)
	SortBy  string // optional; column to sort by
	SortDir string // optional; "asc" or "desc"
}

type TodoUsecase interface {
	Create(ctx context.Context, in CreateTodoInput) (*Todo, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	List(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Update(ctx context.Context, in UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}
