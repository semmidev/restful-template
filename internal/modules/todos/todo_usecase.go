package todos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/cache"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/observability"
)

const todoCacheTTL = 5 * time.Minute

type Usecase struct {
	repo   TodoRepository
	cache  cache.CacheRepository
	tracer observability.Tracer
}

func NewTodo(repo TodoRepository, cache cache.CacheRepository, tracer observability.Tracer) *Usecase {
	return &Usecase{repo: repo, cache: cache, tracer: tracer}
}

func (s *Usecase) Create(ctx context.Context, in CreateTodoInput) (*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "todo.Create")
	defer span.End()

	t := NewTodoEntity(in)
	if err := t.Validate(); err != nil {
		return nil, apperrors.NewInvalidInput("Invalid todo data", err)
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, apperrors.NewInternal("Failed to create todo", err)
	}
	return t, nil
}

// Get retrieves a todo by ID using Redis as a read-through cache.
// The cache key includes userID to scope ownership and prevent cross-user reads.
func (s *Usecase) Get(ctx context.Context, userID, id uuid.UUID) (*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "todo.Get")
	defer span.End()

	key := todoCacheKey(userID, id)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var t Todo
		if jsonErr := json.Unmarshal([]byte(cached), &t); jsonErr == nil {
			return &t, nil
		}
	}

	t, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, apperrors.NewNotFound("The requested todo does not exist", err)
	}

	// Cache writes are best-effort: a failure here must not degrade availability.
	if b, jsonErr := json.Marshal(t); jsonErr == nil {
		_ = s.cache.Set(ctx, key, string(b), todoCacheTTL)
	}

	return t, nil
}

func (s *Usecase) List(ctx context.Context, q ListTodosQuery) ([]*Todo, int, error) {
	ctx, span := s.tracer.Start(ctx, "todo.List")
	defer span.End()

	q.Normalize()
	todos, count, err := s.repo.ListByUser(ctx, q)
	if err != nil {
		return nil, 0, apperrors.NewInternal("Failed to list todos", err)
	}
	return todos, count, nil
}

// Update applies the given input to the pre-loaded entity and persists it.
//
// Accepting a pre-loaded entity avoids a redundant repo.GetByID call: the
// handler already fetches the entity for ETag validation, so passing it here
// reduces a PATCH from 3 DB calls to 2.
func (s *Usecase) Update(ctx context.Context, existing *Todo, in UpdateTodoInput) (*Todo, error) {
	ctx, span := s.tracer.Start(ctx, "todo.Update")
	defer span.End()

	existing.ApplyUpdate(in)

	if err := existing.Validate(); err != nil {
		return nil, apperrors.NewInvalidInput("Invalid todo data", err)
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, apperrors.NewInternal("Failed to update todo", err)
	}

	key := todoCacheKey(existing.UserID, existing.ID)
	_ = s.cache.Delete(ctx, key)
	if b, jsonErr := json.Marshal(existing); jsonErr == nil {
		_ = s.cache.Set(ctx, key, string(b), todoCacheTTL)
	}

	return existing, nil
}

func (s *Usecase) Delete(ctx context.Context, userID, id uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "todo.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return apperrors.NewInternal("Failed to delete todo", err)
	}

	_ = s.cache.Delete(ctx, todoCacheKey(userID, id))
	return nil
}

func (s *Usecase) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "todo.DeleteAllByUserID")
	defer span.End()

	if err := s.repo.DeleteAllByUserID(ctx, userID); err != nil {
		return apperrors.NewInternal("Failed to delete all todos", err)
	}
	// Cache entries for individual todos by this user will naturally expire
	// within todoCacheTTL. We don't SCAN for user-keyed entries to avoid
	// O(n) Redis operations on what is a rare account-deletion path.
	return nil
}

// todoCacheKey returns the Redis key for a single todo entity.
func todoCacheKey(userID, todoID uuid.UUID) string {
	return fmt.Sprintf("todo:%s:%s", userID, todoID)
}
