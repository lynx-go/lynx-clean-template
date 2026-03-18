package repo

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/domain/shared"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type UsersRepo interface {
	crud.Repository[idgen.ID, *User]
	CreateWithFirstUserSuperAdmin(ctx context.Context, u *User) (isFirstUser bool, err error)

	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)

	IsSuperAdmin(ctx context.Context, userId idgen.ID) (bool, error)
	// SetSuperAdmin grants (grant=true) or revokes (grant=false) super admin privilege.
	// When revoking, returns an error if the target is the last super admin.
	SetSuperAdmin(ctx context.Context, userID idgen.ID, grant bool) error
}

type UserQueryFilter struct {
	UserID   idgen.ID `json:"user_id"`
	Search   string   `json:"search"`
	Email    string   `json:"email"`
	Phone    string   `json:"phone"`
	SortBy   string   `json:"sort_by"`
	SortDesc bool     `json:"sort_desc"`
}

type UserUpdate struct {
	ID           idgen.ID      `json:"id"`
	LastSignInAt *time.Time    `json:"last_sign_in_at"`
	DisplayName  *string       `json:"display_name"`
	AvatarURL    *string       `json:"avatar_url"`
	Phone        *string       `json:"phone"`
	Email        *string       `json:"email"`
	Gender       *int8         `json:"gender"`
	Status       *int8         `json:"status"`
	IsSuperAdmin *bool         `json:"is_super_admin"`
	Password     *string       `json:"password"`
	UserMetadata *UserMetadata `json:"user_metadata"`
	UpdatedAt    time.Time     `json:"updated_at"`
	UpdatedBy    idgen.ID      `json:"updated_by"`
}

type UserCreate struct {
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	IsSuperAdmin bool      `json:"is_super_admin"`
	Role         string    `json:"role"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    idgen.ID  `json:"created_by"`
}

type User struct {
	shared.AggregateRoot `json:"-" bun:"-"`
	ID                   idgen.ID      `json:"id"`
	Username             string        `json:"username"`
	DisplayName          string        `json:"display_name"`
	PasswordHash         string        `json:"-"` // 密码不能输出
	AvatarURL            string        `json:"avatar_url"`
	Phone                string        `json:"phone"`
	PhoneConfirmedAt     time.Time     `json:"phone_confirmed_at"`
	Email                string        `json:"email"`
	EmailConfirmedAt     time.Time     `json:"email_confirmed_at"`
	Status               int8          `json:"status"`
	Gender               int8          `json:"gender"`
	ConfirmedAt          time.Time     `json:"confirmed_at"`
	Role                 string        `json:"role"`
	IsSuperAdmin         bool          `json:"is_super_admin"`
	AppMetadata          *AppMetadata  `json:"app_metadata"`
	UserMetadata         *UserMetadata `json:"user_metadata"`
	LastSignInAt         time.Time     `json:"last_sign_in_at"`
	BannedUntil          time.Time     `json:"banned_until"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	CreatedBy            idgen.ID      `json:"created_by"`
	UpdatedBy            idgen.ID      `json:"updated_by"`
}

type UserMetadata struct {
	Bio      string   `json:"bio"`
	Urls     []string `json:"urls"`
	Language string   `json:"language"`
}

type AppMetadata struct{}
