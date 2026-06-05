package asynqtask

import "context"

const (
	// QueueCritical handles high-priority tasks.
	QueueCritical = "critical"
	// QueueDefault handles standard-priority tasks.
	QueueDefault = "default"
)

// TaskEnqueuer is the interface module-level distributor adapters depend on.
// *Distributor satisfies this automatically; modules never hold a concrete *Distributor,
// which keeps them decoupled from the Redis implementation and easy to mock in tests.
type TaskEnqueuer interface {
	EnqueueTask(ctx context.Context, taskType string, payload any, queue string) error
}
