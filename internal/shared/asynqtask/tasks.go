package asynqtask

import (
	"github.com/google/uuid"
)

const (
	// TaskSendWelcomeEmail is the unique identifier for the welcome email task.
	TaskSendWelcomeEmail = "task:send_welcome_email"

	// QueueCritical handles high-priority tasks.
	QueueCritical = "critical"
	// QueueDefault handles standard-priority tasks.
	QueueDefault = "default"
)

// TaskPayloadSendWelcomeEmail defines the payload for sending a welcome email.
type TaskPayloadSendWelcomeEmail struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}
