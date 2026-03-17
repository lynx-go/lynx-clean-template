package files

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"strings"
	"time"

	awsv2cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx-clean-template/internal/domain/files/repo"
	"github.com/lynx-go/lynx-clean-template/internal/domain/shared"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/samber/lo"

	apierrors "github.com/lynx-go/lynx-clean-template/pkg/errors"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen"
	"github.com/lynx-go/lynx-clean-template/pkg/idgen/uuid"
	"github.com/lynx-go/lynx-clean-template/pkg/timeutil"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

type Service struct {
	buckets               map[string]*blob.Bucket
	closeFns              []func()
	bucketCategoriesIndex []bucketCategories
	filePathTpls          map[string]*template.Template
	configs               map[string]*config.File_Bucket
	repo                  repo.Files
	logger                shared.Logger
}

type bucketCategories struct {
	bucket     string
	categories []string
}

func newS3Bucket(ctx context.Context, c *config.File_Bucket) (*blob.Bucket, error) {
	cfg, err := awsv2cfg.LoadDefaultConfig(ctx,
		awsv2cfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyId, c.AccessKeySecret, "")),
		awsv2cfg.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}
	clientV2 := s3v2.NewFromConfig(cfg, func(o *s3v2.Options) {
		o.BaseEndpoint = lo.ToPtr(c.Endpoint)

	})
	return s3blob.OpenBucketV2(ctx, clientV2, c.BucketName, nil)
}

func newService(repo repo.Files, logger shared.Logger) *Service {
	return &Service{
		buckets:               map[string]*blob.Bucket{},
		closeFns:              make([]func(), 0),
		bucketCategoriesIndex: make([]bucketCategories, 0),
		filePathTpls:          map[string]*template.Template{},
		configs:               map[string]*config.File_Bucket{},
		repo:                  repo,
		logger:                logger,
	}
}

func NewService(
	app lynx.Lynx,
	config *config.AppConfig,
	repo repo.Files,
	logger shared.Logger,
) (*Service, func(), error) {
	bucketCfgs := config.GetFile().GetBuckets()
	if len(bucketCfgs) == 0 {
		//return nil, func() {}, errors.New("no buckets configured")
		svc := newService(repo, logger)
		return svc, svc.Close, nil
	}
	// bucket, err := blob.OpenBucket("s3://mybucket?" +
	//    "endpoint=my.minio.local:8080&" +
	//    "disableSSL=true&" +
	//    "s3ForcePathStyle=true")
	buckets := make(map[string]*blob.Bucket)
	closeFns := make([]func(), 0)
	var bucketCategoriesIndex []bucketCategories
	ctx := app.Context()
	for k, c := range bucketCfgs {

		bucket, err := newS3Bucket(ctx, c)
		if err != nil {
			return nil, nil, err
		}
		closeFns = append(closeFns, func() {
			if err := bucket.Close(); err != nil {
				logger.ErrorContext(ctx, "failed to close bucket", err)
			}
		})
		buckets[k] = bucket
		bucketCategoriesIndex = append(bucketCategoriesIndex, bucketCategories{
			bucket:     k,
			categories: c.GetIncludeCategories(),
		})
	}
	filePathTmpls := make(map[string]*template.Template)
	for _, c := range bucketCfgs {
		for k, t := range c.GetFilePathTemplates() {
			tpl, err := template.New(k).Parse(t)
			if err != nil {
				return nil, nil, err
			}
			filePathTmpls[k] = tpl
		}

	}
	svc := &Service{
		buckets:               buckets,
		closeFns:              closeFns,
		bucketCategoriesIndex: bucketCategoriesIndex,
		filePathTpls:          filePathTmpls,
		configs:               config.File.Buckets,
		repo:                  repo,
		logger:                logger,
	}
	return svc, svc.Close, nil
}

type GetPresignedURLRequest struct {
	UserID   idgen.ID `json:"user_id"`
	Category string   `json:"category"`
	MimeType string   `json:"mimetype"`
	Filename string   `json:"filename"`
	FilePath string   `json:"file_path"`
	FileID   string   `json:"file_id"`
	FileSize int64    `json:"file_size"` // Size in bytes
}

func (svc *Service) Close() {
	for _, fn := range svc.closeFns {
		fn()
	}
}

type GetPresignedURLResult struct {
	FileID       string `json:"file_id"`
	PresignedURL string `json:"presigned_url"`
	FilePath     string `json:"file_path"`
	ExposedURL   string `json:"exposed_url"`
}

func (svc *Service) getBucketName(category string) string {
	v, ok := lo.Find(svc.bucketCategoriesIndex, func(v bucketCategories) bool {
		return lo.Contains(v.categories, category)
	})
	if ok {
		return v.bucket
	}
	return svc.bucketCategoriesIndex[0].bucket
}

