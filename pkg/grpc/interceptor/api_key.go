package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// APIKeyChecker defines the interface for checking API keys
type APIKeyChecker interface {
	GetAPIKey(ctx context.Context) (string, error)
}

// APIKeyInterceptor validates X-API-Key header for developer API services
type APIKeyInterceptor struct {
	checker    APIKeyChecker
	pathPrefix string
}

func NewAPIKeyInterceptor(checker APIKeyChecker, pathPrefix string) *APIKeyInterceptor {
	return &APIKeyInterceptor{checker: checker, pathPrefix: pathPrefix}
}

// UnaryAPIKeyMiddleware checks for valid API key in X-API-Key header for developer API services
func (i *APIKeyInterceptor) UnaryAPIKeyMiddleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// Only apply API key check to developer services
	if !strings.HasPrefix(info.FullMethod, i.pathPrefix) {
		return handler(ctx, req)
	}

	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// Extract API key from x-api-key header
	apiKeys := md["x-api-key"]
	if len(apiKeys) == 0 {
		return nil, status.Error(codes.Unauthenticated, "API key is not provided")
	}

	providedKey := apiKeys[0]

	// Get the expected API key from runtime vars
	expectedKey, err := i.checker.GetAPIKey(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to verify API key")
	}

	// Validate the API key
	if providedKey != expectedKey {
		return nil, status.Error(codes.PermissionDenied, "invalid API key")
	}

	// API key is valid, proceed with the request
	return handler(ctx, req)
}
