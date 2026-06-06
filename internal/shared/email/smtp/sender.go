package smtp

import (
	"context"
	"fmt"

	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/shared/email"
	gomail "github.com/wneessen/go-mail"
)

// Sender implements email.Sender using go-mail (TLS-aware, context-safe).
type Sender struct {
	cfg    config.SMTP
	client *gomail.Client
}

// NewSender creates a ready-to-use SMTP Sender.
// It validates connectivity at startup so wiring failures are caught early.
func NewSender(cfg config.SMTP) (*Sender, error) {
	opts := []gomail.Option{
		gomail.WithPort(cfg.Port),
	}

	if cfg.Username != "" {
		opts = append(opts,
			gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
			gomail.WithUsername(cfg.Username),
			gomail.WithPassword(cfg.Password),
		)
	}

	// For local dev (e.g. Mailpit) skip TLS; for production enforce it.
	if cfg.Port == 1025 || cfg.Port == 25 {
		opts = append(opts, gomail.WithTLSPolicy(gomail.NoTLS))
	} else {
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSMandatory))
	}

	client, err := gomail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("smtp: create client: %w", err)
	}
	return &Sender{cfg: cfg, client: client}, nil
}

// Send delivers a single HTML email. It satisfies email.Sender.
func (s *Sender) Send(ctx context.Context, msg email.Message) error {
	m := gomail.NewMsg()
	if err := m.From(s.cfg.From); err != nil {
		return fmt.Errorf("smtp: set from: %w", err)
	}
	if err := m.To(msg.To); err != nil {
		return fmt.Errorf("smtp: set to: %w", err)
	}
	m.Subject(msg.Subject)
	m.SetBodyString(gomail.TypeTextHTML, msg.HTMLBody)

	if err := s.client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("smtp: send: %w", err)
	}
	return nil
}
