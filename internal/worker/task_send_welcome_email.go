package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/semmidev/restful-template/internal/shared/asynqtask"
)

// ProcessTaskSendWelcomeEmail decodes the payload and logs that a welcome email is sent.
func (p *RedisTaskProcessor) ProcessTaskSendWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	var payload asynqtask.TaskPayloadSendWelcomeEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	p.logger.InfoContext(ctx, "Simulating sending welcome email",
		slog.String("user_id", payload.UserID.String()),
		slog.String("email", payload.Email),
	)

	// Here we could inject an email sender, generate HTML template, etc.
	// For this template, logging it is enough to demonstrate background task processing.

	p.logger.InfoContext(ctx, "successfully processed welcome email task",
		slog.String("user_id", payload.UserID.String()),
		slog.String("email", payload.Email),
	)
	return nil
}
