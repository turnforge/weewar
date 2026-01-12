//go:build !wasm
// +build !wasm

package r2

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// R2FileStoreService implements a Cloudflare R2-backed file storage service
type R2FileStoreService struct {
	Client *R2Client
}

// NewR2FileStoreService creates a new R2FileStoreService
func NewR2FileStoreService(client *R2Client) *R2FileStoreService {
	return &R2FileStoreService{
		Client: client,
	}
}

// validatePath ensures the path is safe (no directory traversal, no absolute paths)
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
		return fmt.Errorf("absolute paths are not allowed: %s", path)
	}

	// Clean and check for directory traversal
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") {
		return fmt.Errorf("path cannot escape root: %s", path)
	}

	return nil
}

// populateSignedURLs generates signed URLs with standard expiries for a file
func (s *R2FileStoreService) populateSignedURLs(ctx context.Context, file *v1.File) {
	file.SignedUrls = make(map[string]string)

	// 15 minute URL (for quick previews)
	if url, err := s.Client.GetPresignedURL(ctx, file.Path, 15*time.Minute); err == nil {
		file.SignedUrls["15m"] = url
	}

	// 1 hour URL (default)
	if url, err := s.Client.GetPresignedURL(ctx, file.Path, time.Hour); err == nil {
		file.SignedUrls["1h"] = url
	}

	// 24 hour URL (for sharing)
	if url, err := s.Client.GetPresignedURL(ctx, file.Path, 24*time.Hour); err == nil {
		file.SignedUrls["24h"] = url
	}
}

// PutFile uploads a file to R2
func (s *R2FileStoreService) PutFile(ctx context.Context, req *v1.PutFileRequest) (*v1.PutFileResponse, error) {
	if req.File == nil {
		return nil, fmt.Errorf("file is required")
	}
	if err := validatePath(req.File.Path); err != nil {
		return nil, err
	}

	contentType := req.File.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.Client.Upload(ctx, req.File.Path, req.Content, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	now := tspb.Now()
	downloadURL := s.Client.GetPublicURL(req.File.Path)
	if downloadURL == "" {
		// Fall back to presigned URL if no public URL configured
		downloadURL, _ = s.Client.GetPresignedURLDefault(ctx, req.File.Path)
	}

	file := &v1.File{
		Path:        req.File.Path,
		ContentType: contentType,
		FileSize:    uint64(len(req.Content)),
		IsPublic:    req.File.IsPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
		DownloadUrl: downloadURL,
	}

	return &v1.PutFileResponse{File: file}, nil
}

// GetFile returns metadata about a file in R2
func (s *R2FileStoreService) GetFile(ctx context.Context, req *v1.GetFileRequest) (*v1.GetFileResponse, error) {
	if err := validatePath(req.Path); err != nil {
		return nil, err
	}

	// Get object metadata via HeadObject
	head, err := s.Client.HeadObject(ctx, req.Path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", req.Path)
	}

	// Get the primary download URL (public or default presigned)
	downloadURL := s.Client.GetPublicURL(req.Path)
	if downloadURL == "" {
		downloadURL, _ = s.Client.GetPresignedURLDefault(ctx, req.Path)
	}

	file := &v1.File{
		Path:        req.Path,
		DownloadUrl: downloadURL,
	}

	// Populate metadata from HeadObject response
	if head.ContentLength != nil {
		file.FileSize = uint64(*head.ContentLength)
	}
	if head.ContentType != nil {
		file.ContentType = *head.ContentType
	}
	if head.LastModified != nil {
		file.UpdatedAt = tspb.New(*head.LastModified)
	}

	if req.IncludeSignedUrls {
		s.populateSignedURLs(ctx, file)
	}

	return &v1.GetFileResponse{File: file}, nil
}

// DeleteFile removes a file from R2
func (s *R2FileStoreService) DeleteFile(ctx context.Context, req *v1.DeleteFileRequest) (*v1.DeleteFileResponse, error) {
	if err := validatePath(req.Path); err != nil {
		return nil, err
	}

	exists, err := s.Client.Exists(ctx, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("file not found: %s", req.Path)
	}

	if err := s.Client.Delete(ctx, req.Path); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return &v1.DeleteFileResponse{
		File: &v1.File{Path: req.Path},
	}, nil
}

// ListFiles lists files in R2 with a given prefix
func (s *R2FileStoreService) ListFiles(ctx context.Context, req *v1.ListFilesRequest) (*v1.ListFilesResponse, error) {
	prefix := req.Path
	if prefix != "" {
		if err := validatePath(prefix); err != nil {
			return nil, err
		}
		// Ensure prefix ends with / for directory-like listing
		if !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
	}

	// Default max results
	maxKeys := int32(100)
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		maxKeys = req.Pagination.PageSize
	}

	objects, err := s.Client.ListObjects(ctx, prefix, maxKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	items := make([]*v1.File, 0, len(objects))
	for _, obj := range objects {
		downloadURL := s.Client.GetPublicURL(obj.Key)
		if downloadURL == "" {
			downloadURL, _ = s.Client.GetPresignedURLDefault(ctx, obj.Key)
		}

		file := &v1.File{
			Path:        obj.Key,
			FileSize:    uint64(obj.Size),
			DownloadUrl: downloadURL,
		}

		if obj.LastModified != nil {
			file.UpdatedAt = tspb.New(*obj.LastModified)
		}

		if req.IncludeSignedUrls {
			s.populateSignedURLs(ctx, file)
		}

		items = append(items, file)
	}

	return &v1.ListFilesResponse{
		Items: items,
		Pagination: &v1.PaginationResponse{
			TotalResults: int32(len(items)),
			HasMore:      len(items) >= int(maxKeys),
		},
	}, nil
}
