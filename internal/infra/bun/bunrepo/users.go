package bunrepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lynx-go/lynx-clean-template/internal/domain/users/repo"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/mapper"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

type UsersRepo struct {
	db *bun.DB
}

func NewUsersRepo(data *clients.DataClients) repo.UsersRepo {
	return &UsersRepo{db: data.GetBunDB()}
}

// Field mappings for filter and order by
var userFieldMappings = map[string]string{
	"username":        "u.username",
	"display_name":    "u.display_name",
	"email":           "u.email",
	"phone":           "u.phone",
	"role":            "u.role",
	"status":          "u.status",
	"gender":          "u.gender",
	"is_super_admin":  "u.is_super_admin",
	"created_at":      "u.created_at",
	"updated_at":      "u.updated_at",
	"last_sign_in_at": "u.last_sign_in_at",
}

// CrudRepository methods

func (r *UsersRepo) Create(ctx context.Context, u *repo.User) error {
	user := &model.User{
		ID:           u.ID.String(),
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		PasswordHash: u.PasswordHash,
		AvatarURL:    u.AvatarURL,
		Phone:        u.Phone,
		Email:        u.Email,
		Status:       u.Status,
		Gender:       u.Gender,
		Role:         u.Role,
		IsSuperAdmin: u.IsSuperAdmin,
		AppMetadata:  mapper.FromAppMetadata(u.AppMetadata),
		UserMetadata: mapper.FromUserMetadata(u.UserMetadata),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.CreatedAt,
	}

	// Set optional time fields if not zero
	if !u.PhoneConfirmedAt.IsZero() {
		user.PhoneConfirmedAt = u.PhoneConfirmedAt
	}
	if !u.EmailConfirmedAt.IsZero() {
		user.EmailConfirmedAt = u.EmailConfirmedAt
	}
	if !u.ConfirmedAt.IsZero() {
		user.ConfirmedAt = u.ConfirmedAt
	}
	if !u.LastSignInAt.IsZero() {
		user.LastSignInAt = u.LastSignInAt
	}
	if !u.BannedUntil.IsZero() {
		user.BannedUntil = u.BannedUntil
	}

	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *UsersRepo) Update(ctx context.Context, u *repo.User, updateColumns ...string) error {
	user := &model.User{
		ID: u.ID.String(),
	}

	// All updatable columns
	allColumns := map[string]func(*model.User, *repo.User){
		"username":           func(m *model.User, u *repo.User) { m.Username = u.Username },
		"display_name":       func(m *model.User, u *repo.User) { m.DisplayName = u.DisplayName },
		"password_hash":      func(m *model.User, u *repo.User) { m.PasswordHash = u.PasswordHash },
		"avatar_url":         func(m *model.User, u *repo.User) { m.AvatarURL = u.AvatarURL },
		"phone":              func(m *model.User, u *repo.User) { m.Phone = u.Phone },
		"email":              func(m *model.User, u *repo.User) { m.Email = u.Email },
		"status":             func(m *model.User, u *repo.User) { m.Status = u.Status },
		"gender":             func(m *model.User, u *repo.User) { m.Gender = u.Gender },
		"role":               func(m *model.User, u *repo.User) { m.Role = u.Role },
		"is_super_admin":     func(m *model.User, u *repo.User) { m.IsSuperAdmin = u.IsSuperAdmin },
		"app_metadata":       func(m *model.User, u *repo.User) { m.AppMetadata = mapper.FromAppMetadata(u.AppMetadata) },
		"user_metadata":      func(m *model.User, u *repo.User) { m.UserMetadata = mapper.FromUserMetadata(u.UserMetadata) },
		"updated_at":         func(m *model.User, u *repo.User) { m.UpdatedAt = u.UpdatedAt },
		"phone_confirmed_at": func(m *model.User, u *repo.User) { m.PhoneConfirmedAt = u.PhoneConfirmedAt },
		"email_confirmed_at": func(m *model.User, u *repo.User) { m.EmailConfirmedAt = u.EmailConfirmedAt },
		"confirmed_at":       func(m *model.User, u *repo.User) { m.ConfirmedAt = u.ConfirmedAt },
		"last_sign_in_at":    func(m *model.User, u *repo.User) { m.LastSignInAt = u.LastSignInAt },
		"banned_until":       func(m *model.User, u *repo.User) { m.BannedUntil = u.BannedUntil },
	}

	// Determine which columns to update
	columns := updateColumns
	if len(updateColumns) == 0 {
		// If no columns specified, update all updatable columns
		columns = []string{
			"username", "display_name", "password_hash", "avatar_url", "phone", "email",
			"status", "gender", "role", "is_super_admin", "app_metadata", "user_metadata", "updated_at",
		}
		// Add optional time columns only if not zero
		if !u.PhoneConfirmedAt.IsZero() {
			columns = append(columns, "phone_confirmed_at")
		}
		if !u.EmailConfirmedAt.IsZero() {
			columns = append(columns, "email_confirmed_at")
		}
		if !u.ConfirmedAt.IsZero() {
			columns = append(columns, "confirmed_at")
		}
		if !u.LastSignInAt.IsZero() {
			columns = append(columns, "last_sign_in_at")
		}
		if !u.BannedUntil.IsZero() {
			columns = append(columns, "banned_until")
		}
	}

	// Set values for specified columns
	for _, col := range columns {
		if setter, ok := allColumns[col]; ok {
			setter(user, u)
		}
	}

	_, err := r.db.NewUpdate().
		Model(user).
		Column(columns...).
		Where("u.id = ?", u.ID.String()).
		Exec(ctx)

	return err
}

