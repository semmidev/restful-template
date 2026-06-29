package auth

import (
	"context"

	"github.com/semmidev/restful-template/internal/shared/asynqtask"
)

// authDistributor wraps any asynqtask.TaskEnqueuer and implements auth.TaskDistributor.
// Depending on the interface (not *asynqtask.Distributor) keeps the auth module
// decoupled from the concrete Redis implementation and trivially mockable in tests.
type authDistributor struct {
	enqueuer asynqtask.TaskEnqueuer
}

// NewTaskDistributor returns a TaskDistributor backed by the given TaskEnqueuer.
// Pass asynqtask.NewDistributor(...) in production or a test double in unit tests.
func NewTaskDistributor(enqueuer asynqtask.TaskEnqueuer) TaskDistributor {
	return &authDistributor{enqueuer: enqueuer}
}

// DistributeTaskSendWelcomeEmail enqueues a welcome-email task into the critical queue.
func (a *authDistributor) DistributeTaskSendWelcomeEmail(ctx context.Context, payload *TaskPayloadSendWelcomeEmail) error {
	return a.enqueuer.EnqueueTask(ctx, TaskSendWelcomeEmail, payload, asynqtask.QueueCritical)
}
