package shared

import "context"

// EventPublisher is the port for publishing domain events.
// Implementations are provided by the infrastructure layer.
type EventPublisher interface {
	Publish(ctx context.Context, topic, event string, data any) error
}
