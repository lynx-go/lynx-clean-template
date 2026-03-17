package grpc

import (
	"context"

	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/app"
)

type GroupsService struct {
	apipb.UnimplementedGroupsServiceServer
	uc *app.Groups
}

func (svc *GroupsService) CreateGroup(ctx context.Context, req *apipb.CreateGroupRequest) (*apipb.Group, error) {
	return svc.uc.Create(ctx, req)
}

func (svc *GroupsService) GetGroup(ctx context.Context, req *apipb.GetGroupRequest) (*apipb.Group, error) {
	return svc.uc.Get(ctx, req)
}

func (svc *GroupsService) ListGroups(ctx context.Context, req *apipb.ListGroupsRequest) (*apipb.ListGroupsResponse, error) {
	return svc.uc.List(ctx, req)
}

func (svc *GroupsService) UpdateGroup(ctx context.Context, req *apipb.UpdateGroupRequest) (*apipb.Group, error) {
	return svc.uc.Update(ctx, req)
}

func (svc *GroupsService) DeleteGroup(ctx context.Context, req *apipb.DeleteGroupRequest) (*apipb.DeleteGroupResponse, error) {
	return svc.uc.Delete(ctx, req)
}

func NewGroupsService(
	uc *app.Groups,
) *GroupsService {
	return &GroupsService{uc: uc}
}
