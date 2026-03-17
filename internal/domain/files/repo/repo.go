package repo

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
)

type Files interface {
	Create(ctx context.Context, v FileCreate) error
	Get(ctx context.Context, fileId string) (*FileInfo, error)
	List(ctx context.Context, params crud.ListParams) ([]*FileInfo, int64, string, error)
	Update(ctx context.Context, id string, v FileUpdate) error
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
	BatchGet(ctx context.Context, ids []string) ([]*FileInfo, error)
}

type FileCreate struct {
	ID        string         `json:"id"`
	File      string         `json:"file"`
	Bucket    string         `json:"bucket"`
	Category  string         `json:"category"`
	FileType  string         `json:"filetype"`
	Status    int            `json:"status"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy idgen.ID       `json:"created_by"`
}

type FileUpdate struct {
	File      *string         `json:"file,omitempty"`
	Bucket    *string         `json:"bucket,omitempty"`
	Category  *string         `json:"category,omitempty"`
	FileType  *string         `json:"filetype,omitempty"`
	Status    *int            `json:"status,omitempty"`
	Metadata  *map[string]any `json:"metadata,omitempty"`
	UpdatedBy idgen.ID        `json:"updated_by"`
}

type FileInfo struct {
	ID        string         `json:"id"`
	File      string         `json:"file"`
	Bucket    string         `json:"bucket"`
	Category  string         `json:"category"`
	FileType  string         `json:"filetype"`
	Status    int            `json:"status"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy idgen.ID       `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy idgen.ID       `json:"updated_by"`
}
