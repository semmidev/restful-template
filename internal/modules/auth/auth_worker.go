package auth

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/semmidev/restful-template/internal/shared/asynqtask"
	"github.com/semmidev/restful-template/internal/shared/email"
)

//go:embed templates/welcome.html
var welcomeHTML string

var welcomeTmpl = template.Must(template.New("welcome").Parse(welcomeHTML))

type WelcomeData struct {
	UserName string
	AppURL   string
}

// AuthWorker holds dependencies for processing background tasks related to the auth domain.
type AuthWorker struct {
	logger *slog.Logger
	sender email.Sender
	appURL string
}

// NewAuthWorker creates a new AuthWorker.
func NewAuthWorker(logger *slog.Logger, sender email.Sender, appURL string) *AuthWorker {
	return &AuthWorker{
		logger: logger,
		sender: sender,
		appURL: appURL,
	}
}

// HandleSendWelcomeEmail returns an asynqtask.TaskHandler for the send_welcome_email task.
//
// The function signature accepts []byte (raw JSON payload) instead of *asynq.Task —
// the auth module never imports github.com/hibiken/asynq directly. All asynq-specific
// types are contained within internal/shared/asynqtask, which acts as the adapter layer.
//
// This also makes unit testing trivial: call the handler with a plain []byte,
// no need to construct an *asynq.Task.
func (w *AuthWorker) HandleSendWelcomeEmail() asynqtask.TaskHandler {
	return func(ctx context.Context, payload []byte) error {
		var p TaskPayloadSendWelcomeEmail
		if err := json.Unmarshal(payload, &p); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", asynqtask.ErrSkipRetry)
		}

		w.logger.InfoContext(ctx, "sending welcome email",
			slog.String("user_id", p.UserID.String()),
			slog.String("email", p.Email),
		)

		var buf bytes.Buffer
		data := WelcomeData{
			UserName: p.Email, // Fallback to email since User model only has email
			AppURL:   w.appURL,
		}

		if err := welcomeTmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to render welcome template: %w", err)
		}

		msg := email.Message{
			To:       p.Email,
			Subject:  "Welcome to Todo App!",
			HTMLBody: buf.String(),
		}

		if err := w.sender.Send(ctx, msg); err != nil {
			// If SMTP fails, the error is returned so Asynq can retry it.
			return fmt.Errorf("smtp send failed: %w", err)
		}

		w.logger.InfoContext(ctx, "successfully sent welcome email",
			slog.String("user_id", p.UserID.String()),
			slog.String("email", p.Email),
		)

		return nil
	}
}
