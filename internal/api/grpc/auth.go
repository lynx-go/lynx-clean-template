package grpc

import (
	"context"

	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewAuthService(
	uc *app.Account,
) *AuthService {
	return &AuthService{uc: uc}
}

type AuthService struct {
	uc *app.Account
	apipb.UnimplementedAuthServiceServer
}

func (a *AuthService) Token(ctx context.Context, req *apipb.TokenRequest) (*apipb.TokenResponse, error) {
	switch req.GrantType {
	case app.GrantTypePassword:
		return a.uc.AuthorizeByPassword(ctx, req)
	case app.GrantTypeRefreshToken:
		return a.uc.RefreshToken(ctx, req)

	default:
		return nil, status.Error(codes.InvalidArgument, "invalid grant_type")
	}
}

func (a *AuthService) SignUp(ctx context.Context, req *apipb.SignUpRequest) (*apipb.SignUpResponse, error) {
	return a.uc.SignUp(ctx, req)
}

var _ apipb.AuthServiceServer = new(AuthService)
