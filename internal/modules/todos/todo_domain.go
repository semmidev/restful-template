package todos

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
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
	if t.Status != TodoStatusPending && t.Status != TodoStatusInProgress && t.Status != TodoStatusDone {
		return errors.New("invalid todo status")
	}
	return nil
}

func (t *Todo) ChangeStatus(status TodoStatus) {
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
}

func NewTodoEntity(in CreateTodoInput) *Todo {
	now := time.Now().UTC()
	return &Todo{
		ID:          uuidgen.New(),
		UserID:      in.UserID,
		Title:       in.Title,
		Description: in.Description,
		Cover:       in.Cover,
		Status:      TodoStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (t *Todo) ApplyUpdate(in UpdateTodoInput) {
	if len(in.UpdateMask) > 0 {
		for _, field := range in.UpdateMask {
			switch field {
			case "title":
				if in.Title != nil {
					t.Title = *in.Title
				} else {
					t.Title = ""
				}
			case "description":
				if in.Description != nil {
					t.Description = *in.Description
				} else {
					t.Description = ""
				}
			case "cover":
				if in.Cover != nil {
					if *in.Cover == "" {
						t.Cover = nil
					} else {
						t.Cover = in.Cover
					}
				} else {
					t.Cover = nil
				}
			case "status":
				if in.Status != nil {
					t.ChangeStatus(*in.Status)
				} else {
					t.ChangeStatus(TodoStatusPending)
				}
			}
		}
		t.UpdatedAt = time.Now().UTC()
	} else {
		if in.Title != nil {
			t.Title = *in.Title
		}
		if in.Description != nil {
			t.Description = *in.Description
		}
		if in.Cover != nil {
			if *in.Cover == "" {
				t.Cover = nil
			} else {
				t.Cover = in.Cover
			}
		}
		if in.Status != nil {
			t.ChangeStatus(*in.Status)
		}
		t.UpdatedAt = time.Now().UTC()
	}
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
	UpdateMask  []string
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

func (q *ListTodosQuery) Normalize() {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortDir == "" {
		q.SortDir = "desc"
	}
}

type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	ListByUser(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// TodoService is the interface consumed by the HTTP handler.
// It exists as an interface (not a concrete type) to allow handler unit-testing
// with mocks via humatest without requiring a real database or cache.
//
// Update accepts a pre-loaded *Todo so the handler's ETag fetch is reused
// and avoids a redundant GetByID call inside the usecase.
type TodoService interface {
	List(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Create(ctx context.Context, input CreateTodoInput) (*Todo, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	Update(ctx context.Context, existing *Todo, input UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
