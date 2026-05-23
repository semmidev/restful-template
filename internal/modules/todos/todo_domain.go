package todos

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusDone       TodoStatus = "done"
)

type Todo struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Cover       *string
	Status      TodoStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t *Todo) Validate() error {
	// Simple validation
	return nil
}

type CreateTodoInput struct {
	UserID      uuid.UUID
	Title       string  `json:"title" required:"true"`
	Description string  `json:"description,omitempty"`
	Cover       *string `json:"cover,omitempty"`
}

type UpdateTodoInput struct {
	UserID      uuid.UUID
	ID          uuid.UUID
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Cover       *string     `json:"cover,omitempty"`
	Status      *TodoStatus `json:"status,omitempty"`
}

type ListTodosQuery struct {
	UserID  uuid.UUID
	Limit   int
	Offset  int
	Keyword string
	SortBy  string
	SortDir string
	Status  *TodoStatus `query:"status"`
	Search  string      `query:"search"`
}

type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	ListByUser(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
