package shared

import "context"

// Logger is the port for structured logging in the domain layer.
// Implementations are provided by the infrastructure layer.
type Logger interface {
	ErrorContext(ctx context.Context, msg string, err error, keysAndValues ...any)
	InfoContext(ctx context.Context, msg string, keysAndValues ...any)
	DebugContext(ctx context.Context, msg string, keysAndValues ...any)
}
