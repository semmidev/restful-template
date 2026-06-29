package auth

import "github.com/google/uuid"

const (
	// TaskSendWelcomeEmail is the unique asynq task type for sending a welcome email.
	// Owned by the auth module because it is the only producer and consumer of this task.
	TaskSendWelcomeEmail = "task:send_welcome_email"
)

// TaskPayloadSendWelcomeEmail carries the data needed to send a welcome email
// to a newly registered user.
type TaskPayloadSendWelcomeEmail struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}
