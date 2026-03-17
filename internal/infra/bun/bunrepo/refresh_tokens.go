package bunrepo

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

type RefreshTokensRepo struct {
	db *bun.DB
}

func NewRefreshTokensRepo(data *clients.DataClients) repo.RefreshTokensRepo {
	return &RefreshTokensRepo{db: data.GetBunDB()}
}

func (r *RefreshTokensRepo) Get(ctx context.Context, id string) (*repo.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.NewSelect().
		Model(&token).
		Where("rt.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return toDomainRefreshToken(&token), nil
}

func (r *RefreshTokensRepo) Create(ctx context.Context, val repo.RefreshTokenCreate) (string, error) {
	token := &model.RefreshToken{
		ID:        val.ID,
		UserID:    val.UserID.String(),
		Token:     val.Token,
		Revoked:   false,
		CreatedAt: val.CreatedAt,
		UpdatedAt: val.CreatedAt,
	}

	_, err := r.db.NewInsert().Model(token).Exec(ctx)
	if err != nil {
		return "", err
	}

	return token.ID, nil
}

func (r *RefreshTokensRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*repo.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.NewSelect().
		Model(&token).
		Where("rt.token = ?", refreshToken).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return toDomainRefreshToken(&token), nil
}

func (r *RefreshTokensRepo) Update(ctx context.Context, update repo.RefreshTokenUpdate) error {
	token := &model.RefreshToken{
		ID:        update.ID,
		UpdatedAt: update.UpdatedAt,
	}

	columns := []string{"updated_at"}

	if update.Revoked != nil {
		token.Revoked = *update.Revoked
		columns = append(columns, "revoked")
	}

	_, err := r.db.NewUpdate().
		Model(token).
		Column(columns...).
		Where("rt.id = ?", update.ID).
		Exec(ctx)

	return err
}

func (r *RefreshTokensRepo) ListByUser(ctx context.Context, userID idgen.ID) ([]*repo.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := r.db.NewSelect().
		Model(&tokens).
		Where("rt.user_id = ?", userID.String()).
		Order("rt.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return lo.Map(tokens, func(t model.RefreshToken, _ int) *repo.RefreshToken {
		return toDomainRefreshToken(&t)
	}), nil
}

// Helper functions

func toDomainRefreshToken(t *model.RefreshToken) *repo.RefreshToken {
	return &repo.RefreshToken{
		ID:        t.ID,
		UserID:    idgen.ID(t.UserID),
		Token:     t.Token,
		Revoked:   t.Revoked,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
