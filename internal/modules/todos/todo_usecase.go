package todos

import (
	"context"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/cache"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/observability"
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
	t := NewTodoEntity(in)
	if err := t.Validate(); err != nil {
		return nil, apperrors.NewInvalidInput("Invalid todo data", err)
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
	q.Normalize()
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

	t.ApplyUpdate(in)

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
