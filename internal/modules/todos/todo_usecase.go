package todos

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"

	"github.com/semmidev/restful-template/internal/shared/cache"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
)

type Usecase struct {
	repo   TodoRepository
	cache  cache.CacheRepository
	tracer observability.Tracer
}

func NewTodo(repo TodoRepository, cache cache.CacheRepository, tracer observability.Tracer) *Usecase {
	return &Usecase{repo: repo, cache: cache, tracer: tracer}
}

func (s *Usecase) Create(ctx context.Context, in CreateTodoInput) (*Todo, error) {
	now := time.Now().UTC()
	t := &Todo{
		ID:          uuidgen.New(),
		UserID:      in.UserID,
		Title:       in.Title,
		Description: in.Description,
		Cover:       in.Cover,
		Status:      TodoStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := t.Validate(); err != nil {
		return nil, apperrors.NewInvalidInput("Invalid todo data", err)
	}

	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "Todo.validateAndFormat")
		span.End()
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, apperrors.NewInternal("Failed to create todo", err)
	}
	return t, nil
}

func (s *Usecase) Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error) {
	t, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, apperrors.NewNotFound("The requested todo does not exist", err)
	}
	return t, nil
}

func (s *Usecase) List(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortDir == "" {
		q.SortDir = "desc"
	}
	todos, count, err := s.repo.ListByUser(ctx, q)
	if err != nil {
		return nil, 0, apperrors.NewInternal("Failed to list todos", err)
	}
	return todos, count, nil
}

func (s *Usecase) Update(ctx context.Context, in UpdateTodoInput) (*Todo, error) {
	t, err := s.repo.GetByID(ctx, in.UserID, in.ID)
	if err != nil {
		return nil, apperrors.NewNotFound("The requested todo does not exist", err)
	}

	t.UpdateDetails(in.Title, in.Description, in.Cover)
	if in.Status != nil {
		t.ChangeStatus(*in.Status)
	}

	if err := t.Validate(); err != nil {
		return nil, apperrors.NewInvalidInput("Invalid todo data", err)
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, apperrors.NewInternal("Failed to update todo", err)
	}
	return t, nil
}

func (s *Usecase) Delete(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return apperrors.NewInternal("Failed to delete todo", err)
	}
	return nil
}

func (s *Usecase) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.DeleteAllByUserID(ctx, userID); err != nil {
		return apperrors.NewInternal("Failed to delete all todos", err)
	}
	return nil
}
