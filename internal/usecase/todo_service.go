package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/domain"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
)

// TodoService implements domain.TodoUsecase.
type TodoService struct {
	repo   domain.TodoRepository
	cache  domain.CacheRepository
	tracer domain.Tracer
}

func NewTodoService(repo domain.TodoRepository, cache domain.CacheRepository, tracer domain.Tracer) *TodoService {
	return &TodoService{repo: repo, cache: cache, tracer: tracer}
}

func (s *TodoService) Create(ctx context.Context, in domain.CreateTodoInput) (*domain.Todo, error) {
	now := time.Now().UTC()
	t := &domain.Todo{
		ID:          uuidgen.New(), // UUID v7 — time-ordered, sortable
		UserID:      in.UserID,
		Title:       in.Title,
		Description: in.Description,
		Cover:       in.Cover,
		Status:      domain.TodoPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}

	// Example: tracing a heavy internal block manually
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "TodoService.validateAndFormat")
		// perform some heavy logic or formatting
		span.End()
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TodoService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Todo, error) {
	return s.repo.FindByID(ctx, userID, id)
}

func (s *TodoService) List(ctx context.Context, q domain.ListTodosQuery) ([]*domain.Todo, int, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortDir == "" {
		q.SortDir = "desc"
	}
	return s.repo.ListByUser(ctx, q)
}

func (s *TodoService) Update(ctx context.Context, in domain.UpdateTodoInput) (*domain.Todo, error) {
	t, err := s.repo.FindByID(ctx, in.UserID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Description != nil {
		t.Description = in.Description
	}
	if in.Cover != nil {
		t.Cover = in.Cover
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	t.UpdatedAt = time.Now().UTC()
	if err := t.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TodoService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}
