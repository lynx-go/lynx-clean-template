package bunrepo

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/internal/domain/groups/repo"
	"github.com/lynx-go/lynx-clean-template/internal/domain/shared/consts"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen/uuid"
	"github.com/uptrace/bun"
)

type GroupsRepo struct {
	db *bun.DB
}

func NewGroupsRepo(data *clients.DataClients) repo.GroupsRepo {
	return &GroupsRepo{db: data.GetBunDB()}
}

func (r *GroupsRepo) Get(ctx context.Context, id idgen.ID) (*repo.Group, error) {
	var group model.Group
	err := r.db.NewSelect().
		Model(&group).
		Where("id = ?", id.String()).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainGroup(&group), nil
}

func (r *GroupsRepo) Create(ctx context.Context, v *repo.Group) error {
	_, err := r.db.NewInsert().Model(&model.Group{
		ID:          v.ID,
		Name:        v.Name,
		DisplayName: v.DisplayName,
		Icon:        v.Icon,
		Description: v.Description,
		PlanID:      v.PlanID,
		Status:      v.Status,
		Type:        v.Type,
		OwnerID:     v.OwnerID.String(),
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		CreatedBy:   v.CreatedBy.String(),
		UpdatedBy:   v.UpdatedBy.String(),
	}).Exec(ctx)
	return err
}

func (r *GroupsRepo) Update(ctx context.Context, v *repo.Group, columns ...string) error {
	_, err := r.db.NewUpdate().
		Model(&model.Group{
			ID:          v.ID,
			Name:        v.Name,
			DisplayName: v.DisplayName,
			Icon:        v.Icon,
			Description: v.Description,
			PlanID:      v.PlanID,
			Status:      v.Status,
			Type:        v.Type,
			OwnerID:     v.OwnerID.String(),
			UpdatedAt:   v.UpdatedAt,
			UpdatedBy:   v.UpdatedBy.String(),
		}).
		Column(columns...).
		Where("id = ?", v.ID).
		Exec(ctx)
	return err
}

func (r *GroupsRepo) Delete(ctx context.Context, v *repo.Group) error {
	_, err := r.db.NewDelete().
		Model((*model.Group)(nil)).
		Where("id = ?", v.ID).
		Exec(ctx)
	return err
}

func (r *GroupsRepo) List(ctx context.Context, params crud.ListParams) ([]*repo.Group, int, string, error) {
	var groups []model.Group
	query := r.db.NewSelect().Model(&groups)

	// Apply filter if provided
	if params.Filter != "" {
		query = query.Where(params.Filter)
	}

	// Apply order by
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	// Apply pagination
	if params.PageSize > 0 {
		query = query.Limit(int(params.PageSize))
	}
	if params.PageToken != "" {
		// Parse token for offset
		query = query.Offset(0) // Simplified - would need proper token parsing
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	results := make([]*repo.Group, len(groups))
	for i, g := range groups {
		results[i] = toDomainGroup(&g)
	}

	return results, total, "", nil
}

func (r *GroupsRepo) BatchGet(ctx context.Context, ids []idgen.ID) ([]*repo.Group, error) {
	var groups []model.Group
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = id.String()
	}
	err := r.db.NewSelect().
		Model(&groups).
		Where("id IN (?)", bun.In(strIds)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*repo.Group, len(groups))
	for i, g := range groups {
		results[i] = toDomainGroup(&g)
	}
	return results, nil
}

func (r *GroupsRepo) BatchDelete(ctx context.Context, ids []idgen.ID) error {
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = id.String()
	}
	_, err := r.db.NewDelete().
		Model((*model.Group)(nil)).
		Where("id IN (?)", bun.In(strIds)).
		Exec(ctx)
	return err
}

func (r *GroupsRepo) AddMember(ctx context.Context, v repo.MemberAdd) error {
	member := &model.GroupMember{
		ID:        uuid.NewString(),
		GroupID:   v.GroupID,
		UserID:    v.UserID.String(),
		Role:      v.Role.String(),
		Status:    1,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.CreatedAt,
		CreatedBy: v.CreatedBy.String(),
		UpdatedBy: v.CreatedBy.String(),
	}

	_, err := r.db.NewInsert().Model(member).Exec(ctx)
	return err
}

