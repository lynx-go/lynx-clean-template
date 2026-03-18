package grpc

import (
	"context"

	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/app"
)

type UsersService struct {
	apipb.UnimplementedUsersServiceServer
	uc *app.Users
}

func (svc *UsersService) GetUserProfile(ctx context.Context, req *apipb.GetUserProfileRequest) (*apipb.UserProfile, error) {
	return svc.uc.GetUserProfile(ctx, req)
}

func (svc *UsersService) UpdateMyProfile(ctx context.Context, req *apipb.UpdateMyProfileRequest) (*apipb.UserProfile, error) {
	return svc.uc.UpdateMyProfile(ctx, req)
}

func (svc *UsersService) GrantSuperAdmin(ctx context.Context, req *apipb.GrantSuperAdminRequest) (*apipb.UserProfile, error) {
	return svc.uc.GrantSuperAdmin(ctx, req)
}

func (svc *UsersService) RevokeSuperAdmin(ctx context.Context, req *apipb.RevokeSuperAdminRequest) (*apipb.UserProfile, error) {
	return svc.uc.RevokeSuperAdmin(ctx, req)
}

func NewUsersService(
	uc *app.Users,
) *UsersService {
	return &UsersService{uc: uc}
}
