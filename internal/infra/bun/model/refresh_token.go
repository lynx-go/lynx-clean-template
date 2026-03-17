package model

import (
	"time"

	"github.com/uptrace/bun"
)

type RefreshToken struct {
	bun.BaseModel `bun:"table:refresh_tokens,alias:rt"`

	ID        string    `bun:"id,pk" json:"id"`
	UserID    string    `bun:"user_id" json:"user_id"`
	Token     string    `bun:"token" json:"token"`
	Revoked   bool      `bun:"revoked" json:"revoked"`
	CreatedAt time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at" json:"updated_at"`
}