func (r *GroupsRepo) RemoveMember(ctx context.Context, groupId string, userId idgen.ID) error {
	_, err := r.db.NewDelete().
		Model((*model.GroupMember)(nil)).
		Where("group_id = ?", groupId).
		Where("user_id = ?", userId.String()).
		Exec(ctx)
	return err
}

func (r *GroupsRepo) ChangeOwner(ctx context.Context, groupId string, userId idgen.ID) error {
	_, err := r.db.NewUpdate().
		Model((*model.Group)(nil)).
		Set("owner_id = ?", userId.String()).
		Where("id = ?", groupId).
		Exec(ctx)
	return err
}

func (r *GroupsRepo) ListByMemberID(ctx context.Context, userId idgen.ID, params crud.ListParams) ([]repo.Group, int, string, error) {
	var groups []model.Group

	query := r.db.NewSelect().
		Model(&groups).
		Join("JOIN group_members AS gm ON gm.group_id = \"group\".id").
		Where("gm.user_id = ?", userId.String()).
		Where("gm.status = 1").
		Where("\"group\".status = 1")

	// Apply filter if provided
	if params.Filter != "" {
		query = query.Where(params.Filter)
	}

	// Apply order by
	if params.OrderBy != "" {
		query = query.Order(params.OrderBy)
	} else {
		query = query.Order("\"group\".created_at DESC")
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	// Apply pagination
	if params.PageSize > 0 {
		query = query.Limit(int(params.PageSize))
	}
	if params.PageToken != "" {
		// Parse token for offset
		query = query.Offset(0) // Simplified
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	results := make([]repo.Group, len(groups))
	for i, g := range groups {
		results[i] = *toDomainGroup(&g)
	}

	return results, total, "", nil
}

func (r *GroupsRepo) GetMembers(ctx context.Context, id string) ([]repo.GroupMemberGet, error) {
	var members []model.GroupMember
	err := r.db.NewSelect().
		Model(&members).
		Where("group_id = ?", id).
		Where("status = 1").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]repo.GroupMemberGet, len(members))
	for i, m := range members {
		results[i] = repo.GroupMemberGet{
			GroupID: m.GroupID,
			UserID:  idgen.ID(m.UserID),
			Role:    consts.Role(m.Role),
		}
	}
	return results, nil
}

func (r *GroupsRepo) GetMemberRole(ctx context.Context, groupId string, userId idgen.ID) (consts.Role, error) {
	var member model.GroupMember
	err := r.db.NewSelect().
		Model(&member).
		Where("group_id = ?", groupId).
		Where("user_id = ?", userId.String()).
		Where("status = 1").
		Scan(ctx)
	if err != nil {
		return consts.RoleViewer, err
	}
	return consts.Role(member.Role), nil
}

func (r *GroupsRepo) CountByOwnerAndType(ctx context.Context, ownerId idgen.ID, groupType repo.GroupType) (int, error) {
	count, err := r.db.NewSelect().
		Model((*model.Group)(nil)).
		Where("owner_id = ?", ownerId.String()).
		Where("type = ?", string(groupType)).
		Count(ctx)
	return count, err
}

func (r *GroupsRepo) GetByName(ctx context.Context, name string) (*repo.Group, error) {
	var group model.Group
	err := r.db.NewSelect().
		Model(&group).
		Where("name = ?", name).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainGroup(&group), nil
}

func toDomainGroup(g *model.Group) *repo.Group {
	return &repo.Group{
		ID:          g.ID,
		Name:        g.Name,
		DisplayName: g.DisplayName,
		Icon:        g.Icon,
		Description: g.Description,
		PlanID:      g.PlanID,
		Status:      g.Status,
		Type:        g.Type,
		OwnerID:     idgen.ID(g.OwnerID),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		CreatedBy:   idgen.ID(g.CreatedBy),
		UpdatedBy:   idgen.ID(g.UpdatedBy),
	}
}
