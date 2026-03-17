package shared

import "context"

// FileURLResolver resolves resource references into externally accessible URLs.
// It intentionally keeps a neutral contract; scheme-specific rules are implemented
// by concrete subdomains/adapters (for example, files).
type FileURLResolver interface {
	Render(ctx context.Context, ref string) string
	RenderWithContext(ctx context.Context, ref string, defaultBucket string) string
	RenderBatch(ctx context.Context, refs []string) map[string]string
}
