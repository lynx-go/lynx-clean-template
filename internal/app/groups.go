package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/domain/groups/repo"
	"github.com/lynx-go/lynx-clean-template/internal/domain/shared/consts"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/contexts"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	apierrors "github.com/lynx-go/lynx-clean-template/pkg/errors"
	"github.com/lynx-go/lynx-clean-template/pkg/icons"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/lynx-go/x/log"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Groups struct {
	groupsRepo repo.GroupsRepo
}

func NewGroups(
	groupsRepo repo.GroupsRepo,
) *Groups {
	return &Groups{
		groupsRepo: groupsRepo,
	}
}

func (uc *Groups) List(ctx context.Context, req *apipb.ListGroupsRequest) (*apipb.ListGroupsResponse, error) {
	params, err := crud.ParseListParams(req.PageSize, req.PageToken, req.Filter, req.OrderBy)
	if err != nil {
		return nil, err
	}
	currentUser, _ := contexts.UserID(ctx)
	if !currentUser.IsValid() {
		return nil, apierrors.New(401, "You need to be logged in")
	}
	list, total, nextPageToken, err := uc.groupsRepo.ListByMemberID(ctx, currentUser, params)
	if err != nil {
		return nil, err
	}

	return &apipb.ListGroupsResponse{
		Items: lo.Map(list, func(v repo.Group, i int) *apipb.Group {
			return toProtoGroupFromRepo(v)
		}),
		NextPageToken: nextPageToken,
		TotalSize:     int32(total),
	}, nil
}

type CreateGroupRequest struct {
	*apipb.CreateGroupRequest
	UserID idgen.ID `json:"user_id"`
}

func (uc *Groups) randomLogo() string {
	return icons.RandIcon()
}

func (uc *Groups) CreateWithUserID(ctx context.Context, req *CreateGroupRequest) (*apipb.Group, error) {
	now := time.Now()
	groupID := uuid.NewString()

	// Use provided icon or generate random one
	icon := req.Icon
	if icon == "" {
		icon = uc.randomLogo()
	}

	group := &repo.Group{
		ID:          groupID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Icon:        icon,
		Description: req.Description,
		PlanID:      "", // No plan field in CreateGroupRequest
		Status:      repo.StatusActive.Int(),
		OwnerID:     req.UserID,
		CreatedBy:   req.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
		UpdatedBy:   req.UserID,
	}

	if err := uc.groupsRepo.Create(ctx, group); err != nil {
		log.ErrorContext(ctx, "failed to create group", err, "group_name", req.Name, "owner_id", req.UserID)
		return nil, errors.Wrap(err, "failed to create group")
	}

	log.InfoContext(ctx, "group created", "group_id", groupID, "group_name", req.Name, "owner_id", req.UserID)

	var err error
	group, err = uc.groupsRepo.Get(ctx, idgen.ID(group.ID))
	if err != nil {
		return nil, err
	}
	return toProtoGroup(group), nil
}

func toProtoGroup(group *repo.Group) *apipb.Group {
	return &apipb.Group{
		Id:          group.ID,
		Name:        group.Name,
		DisplayName: group.DisplayName,
		Icon:        group.Icon,
		Description: group.Description,
		Plan:        group.PlanID,
		CreatedBy:   group.CreatedBy.String(),
		CreatedAt:   timestamppb.New(group.CreatedAt),
		UpdatedBy:   group.UpdatedBy.String(),
		UpdatedAt:   timestamppb.New(group.UpdatedAt),
	}
}

func toProtoGroupFromRepo(group repo.Group) *apipb.Group {
	return &apipb.Group{
		Id:          group.ID,
		Name:        group.Name,
		DisplayName: group.DisplayName,
		Icon:        group.Icon,
		Description: group.Description,
		Plan:        group.PlanID,
		CreatedBy:   group.CreatedBy.String(),
		CreatedAt:   timestamppb.New(group.CreatedAt),
		UpdatedBy:   group.UpdatedBy.String(),
		UpdatedAt:   timestamppb.New(group.UpdatedAt),
	}
}

func (uc *Groups) Get(ctx context.Context, req *apipb.GetGroupRequest) (*apipb.Group, error) {
	currentUser, _ := contexts.UserID(ctx)
	if !currentUser.IsValid() {
		return nil, apierrors.New(401, "You need to be logged in")
	}

	// Check if user is a member of the group
	_, err := uc.groupsRepo.GetMemberRole(ctx, req.GroupId, currentUser)
	if err != nil {
		return nil, apierrors.New(403, "You are not a member of this group")
	}

	v, err := uc.groupsRepo.Get(ctx, idgen.ID(req.GroupId))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, apierrors.New(404, fmt.Sprintf("group '%s' not found", req.GroupId))
	}
	return toProtoGroup(v), nil
}

