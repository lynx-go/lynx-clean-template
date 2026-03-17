package model

import (
	"time"

	"github.com/uptrace/bun"
)

type File struct {
	bun.BaseModel `bun:"table:files,alias:f"`

	ID        string         `bun:"id,pk" json:"id"`
	File      string         `bun:"file" json:"file"`
	Bucket    string         `bun:"bucket" json:"bucket"`
	Category  string         `bun:"category" json:"category"`
	FileType  string         `bun:"filetype" json:"filetype"`
	Status    int            `bun:"status" json:"status"`
	Metadata  map[string]any `bun:"metadata,type:jsonb" json:"metadata"`
	CreatedAt time.Time      `bun:"created_at" json:"created_at"`
	CreatedBy string         `bun:"created_by" json:"created_by"`
	UpdatedAt time.Time      `bun:"updated_at" json:"updated_at"`
	UpdatedBy string         `bun:"updated_by" json:"updated_by"`
}