func (r *UsersRepo) Delete(ctx context.Context, u *repo.User) error {
	_, err := r.db.NewDelete().
		Model((*model.User)(nil)).
		Where("u.id = ?", u.ID.String()).
		Exec(ctx)

	return err
}

func (r *UsersRepo) Get(ctx context.Context, id idgen.ID) (*repo.User, error) {
	var user model.User
	err := r.db.NewSelect().
		Model(&user).
		Where("u.id = ?", id.String()).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return toDomainUser2(&user), nil // alias; see below
}

func toDomainUser2(u *model.User) *repo.User { return mapper.ToDomainUser(u) }

func (r *UsersRepo) List(ctx context.Context, params crud.ListParams) ([]*repo.User, int, string, error) {
	config := ListQueryConfig[model.User, *repo.User]{
		FieldMappings: userFieldMappings,
		DefaultOrder:  "u.created_at DESC",
		Converter: func(u model.User) *repo.User {
			return mapper.ToDomainUser(&u)
		},
	}

	return ExecuteListQuery(ctx, r.db, params, config, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("u.status != ?", -2)
	})
}

func (r *UsersRepo) BatchGet(ctx context.Context, ids []idgen.ID) ([]*repo.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	stringIds := idgen.IDsToStrings(ids)
	var users []model.User
	err := r.db.NewSelect().
		Model(&users).
		Where("u.id IN (?)", bun.In(stringIds)).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return lo.Map(users, func(u model.User, _ int) *repo.User {
		return mapper.ToDomainUser(&u)
	}), nil
}

func (r *UsersRepo) BatchDelete(ctx context.Context, ids []idgen.ID) error {
	if len(ids) == 0 {
		return nil
	}

	stringIds := idgen.IDsToStrings(ids)

	_, err := r.db.NewDelete().
		Model((*model.User)(nil)).
		Where("u.id IN (?)", bun.In(stringIds)).
		Exec(ctx)

	return err
}

// UsersRepo-specific methods

// Allowed sort fields for QueryList to prevent SQL injection
var allowedUserSortFields = map[string]bool{
	"username":     true,
	"display_name": true,
	"email":        true,
	"phone":        true,
	"role":         true,
	"status":       true,
	"created_at":   true,
}

func (r *UsersRepo) GetByUsername(ctx context.Context, username string) (*repo.User, error) {
	var user model.User
	err := r.db.NewSelect().
		Model(&user).
		Where("u.username = ?", username).
		Where("u.status != ?", -2).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return mapper.ToDomainUser(&user), nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*repo.User, error) {
	var user model.User
	err := r.db.NewSelect().
		Model(&user).
		Where("u.email = ?", email).
		Where("u.status != ?", -2).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return mapper.ToDomainUser(&user), nil
}

func (r *UsersRepo) IsSuperAdmin(ctx context.Context, userId idgen.ID) (bool, error) {
	var user model.User
	err := r.db.NewSelect().
		Model(&user).
		Column("is_super_admin").
		Where("u.id = ?", userId.String()).
		Scan(ctx)

	if err != nil {
		return false, err
	}

	return user.IsSuperAdmin, nil
}
