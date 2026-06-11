package todos

import (
	"context"
	"log/slog"
	"time"

	"github.com/semmidev/restful-template/internal/shared/cache"
	"golang.org/x/sync/errgroup"
)

type TodoJob struct {
	repo   TodoRepository
	cache  cache.CacheRepository
	logger *slog.Logger
}

func NewTodoJob(repo TodoRepository, cache cache.CacheRepository, logger *slog.Logger) *TodoJob {
	return &TodoJob{repo: repo, cache: cache, logger: logger}
}

// EscalateUrgency checks for tasks close to their due date and escalates their urgency.
func (j *TodoJob) EscalateUrgency() {
	ctx := context.Background()
	threshold := time.Now().UTC().Add(24 * time.Hour)

	updated, err := j.repo.EscalateUrgency(ctx, threshold)
	if err != nil {
		j.logger.Error("todo urgency escalation failed", "err", err)
		return
	}

	if len(updated) > 0 {
		g, groupCtx := errgroup.WithContext(ctx)
		g.SetLimit(10) // Concurrency limit of 10 parallel deletes to avoid connection exhaustion

		for _, t := range updated {
			t := t
			g.Go(func() error {
				key := todoCacheKey(t.UserID, t.ID)
				_ = j.cache.Delete(groupCtx, key)
				return nil
			})
		}
		_ = g.Wait()

		j.logger.Info("escalated urgency for close due-date todos", "count", len(updated))
	}
}
