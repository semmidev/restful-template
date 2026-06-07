package todos

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
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

// Validate checks business invariants on a Todo entity.
// Uses SafeError so the Code field is available for RFC 9457 responses —
// never raw errors.New which would lose structured context.
func (t *Todo) Validate() error {
	if t.Title == "" {
		return apperrors.NewInvalidInput("todo title cannot be empty", apperrors.ErrInvalidInput)
	}
	if t.Status != TodoStatusPending && t.Status != TodoStatusInProgress && t.Status != TodoStatusDone {
		return apperrors.NewInvalidInput("invalid todo status", apperrors.ErrInvalidInput)
	}
	return nil
}

// ChangeStatus updates the status field only; the caller is responsible
// for setting UpdatedAt on the aggregate after all mutations are applied.
func (t *Todo) ChangeStatus(status TodoStatus) {
	t.Status = status
}

func NewTodoEntity(in CreateTodoInput) *Todo {
	now := time.Now().UTC().Truncate(time.Microsecond)
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

// ApplyUpdate mutates the entity according to the given input.
//
// If UpdateMask is provided (AIP-134), only the listed fields are touched.
// Otherwise all non-nil fields in the input are applied.
// UpdatedAt is stamped exactly once at the end of the method — not inside
// ChangeStatus — so the timestamp is consistent regardless of how many
// fields are mutated.
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
	}
	// Stamp UpdatedAt exactly once regardless of which branch ran.
	t.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
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
// and avoids a redundant GetByID call inside the service.
type TodoService interface {
	List(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error)
	Create(ctx context.Context, input CreateTodoInput) (*Todo, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error)
	Update(ctx context.Context, existing *Todo, input UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
