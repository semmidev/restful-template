// Package asynqtask provides the Asynq-based implementation for distributing
// background tasks to Redis-backed queues.
package asynqtask

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Distributor implements a generic task distributor using Asynq and Redis.
// It serializes payloads and enqueues them into the appropriate queue.
// Module-specific dispatch methods live in each module's own distributor adapter.
type Distributor struct {
	client *asynq.Client
}

// NewDistributor constructs a new Asynq task distributor with the given Redis options.
func NewDistributor(redisOpt asynq.RedisClientOpt) *Distributor {
	client := asynq.NewClient(redisOpt)
	return &Distributor{
		client: client,
	}
}

// EnqueueTask serializes payload and enqueues an Asynq task by type and queue name.
// Module-specific distributor adapters call this to avoid polluting the shared package
// with domain knowledge.
func (d *Distributor) EnqueueTask(
	ctx context.Context,
	taskType string,
	payload any,
	queue string,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(taskType, jsonPayload,
		asynq.MaxRetry(5),
		asynq.Timeout(10*time.Second),
		asynq.Queue(queue),
	)

	_, err = d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueue task: %w", err)
	}

	return nil
}
