package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/semmidev/restful-template/internal/shared/asynqtask"
)

// HandleSendWelcomeEmail returns an asynqtask.TaskHandler for the send_welcome_email task.
//
// The function signature accepts []byte (raw JSON payload) instead of *asynq.Task —
// the auth module never imports github.com/hibiken/asynq directly. All asynq-specific
// types are contained within internal/shared/asynqtask, which acts as the adapter layer.
//
// This also makes unit testing trivial: call the handler with a plain []byte,
// no need to construct an *asynq.Task.
func HandleSendWelcomeEmail(logger *slog.Logger) asynqtask.TaskHandler {
	return func(ctx context.Context, payload []byte) error {
		var p TaskPayloadSendWelcomeEmail
		if err := json.Unmarshal(payload, &p); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", asynqtask.ErrSkipRetry)
		}

		logger.InfoContext(ctx, "simulating sending welcome email",
			slog.String("user_id", p.UserID.String()),
			slog.String("email", p.Email),
		)

		// Here we could inject an email sender, generate HTML template, etc.
		// For this template, logging is enough to demonstrate background task processing.

		logger.InfoContext(ctx, "successfully processed welcome email task",
			slog.String("user_id", p.UserID.String()),
			slog.String("email", p.Email),
		)

		return nil
	}
}
