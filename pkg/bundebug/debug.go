package bundebug

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/uptrace/bun"
)

type Option func(*QueryHook)

// WithEnabled enables/disables the hook.
func WithEnabled(on bool) Option {
	return func(h *QueryHook) {
		h.enabled = on
	}
}

// WithVerbose configures the hook to log all queries
// (by default, only failed queries are logged).
func WithVerbose(on bool) Option {
	return func(h *QueryHook) {
		h.verbose = on
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(h *QueryHook) {
		h.logger = logger
	}
}

// FromEnv configures the hook using the environment variable value.
// For example, WithEnv("BUNDEBUG"):
//   - BUNDEBUG=0 - disables the hook.
//   - BUNDEBUG=1 - enables the hook.
//   - BUNDEBUG=2 - enables the hook and verbose mode.
func FromEnv(keys ...string) Option {
	if len(keys) == 0 {
		keys = []string{"BUNDEBUG"}
	}
	return func(h *QueryHook) {
		for _, key := range keys {
			if env, ok := os.LookupEnv(key); ok {
				h.enabled = env != "" && env != "0"
				h.verbose = env == "2"
				break
			}
		}
	}
}

type QueryHook struct {
	enabled bool
	verbose bool
	logger  *slog.Logger
}

var _ bun.QueryHook = (*QueryHook)(nil)

func NewQueryHook(opts ...Option) *QueryHook {
	h := &QueryHook{
		enabled: true,
		logger:  slog.Default(),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *QueryHook) BeforeQuery(
	ctx context.Context, event *bun.QueryEvent,
) context.Context {
	return ctx
}

func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !h.enabled {
		return
	}

	if !h.verbose {
		switch event.Err {
		case nil, sql.ErrNoRows, sql.ErrTxDone:
			return
		}
	}

	now := time.Now()
	dur := now.Sub(event.StartTime)

	args := []any{
		"elapsed", fmt.Sprintf(" %s ", dur.Round(time.Microsecond)),
		"query", event.Query,
	}

	if event.Err != nil {
		args = append(args, "error", event.Err.Error())
		h.logger.ErrorContext(ctx, "[bun] "+formatOperation(event), args...)
	} else {
		h.logger.DebugContext(ctx, "[bun] "+formatOperation(event), args...)
	}
}

func formatOperation(event *bun.QueryEvent) string {
	operation := event.Operation()
	return operationColor(operation).Sprintf("%s", operation)
}

func operationColor(operation string) *color.Color {
	switch operation {
	case "SELECT":
		return color.New(color.BgGreen, color.FgHiWhite)
	case "INSERT":
		return color.New(color.BgBlue, color.FgHiWhite)
	case "UPDATE":
		return color.New(color.BgYellow, color.FgHiBlack)
	case "DELETE":
		return color.New(color.BgMagenta, color.FgHiWhite)
	default:
		return color.New(color.BgWhite, color.FgHiBlack)
	}
}
