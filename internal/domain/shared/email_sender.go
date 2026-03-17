package shared

import "context"

// EmailMessage defines a provider-agnostic outbound email payload.
type EmailMessage struct {
	To         string
	Subject    string
	Body       string
	TemplateID string
}

// EmailSender sends rendered email messages.
type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) error
}