func (uc *Groups) Create(ctx context.Context, req *apipb.CreateGroupRequest) (*apipb.Group, error) {
	userId, _ := contexts.UserID(ctx)
	if !userId.IsValid() {
		return nil, apierrors.New(401, "You need to be logged in")
	}
	return uc.CreateWithUserID(ctx, &CreateGroupRequest{
		UserID:             userId,
		CreateGroupRequest: req,
	})
}

// Update updates a group. Only owner and maintainer roles can perform this action.
func (uc *Groups) Update(ctx context.Context, req *apipb.UpdateGroupRequest) (*apipb.Group, error) {
	currentUser, _ := contexts.UserID(ctx)
	if !currentUser.IsValid() {
		return nil, apierrors.New(401, "You need to be logged in")
	}

	// Check if group exists
	group, err := uc.groupsRepo.Get(ctx, idgen.ID(req.GroupId))
	if err != nil {
		return nil, apierrors.New(404, "Group not found")
	}

	// Check if user has permission (owner or maintainer)
	role, err := uc.groupsRepo.GetMemberRole(ctx, req.GroupId, currentUser)
	if err != nil {
		return nil, apierrors.New(403, "You don't have permission to update this group")
	}

	// Only owner and maintainer can update groups
	if role != consts.RoleOwner && role != consts.RoleMaintainer {
		return nil, apierrors.New(403, "Only owner and maintainer can update groups")
	}

	// Track which columns to update
	updateColumns := []string{"updated_at", "updated_by"}

	// Update fields based on optional pointers
	now := time.Now()
	group.UpdatedAt = now
	group.UpdatedBy = currentUser

	if req.DisplayName != nil {
		group.DisplayName = *req.DisplayName
		updateColumns = append(updateColumns, "display_name")
	}
	if req.Icon != nil {
		group.Icon = *req.Icon
		updateColumns = append(updateColumns, "icon")
	}
	if req.Description != nil {
		group.Description = *req.Description
		updateColumns = append(updateColumns, "description")
	}

	if err := uc.groupsRepo.Update(ctx, group, updateColumns...); err != nil {
		log.ErrorContext(ctx, "failed to update group", err, "group_id", req.GroupId, "user_id", currentUser)
		return nil, errors.Wrap(err, "failed to update group")
	}

	log.InfoContext(ctx, "group updated", "group_id", req.GroupId, "user_id", currentUser, "columns", updateColumns)

	v, err := uc.groupsRepo.Get(ctx, idgen.ID(req.GroupId))
	if err != nil {
		return nil, apierrors.New(404, "Group not found")
	}
	return toProtoGroup(v), nil
}

// Delete deletes a group. Only owner and maintainer roles can perform this action.
func (uc *Groups) Delete(ctx context.Context, req *apipb.DeleteGroupRequest) (*apipb.DeleteGroupResponse, error) {
	currentUser, _ := contexts.UserID(ctx)
	if !currentUser.IsValid() {
		return nil, apierrors.New(401, "You need to be logged in")
	}

	// Check if user has permission (owner or maintainer)
	role, err := uc.groupsRepo.GetMemberRole(ctx, req.GroupId, currentUser)
	if err != nil {
		return nil, apierrors.New(404, "Group not found or you are not a member")
	}

	// Only owner and maintainer can delete groups
	if role != consts.RoleOwner && role != consts.RoleMaintainer {
		return nil, apierrors.New(403, "Only owner and maintainer can delete groups")
	}

	// Get group for delete
	group, err := uc.groupsRepo.Get(ctx, idgen.ID(req.GroupId))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get group")
	}

	if err := uc.groupsRepo.Delete(ctx, group); err != nil {
		log.ErrorContext(ctx, "failed to delete group", err, "group_id", req.GroupId, "user_id", currentUser)
		return nil, errors.Wrap(err, "failed to delete group")
	}

	log.InfoContext(ctx, "group deleted", "group_id", req.GroupId, "user_id", currentUser)

	return &apipb.DeleteGroupResponse{}, nil
}

// AddUserToAllGroup adds a user to the "all" system group as viewer
func (uc *Groups) AddUserToAllGroup(ctx context.Context, userID idgen.ID) error {
	// Get the "all" group
	allGroup, err := uc.groupsRepo.GetByName(ctx, "all")
	if err != nil {
		return errors.Wrap(err, "failed to get 'all' group")
	}
	if allGroup == nil {
		return errors.New("'all' group not found")
	}

	// Add user as member with viewer role
	memberAdd := repo.MemberAdd{
		GroupID:   allGroup.ID,
		UserID:    userID,
		Role:      consts.RoleViewer,
		CreatedAt: time.Now(),
		CreatedBy: userID,
	}

	return uc.groupsRepo.AddMember(ctx, memberAdd)
}
