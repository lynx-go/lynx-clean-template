package bunrepo

import (
	"context"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/domain/files/repo"
	"github.com/lynx-go/lynx-clean-template/internal/infra/bun/model"
	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/uptrace/bun"
)

// Field mappings for files filter and order by
var filesFieldMappings = map[string]string{
	"id":         "f.id",
	"file":       "f.file",
	"bucket":     "f.bucket",
	"category":   "f.category",
	"filetype":   "f.filetype",
	"status":     "f.status",
	"created_at": "f.created_at",
	"updated_at": "f.updated_at",
}

type FilesRepo struct {
	db *bun.DB
}

func NewFilesRepo(data *clients.DataClients) repo.Files {
	return &FilesRepo{db: data.GetBunDB()}
}

func (r *FilesRepo) Create(ctx context.Context, v repo.FileCreate) error {
	file := &model.File{
		ID:        v.ID,
		File:      v.File,
		Bucket:    v.Bucket,
		Category:  v.Category,
		FileType:  v.FileType,
		Status:    v.Status,
		Metadata:  v.Metadata,
		CreatedAt: v.CreatedAt,
		CreatedBy: v.CreatedBy.String(),
		UpdatedAt: v.CreatedAt,
		UpdatedBy: v.CreatedBy.String(),
	}

	_, err := r.db.NewInsert().Model(file).Exec(ctx)
	return err
}

func (r *FilesRepo) Get(ctx context.Context, fileId string) (*repo.FileInfo, error) {
	var file model.File
	err := r.db.NewSelect().
		Model(&file).
		Where("f.id = ?", fileId).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return toDomainFileInfo(&file), nil
}

func (r *FilesRepo) List(ctx context.Context, params crud.ListParams) ([]*repo.FileInfo, int64, string, error) {
	config := ListQueryConfig[model.File, *repo.FileInfo]{
		FieldMappings: filesFieldMappings,
		DefaultOrder:  "f.created_at DESC",
		Converter: func(f model.File) *repo.FileInfo {
			return toDomainFileInfo(&f)
		},
	}

	items, total, nextPageToken, err := ExecuteListQuery(ctx, r.db, params, config, nil)
	if err != nil {
		return nil, 0, "", err
	}

	return items, int64(total), nextPageToken, nil
}

func (r *FilesRepo) Update(ctx context.Context, id string, v repo.FileUpdate) error {
	updates := &model.File{
		UpdatedAt: time.Now(),
		UpdatedBy: v.UpdatedBy.String(),
	}

	cols := []string{"updated_at", "updated_by"}

	if v.File != nil {
		updates.File = *v.File
		cols = append(cols, "file")
	}
	if v.Bucket != nil {
		updates.Bucket = *v.Bucket
		cols = append(cols, "bucket")
	}
	if v.Category != nil {
		updates.Category = *v.Category
		cols = append(cols, "category")
	}
	if v.FileType != nil {
		updates.FileType = *v.FileType
		cols = append(cols, "filetype")
	}
	if v.Status != nil {
		updates.Status = *v.Status
		cols = append(cols, "status")
	}
	if v.Metadata != nil {
		updates.Metadata = *v.Metadata
		cols = append(cols, "metadata")
	}

	_, err := r.db.NewUpdate().
		Model(updates).
		Column(cols...).
		Where("f.id = ?", id).
		Exec(ctx)

	return err
}

func (r *FilesRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.File)(nil)).
		Where("f.id = ?", id).
		Exec(ctx)
	return err
}

func (r *FilesRepo) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.NewDelete().
		Model((*model.File)(nil)).
		Where("f.id IN (?)", bun.In(ids)).
		Exec(ctx)
	return err
}

func (r *FilesRepo) BatchGet(ctx context.Context, ids []string) ([]*repo.FileInfo, error) {
	if len(ids) == 0 {
		return []*repo.FileInfo{}, nil
	}

	var files []*model.File
	err := r.db.NewSelect().
		Model(&files).
		Where("f.id IN (?)", bun.In(ids)).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*repo.FileInfo, len(files))
	for i, f := range files {
		result[i] = toDomainFileInfo(f)
	}

	return result, nil
}

func toDomainFileInfo(f *model.File) *repo.FileInfo {
	return &repo.FileInfo{
		ID:        f.ID,
		File:      f.File,
		Bucket:    f.Bucket,
		Category:  f.Category,
		FileType:  f.FileType,
		Status:    f.Status,
		Metadata:  f.Metadata,
		CreatedAt: f.CreatedAt,
		CreatedBy: idgen.ID(f.CreatedBy),
		UpdatedAt: f.UpdatedAt,
		UpdatedBy: idgen.ID(f.UpdatedBy),
	}
}
