package model

import (
	"time"

	"github.com/uptrace/bun"
)

type GroupMember struct {
	bun.BaseModel `bun:"table:group_members,alias:group_member"`

	ID        string    `bun:"id,pk" json:"id"`
	GroupID   string    `bun:"group_id" json:"group_id"`
	UserID    string    `bun:"user_id" json:"user_id"`
	Role      string    `bun:"role" json:"role"`
	Status    int       `bun:"status" json:"status"`
	CreatedAt time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at" json:"updated_at"`
	CreatedBy string    `bun:"created_by" json:"created_by"`
	UpdatedBy string    `bun:"updated_by" json:"updated_by"`

	// Relations
	Group *Group `bun:"rel:belongs-to,join:group_id=id" json:"group,omitempty"`
	User  *User  `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
