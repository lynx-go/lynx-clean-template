package repo

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type RefreshTokensRepo interface {
	Get(ctx context.Context, id string) (*RefreshToken, error)
	Create(ctx context.Context, val RefreshTokenCreate) (string, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*RefreshToken, error)
	Update(ctx context.Context, update RefreshTokenUpdate) error
	ListByUser(ctx context.Context, userID idgen.ID) ([]*RefreshToken, error)
}
type RefreshTokenCreate struct {
	ID        string    `json:"id"`
	UserID    idgen.ID  `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    idgen.ID  `json:"user_id"`
	Token     string    `json:"token"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefreshTokenUpdate struct {
	ID        string    `json:"id"`
	Revoked   *bool     `json:"revoked"`
	UpdatedAt time.Time `json:"updated_at"`
}
