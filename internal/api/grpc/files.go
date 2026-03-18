package grpc

import (
	"context"

	apipb "github.com/lynx-go/lynx-clean-template/genproto/api/v1"
	"github.com/lynx-go/lynx-clean-template/internal/app"
)

type FilesService struct {
	apipb.UnimplementedFilesServiceServer
	uc *app.Files
}

func (svc *FilesService) GetPresignedURL(ctx context.Context, request *apipb.GetPresignedURLRequest) (*apipb.GetPresignedURLResponse, error) {
	return svc.uc.GetPresignedURL(ctx, request)
}

func NewFileService(
	uc *app.Files,
) *FilesService {
	return &FilesService{uc: uc}
}

var _ apipb.FilesServiceServer = (*FilesService)(nil)
