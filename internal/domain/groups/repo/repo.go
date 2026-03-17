package repo

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/domain/shared/consts"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

//goland:noinspection GoNameStartsWithPackageName
type GroupsRepo interface {
	crud.Repository[idgen.ID, *Group]

	AddMember(ctx context.Context, v MemberAdd) error
	RemoveMember(ctx context.Context, groupId string, userId idgen.ID) error
	ChangeOwner(ctx context.Context, groupId string, userId idgen.ID) error
	ListByMemberID(ctx context.Context, userId idgen.ID, params crud.ListParams) ([]Group, int, string, error)
	GetMembers(ctx context.Context, id string) ([]GroupMemberGet, error)
	GetMemberRole(ctx context.Context, groupId string, userId idgen.ID) (consts.Role, error)
	CountByOwnerAndType(ctx context.Context, ownerId idgen.ID, groupType GroupType) (int, error)
	GetByName(ctx context.Context, name string) (*Group, error)
}

type GroupMembersRepo interface {
	crud.Repository[idgen.ID, *GroupMember]
}

type GroupMember struct {
	ID        idgen.ID    `json:"id"`
	GroupID   string      `json:"group_id"`
	UserID    idgen.ID    `json:"user_id"`
	Status    int         `json:"status"`
	Role      consts.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
	CreatedBy idgen.ID    `json:"created_by"`
	UpdatedAt time.Time   `json:"updated_at"`
	UpdatedBy idgen.ID    `json:"updated_by"`
}

type GroupMemberGet struct {
	GroupID string      `json:"group_id"`
	UserID  idgen.ID    `json:"user_id"`
	Role    consts.Role `json:"role"`
}

type MemberAdd struct {
	GroupID   string      `json:"group_id"`
	UserID    idgen.ID    `json:"user_id"`
	Role      consts.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
	CreatedBy idgen.ID    `json:"created_by"`
}

type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	PlanID      string    `json:"plan_id"`
	Status      int       `json:"status"`
	Type        string    `json:"type"`
	OwnerID     idgen.ID  `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   idgen.ID  `json:"created_by"`
	UpdatedBy   idgen.ID  `json:"updated_by"`
}

type GroupMemberCreate struct {
	GroupID   string      `json:"group_id"`
	UserID    idgen.ID    `json:"user_id"`
	Role      consts.Role `json:"role"`
	Status    int         `json:"status"`
	CreatedBy idgen.ID    `json:"created_by"`
	CreatedAt time.Time   `json:"created_at"`
}

type GroupCreate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	PlanID      string    `json:"plan_id"`
	Status      int       `json:"status"`
	Type        string    `json:"type"`
	OwnerID     idgen.ID  `json:"owner_id"`
	CreatedBy   idgen.ID  `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type GroupUpdate struct {
	ID          string
	Name        *string
	DisplayName *string
	Icon        *string
	Description *string
	PlanID      *string
	Type        *string
	UpdatedBy   idgen.ID
	UpdatedAt   time.Time
}

type Status int

func (s Status) Int() int {
	return int(s)
}

const (
	StatusActive  Status = 1
	StatusDeleted Status = 0
)

// GroupType constants
type GroupType string

const (
	GroupTypeSystem   GroupType = "system"
	GroupTypePersonal GroupType = "personal"
	GroupTypeCustom   GroupType = "custom"
)
