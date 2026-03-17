package files

import (
	"context"
	"net/url"
	"strings"

	"github.com/lynx-go/lynx-clean-template/internal/domain/files/repo"
	"github.com/lynx-go/lynx-clean-template/internal/domain/shared"
	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
)

type urlRenderer struct {
	configs map[string]*config.File_Bucket // bucket name -> config
	repo    repo.Files                     // for catalog://id/ lookup
}

// NewURLRenderer creates a new URLRenderer instance.
func NewURLRenderer(config *config.AppConfig, repo repo.Files) shared.FileURLResolver {
	return &urlRenderer{
		configs: config.GetFile().GetBuckets(),
		repo:    repo,
	}
}

func (r *urlRenderer) Render(ctx context.Context, ref string) string {
	return r.RenderWithContext(ctx, ref, "default")
}

func (r *urlRenderer) RenderWithContext(ctx context.Context, ref string, defaultBucket string) string {
	// Handle empty reference
	if ref == "" {
		return ""
	}

	// Parse the reference using url.Parse()
	u, err := url.Parse(ref)
	if err != nil {
		return ref // return original on parse error
	}

	// 1. Check if already HTTP/HTTPS URL - return as-is
	if u.Scheme == "http" || u.Scheme == "https" {
		return ref
	}

	// 2. Check for catalog://id/{file_id} prefix - lookup files
	// For "catalog://id/file_abc123", u.Scheme="catalog", u.Host="id", u.Path="/file_abc123"
	if u.Scheme == "catalog" {
		lookupType := u.Host
		switch lookupType {
		case "id":
			fileId := strings.TrimPrefix(u.Path, "/")
			fileInfo, err := r.repo.Get(ctx, fileId)
			if err != nil || fileInfo == nil {
				return ref // return original on error (graceful degradation)
			}
			return r.renderBucketURL(fileInfo.Bucket, fileInfo.File, ref)
		default:
			return ref // unknown lookup type
		}
	}

	// 3. Check for bucket:// prefix - host is bucket name, path is file path
	// For "bucket://default/uploads/avatars/123.jpg", u.Scheme="bucket", u.Host="default", u.Path="/uploads/avatars/123.jpg"
	if u.Scheme == "bucket" {
		bucket := u.Host
		// Remove leading slash from path
		filePath := strings.TrimPrefix(u.Path, "/")
		return r.renderBucketURL(bucket, filePath, ref)
	}

	// 4. No recognized scheme - use provided default bucket with ref as path
	return r.renderBucketURL(defaultBucket, ref, ref)
}

func (r *urlRenderer) renderBucketURL(bucket, filePath, originalRef string) string {
	c, ok := r.configs[bucket]
	if !ok {
		return originalRef // return original if bucket not found
	}
	if c.ExposedUrl == "" {
		return originalRef // return original if no exposed_url configured
	}
	s, _ := url.JoinPath(c.ExposedUrl, filePath)
	return s
}

func (r *urlRenderer) RenderBatch(ctx context.Context, refs []string) map[string]string {
	result := make(map[string]string)
	var fileIDs []string
	refToFileID := make(map[string]string)

	// First pass: identify catalog://id/ refs and collect file IDs
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		u, err := url.Parse(ref)
		if err != nil {
			continue
		}
		if u.Scheme == "catalog" && u.Host == "id" {
			fileId := strings.TrimPrefix(u.Path, "/")
			fileIDs = append(fileIDs, fileId)
			refToFileID[ref] = fileId
		}
	}

	// Batch fetch file info
	fileInfos, _ := r.repo.BatchGet(ctx, fileIDs)
	fileInfoMap := make(map[string]*repo.FileInfo)
	for _, info := range fileInfos {
		fileInfoMap[info.ID] = info
	}

	// Second pass: render all refs
	for _, ref := range refs {
		result[ref] = r.renderSingleWithCache(ctx, ref, fileInfoMap, "default")
	}

	return result
}

// renderSingleWithCache is similar to RenderWithContext but uses cached file info
func (r *urlRenderer) renderSingleWithCache(ctx context.Context, ref string, fileInfoCache map[string]*repo.FileInfo, defaultBucket string) string {
	if ref == "" {
		return ""
	}

	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		return ref
	}

	if u.Scheme == "catalog" && u.Host == "id" {
		fileId := strings.TrimPrefix(u.Path, "/")
		fileInfo, ok := fileInfoCache[fileId]
		if !ok {
			return ref
		}
		return r.renderBucketURL(fileInfo.Bucket, fileInfo.File, ref)
	}

	if u.Scheme == "bucket" {
		bucket := u.Host
		filePath := strings.TrimPrefix(u.Path, "/")
		return r.renderBucketURL(bucket, filePath, ref)
	}

	return r.renderBucketURL(defaultBucket, ref, ref)
}
