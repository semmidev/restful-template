package todos

import (
	"context"
	"log/slog"
	"time"

	"github.com/semmidev/restful-template/internal/shared/cache"
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
		for _, t := range updated {
			// Invalidate Redis cache so subsequent reads see the updated urgency
			key := todoCacheKey(t.UserID, t.ID)
			_ = j.cache.Delete(ctx, key)
		}
		j.logger.Info("escalated urgency for close due-date todos", "count", len(updated))
	}
}
