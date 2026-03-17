package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Group struct {
	bun.BaseModel `bun:"table:groups,alias:group"`

	ID          string    `bun:"id,pk" json:"id"`
	Name        string    `bun:"name" json:"name"`
	DisplayName string    `bun:"display_name" json:"display_name"`
	Icon        string    `bun:"icon" json:"icon"`
	Description string    `bun:"description" json:"description"`
	PlanID      string    `bun:"plan_id" json:"plan_id"`
	Status      int       `bun:"status" json:"status"`
	Type        string    `bun:"type" json:"type"`
	OwnerID     string    `bun:"owner_id" json:"owner_id"`
	CreatedAt   time.Time `bun:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at" json:"updated_at"`
	CreatedBy   string    `bun:"created_by" json:"created_by"`
	UpdatedBy   string    `bun:"updated_by" json:"updated_by"`

	// Relations
	Owner   *User         `bun:"rel:belongs-to,join:owner_id=id" json:"owner,omitempty"`
	Members []GroupMember `bun:"rel:has-many,join:id=group_id" json:"members,omitempty"`
}
