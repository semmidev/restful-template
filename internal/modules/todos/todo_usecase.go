package todos

import (
	"context"
	"time"

	"github.com/google/uuid"
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
		return nil, err
	}

	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "Todo.validateAndFormat")
		span.End()
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Usecase) Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error) {
	return s.repo.GetByID(ctx, userID, id)
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
	return s.repo.ListByUser(ctx, q)
}

func (s *Usecase) Update(ctx context.Context, in UpdateTodoInput) (*Todo, error) {
	t, err := s.repo.GetByID(ctx, in.UserID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Description != nil {
		t.Description = *in.Description
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

func (s *Usecase) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *Usecase) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteAllByUserID(ctx, userID)
}