func (svc *Service) getExposedURL(bucket, filePath string) string {
	c, ok := svc.configs[bucket]
	if !ok {
		return ""
	}
	s, _ := url.JoinPath(c.ExposedUrl, filePath)
	return s
}

func (svc *Service) GetPresignedURL(ctx context.Context, req *GetPresignedURLRequest) (*GetPresignedURLResult, error) {
	bucketName := svc.getBucketName(req.Category)
	bucket := svc.buckets[bucketName]

	// 对文件路径进行签名
	if req.FilePath != "" {
		signedURL, err := bucket.SignedURL(ctx, req.FilePath, &blob.SignedURLOptions{
			Expiry:                   15 * time.Minute, // 15 minutes for downloads
			Method:                   "GET",
			EnforceAbsentContentType: false,
			BeforeSign:               nil,
		})
		if err != nil {
			return nil, err
		}

		return &GetPresignedURLResult{
			PresignedURL: signedURL,
			FilePath:     req.FilePath,
			ExposedURL:   svc.getExposedURL(bucketName, req.FilePath),
		}, nil
	}
	// 查询 file，并签名
	if req.FileID != "" {
		file, err := svc.repo.Get(ctx, req.FileID)
		if err != nil {
			return nil, err
		}
		if file == nil {
			return nil, errors.New("file not found")
		}
		signedURL, err := bucket.SignedURL(ctx, file.File, &blob.SignedURLOptions{
			Expiry:                   15 * time.Minute, // 15 minutes for downloads
			Method:                   "GET",
			EnforceAbsentContentType: false,
			BeforeSign:               nil,
		})
		if err != nil {
			return nil, err
		}

		return &GetPresignedURLResult{
			FileID:       file.ID,
			PresignedURL: signedURL,
			FilePath:     file.File,
			ExposedURL:   svc.getExposedURL(bucketName, file.File),
		}, nil
	}

	fileType, ok := mimeTypeMapping[req.MimeType]
	if !ok {
		return nil, apierrors.Cause("mime type not supported")
	}

	// Validate file size if provided
	if req.FileSize > 0 {
		if req.FileSize > maxFileSize {
			return nil, apierrors.Cause(fmt.Sprintf("file size exceeds maximum allowed size of %d MB", maxFileSize/megabyte))
		}
	}

	fileId := uuid.NewString()
	now := time.Now()
	var filePath string
	tpl, ok := svc.filePathTpls[req.Category]
	if !ok {
		filePath = path.Join(req.Category, fileId+"."+fileType)
	} else {
		sb := &strings.Builder{}
		if err := tpl.Execute(sb, &fileKeyArgs{
			Filename: req.Filename,
			FileID:   fileId,
			Category: req.Category,
			DateID:   timeutil.DateID(now),
			UserID:   req.UserID.String(),
			FileType: fileType,
		}); err != nil {
			return nil, err
		}
		filePath = sb.String()
	}
	basePath := svc.configs[bucketName].GetBasePath()
	if basePath != "" {
		filePath = path.Join(basePath, filePath)
	}

	signedURL, err := bucket.SignedURL(ctx, filePath, &blob.SignedURLOptions{
		Expiry: 15 * time.Minute, // 15 minutes for uploads
		Method: "PUT",
		//ContentType:              req.MimeType,
		EnforceAbsentContentType: false,
		BeforeSign:               nil,
	})
	if err != nil {
		return nil, err
	}

	if err := svc.repo.Create(ctx, repo.FileCreate{
		ID:       fileId,
		File:     filePath,
		Bucket:   bucketName,
		Category: req.Category,
		FileType: fileType,
		Status:   1,
		Metadata: map[string]any{
			"filename": req.Filename,
		},
		CreatedAt: now,
		CreatedBy: req.UserID,
	}); err != nil {
		return nil, err
	}

	return &GetPresignedURLResult{
		FileID:       fileId,
		PresignedURL: signedURL,
		FilePath:     filePath,
		ExposedURL:   svc.getExposedURL(bucketName, filePath),
	}, nil
}

type fileKeyArgs struct {
	Filename string `json:"filename"`
	FileID   string `json:"file_id"`
	Category string `json:"category"`
	DateID   string `json:"date_id"`
	UserID   string `json:"user_id"`
	FileType string `json:"file_type"`
}

var mimeTypeMapping = map[string]string{
	"image/jpeg":      "jpg",
	"image/gif":       "gif",
	"image/png":       "png",
	"image/webp":      "webp",
	"application/pdf": "pdf",
	"text/plain":      "txt",
	"text/csv":        "csv",
	"application/zip": "zip",
}

const (
	megabyte    = 1024 * 1024
	maxFileSize = 100 * megabyte // Maximum file size: 100MB
)
