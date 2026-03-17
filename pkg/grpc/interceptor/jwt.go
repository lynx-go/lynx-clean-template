package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lynx-go/lynx-clean-template/pkg/jwtparser"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	// Validator defines an interface for token validation. This is satisfied by our auth service.
	Validator interface {
		ValidateToken(ctx context.Context, token string) (string, error)
	}

	AuthInterceptor struct {
		validator    Validator
		skipServices []string
	}
)

func NewAuthInterceptor(validator Validator, skipServices []string) (*AuthInterceptor, error) {
	if validator == nil {
		return nil, errors.New("validator cannot be nil")
	}
	return &AuthInterceptor{validator: validator, skipServices: skipServices}, nil
}

func (i *AuthInterceptor) UnaryAuthMiddleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if lo.Contains(i.skipServices, info.FullMethod) {
		return handler(ctx, req)
	}
	// get metadata object
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// extract token from authorization header
	tokens := md["authorization"]
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	token := tokens[0]
	token = strings.TrimPrefix(token, "Bearer ")
	// validate token and retrieve the userID
	userID, err := i.validator.ValidateToken(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
	}

	// add our user ID to the context, so we can use it in our RPC handler
	ctx = context.WithValue(ctx, "_USER", userID)

	// call our handler
	return handler(ctx, req)
}

type DefaultValidator struct {
	secretKey string
}

func (v *DefaultValidator) ValidateToken(ctx context.Context, token string) (string, error) {
	userId, _, _, _, _, _, ok := jwtparser.ParseToken([]byte(v.secretKey), token)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}
	return userId, nil
}

func NewDefaultValidator(
	secretKey string,
) Validator {
	return &DefaultValidator{secretKey: secretKey}
}
