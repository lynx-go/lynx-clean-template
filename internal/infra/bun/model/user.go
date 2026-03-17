package model

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                 string         `bun:"id,pk" json:"id"`
	Username           string         `bun:"username,unique" json:"username"`
	DisplayName        string         `bun:"display_name" json:"display_name"`
	PasswordHash       string         `bun:"password_hash" json:"-"`
	AvatarURL          string         `bun:"avatar_url" json:"avatar_url"`
	Phone              string         `bun:"phone" json:"phone"`
	PhoneConfirmedAt   time.Time      `bun:"phone_confirmed_at" json:"phone_confirmed_at"`
	Email              string         `bun:"email" json:"email"`
	EmailConfirmedAt   time.Time      `bun:"email_confirmed_at" json:"email_confirmed_at"`
	Status             int8           `bun:"status" json:"status"`
	Gender             int8           `bun:"gender" json:"gender"`
	ConfirmedAt        time.Time      `bun:"confirmed_at" json:"confirmed_at"`
	ConfirmationToken  string         `bun:"confirmation_token" json:"confirmation_token"`
	ConfirmationSentAt time.Time      `bun:"confirmation_sent_at" json:"confirmation_sent_at"`
	Role               string         `bun:"role" json:"role"`
	RecoveryToken      string         `bun:"recovery_token" json:"recovery_token"`
	RecoverySentAt     time.Time      `bun:"recovery_sent_at" json:"recovery_sent_at"`
	AppMetadata        map[string]any `bun:"app_metadata,type:jsonb" json:"app_metadata"`
	UserMetadata       map[string]any `bun:"user_metadata,type:jsonb" json:"user_metadata"`
	LastSignInAt       time.Time      `bun:"last_sign_in_at" json:"last_sign_in_at"`
	IsSuperAdmin       bool           `bun:"is_super_admin" json:"is_super_admin"`
	CreatedBy          string         `bun:"created_by" json:"created_by"`
	UpdatedBy          string         `bun:"updated_by" json:"updated_by"`
	BannedUntil        time.Time      `bun:"banned_until" json:"banned_until"`
	CreatedAt          time.Time      `bun:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `bun:"updated_at" json:"updated_at"`
}
