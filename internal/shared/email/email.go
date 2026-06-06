package email

import "context"

// Message represents an outbound email with its recipients and content.
type Message struct {
	// To is the recipient's email address.
	To string
	// Subject is the email header's topic.
	Subject string
	// HTMLBody is the rich-text content of the email.
	HTMLBody string
}

// Sender is the interface for sending outbound communications.
type Sender interface {
	// Send dispatches an email message to its recipient.
	Send(ctx context.Context, msg Message) error
}
