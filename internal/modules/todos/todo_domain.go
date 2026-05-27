package todos

import (
	"context"
	"errors"
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
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Cover       *string    `json:"cover"`
	Status      TodoStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (t *Todo) Validate() error {
	if t.Title == "" {
		return errors.New("todo title cannot be empty")
	}
	return nil
}

func (t *Todo) ChangeStatus(status TodoStatus) {
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
}

func (t *Todo) UpdateDetails(title, desc *string, cover *string) {
	if title != nil {
		t.Title = *title
	}
	if desc != nil {
		t.Description = *desc
	}
	if cover != nil {
		t.Cover = cover
	}
	t.UpdatedAt = time.Now().UTC()
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

// TodoService is the interface that RegisterTodoRoutes consumes.
// *Usecase satisfies this interface implicitly — it exists solely to enable
// handler unit-testing with mocks (e.g. via humatest) without a real database.
type TodoService interface {
	List(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Create(ctx context.Context, input CreateTodoInput) (*Todo, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	Update(ctx context.Context, input UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}
