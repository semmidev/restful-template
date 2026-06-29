package todo

import (
	"time"

	"github.com/google/uuid"
)

type CreateTodoInput struct {
	UserID      uuid.UUID
	Title       string     `json:"title" required:"true"`
	Description string     `json:"description,omitempty"`
	Cover       *string    `json:"cover,omitempty"`
	Importance  bool       `json:"importance"`
	Urgency     bool       `json:"urgency"`
	DueAt       *time.Time `json:"due_at,omitempty"`
}

type UpdateTodoInput struct {
	UserID      uuid.UUID
	ID          uuid.UUID
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Cover       *string     `json:"cover,omitempty"`
	Status      *TodoStatus `json:"status,omitempty"`
	Importance  *bool       `json:"importance,omitempty"`
	Urgency     *bool       `json:"urgency,omitempty"`
	DueAt       *time.Time  `json:"due_at,omitempty"`
	ClearDueAt  bool        `json:"clear_due_at,omitempty"`
	UpdateMask  []string
}

type ListTodosQuery struct {
	UserID   uuid.UUID
	Limit    int
	Offset   int
	Keyword  string
	SortBy   string
	SortDir  string
	Status   *TodoStatus `query:"status"`
	Archived bool        `query:"archived"`
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
