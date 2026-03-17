package bunrepo

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/internal/domain/groups/repo"
	"github.com/lynx-go/lynx-clean-template/internal/domain/shared/consts"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/uptrace/bun"
)

type GroupMembersRepo struct {
	db *bun.DB
}

func NewGroupMembersRepo(data *clients.DataClients) repo.GroupMembersRepo {
	return &GroupMembersRepo{db: data.GetBunDB()}
}

func (r *GroupMembersRepo) Get(ctx context.Context, id idgen.ID) (*repo.GroupMember, error) {
	var member model.GroupMember
	err := r.db.NewSelect().
		Model(&member).
		Where("id = ?", id.String()).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainGroupMember(&member), nil
}

func (r *GroupMembersRepo) Create(ctx context.Context, v *repo.GroupMember) error {
	_, err := r.db.NewInsert().Model(&model.GroupMember{
		ID:        v.ID.String(),
		GroupID:   v.GroupID,
		UserID:    v.UserID.String(),
		Role:      v.Role.String(),
		Status:    v.Status,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		CreatedBy: v.CreatedBy.String(),
		UpdatedBy: v.UpdatedBy.String(),
	}).Exec(ctx)
	return err
}

func (r *GroupMembersRepo) Update(ctx context.Context, v *repo.GroupMember, columns ...string) error {
	_, err := r.db.NewUpdate().
		Model(&model.GroupMember{
			ID:        v.ID.String(),
			GroupID:   v.GroupID,
			UserID:    v.UserID.String(),
			Role:      v.Role.String(),
			Status:    v.Status,
			UpdatedAt: v.UpdatedAt,
			UpdatedBy: v.UpdatedBy.String(),
		}).
		Column(columns...).
		Where("id = ?", v.ID.String()).
		Exec(ctx)
	return err
}

func (r *GroupMembersRepo) Delete(ctx context.Context, v *repo.GroupMember) error {
	_, err := r.db.NewDelete().
		Model((*model.GroupMember)(nil)).
		Where("id = ?", v.ID.String()).
		Exec(ctx)
	return err
}

func (r *GroupMembersRepo) List(ctx context.Context, params crud.ListParams) ([]*repo.GroupMember, int, string, error) {
	var members []model.GroupMember
	query := r.db.NewSelect().Model(&members)

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

	err = query.Scan(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	results := make([]*repo.GroupMember, len(members))
	for i, m := range members {
		results[i] = toDomainGroupMember(&m)
	}

	return results, total, "", nil
}

func (r *GroupMembersRepo) BatchGet(ctx context.Context, ids []idgen.ID) ([]*repo.GroupMember, error) {
	var members []model.GroupMember
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = id.String()
	}
	err := r.db.NewSelect().
		Model(&members).
		Where("id IN (?)", bun.In(strIds)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*repo.GroupMember, len(members))
	for i, m := range members {
		results[i] = toDomainGroupMember(&m)
	}
	return results, nil
}

func (r *GroupMembersRepo) BatchDelete(ctx context.Context, ids []idgen.ID) error {
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = id.String()
	}
	_, err := r.db.NewDelete().
		Model((*model.GroupMember)(nil)).
		Where("id IN (?)", bun.In(strIds)).
		Exec(ctx)
	return err
}

func toDomainGroupMember(m *model.GroupMember) *repo.GroupMember {
	return &repo.GroupMember{
		ID:        idgen.ID(m.ID),
		GroupID:   m.GroupID,
		UserID:    idgen.ID(m.UserID),
		Role:      consts.Role(m.Role),
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		CreatedBy: idgen.ID(m.CreatedBy),
		UpdatedAt: m.UpdatedAt,
		UpdatedBy: idgen.ID(m.UpdatedBy),
	}
}
