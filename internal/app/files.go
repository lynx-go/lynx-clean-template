package app

import (
	"context"

	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/domain/files"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/contexts"
)

func NewFiles(
	svc *files.Service,
) *Files {
	return &Files{svc: svc}
}

type Files struct {
	svc *files.Service
}

func (uc *Files) GetPresignedURL(ctx context.Context, req *apipb.GetPresignedURLRequest) (*apipb.GetPresignedURLResponse, error) {
	userId, _ := contexts.UserID(ctx)

	res, err := uc.svc.GetPresignedURL(ctx, &files.GetPresignedURLRequest{
		UserID:   userId,
		Category: req.Category,
		MimeType: req.MimeType,
		Filename: req.Filename,
		FilePath: req.FilePath,
		FileID:   req.FileId,
	})
	if err != nil {
		return nil, err
	}
	return &apipb.GetPresignedURLResponse{
		FileId:       res.FileID,
		PresignedUrl: res.PresignedURL,
		FilePath:     res.FilePath,
		ExposedUrl:   res.ExposedURL,
	}, nil
}
