//go:build !wasm
// +build !wasm

package fsbe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/services"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var FILES_STORAGE_DIR = ""

// FileStoreService implements a local filesystem-based file storage service
type FileStoreService struct {
	BasePath  string
	ClientMgr *services.ClientMgr
}

// NewFileStoreService creates a new FileStoreService implementation for server mode
func NewFileStoreService(storageDir string, clientMgr *services.ClientMgr) *FileStoreService {
	if storageDir == "" {
		if FILES_STORAGE_DIR == "" {
			FILES_STORAGE_DIR = DevDataPath("storage/files")
		}
		storageDir = FILES_STORAGE_DIR
	}
	service := &FileStoreService{
		ClientMgr: clientMgr,
		BasePath:  storageDir,
	}

	return service
}

// resolvePath safely resolves a relative path to an absolute path within BasePath.
// Returns an error if the path attempts to escape the base directory.
func (s *FileStoreService) resolvePath(relativePath string) (string, error) {
	if relativePath == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	return s.resolvePathOrRoot(relativePath)
}

// resolvePathOrRoot is like resolvePath but allows empty paths (returns BasePath).
func (s *FileStoreService) resolvePathOrRoot(relativePath string) (string, error) {
	// Handle empty or "." as the root
	if relativePath == "" || relativePath == "." {
		return s.BasePath, nil
	}

	// Reject absolute paths
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", relativePath)
	}

	// Clean the path to resolve . and .. components
	cleaned := filepath.Clean(relativePath)

	// Reject paths that start with .. after cleaning
	if strings.HasPrefix(cleaned, "..") {
		return "", fmt.Errorf("path escapes base directory: %s", relativePath)
	}

	// Join with base path
	fullPath := filepath.Join(s.BasePath, cleaned)

	// Double-check: ensure the resolved path is still within BasePath
	// This catches edge cases the above checks might miss
	absBase, err := filepath.Abs(s.BasePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Ensure the absolute path starts with the base path
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) && absPath != absBase {
		return "", fmt.Errorf("path escapes base directory: %s", relativePath)
	}

	return fullPath, nil
}

// PutFile stores a file at the specified path
func (s *FileStoreService) PutFile(ctx context.Context, req *v1.PutFileRequest) (resp *v1.PutFileResponse, err error) {
	if req.File == nil {
		return nil, fmt.Errorf("file is required")
	}
	if req.File.Path == "" {
		return nil, fmt.Errorf("file path is required")
	}

	fullPath, err := s.resolvePath(req.File.Path)
	if err != nil {
		return nil, err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file content
	if err := os.WriteFile(fullPath, req.Content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info for response
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	now := tspb.Now()
	file := &v1.File{
		Path:        req.File.Path,
		ContentType: req.File.ContentType,
		FileSize:    uint64(info.Size()),
		IsPublic:    req.File.IsPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
		DownloadUrl: fmt.Sprintf("/files/%s", req.File.Path),
	}

	return &v1.PutFileResponse{File: file}, nil
}

// DeleteFile deletes a file at the specified path
func (s *FileStoreService) DeleteFile(ctx context.Context, req *v1.DeleteFileRequest) (resp *v1.DeleteFileResponse, err error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	fullPath, err := s.resolvePath(req.Path)
	if err != nil {
		return nil, err
	}

	// Get file info before deletion for the response
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", req.Path)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Don't allow deleting directories
	if info.IsDir() {
		return nil, fmt.Errorf("cannot delete directory: %s", req.Path)
	}

	file := &v1.File{
		Path:     req.Path,
		FileSize: uint64(info.Size()),
	}

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return &v1.DeleteFileResponse{File: file}, nil
}

// GetFile returns metadata about a file at the specified path
func (s *FileStoreService) GetFile(ctx context.Context, req *v1.GetFileRequest) (resp *v1.GetFileResponse, err error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	fullPath, err := s.resolvePath(req.Path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", req.Path)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", req.Path)
	}

	file := &v1.File{
		Path:        req.Path,
		FileSize:    uint64(info.Size()),
		UpdatedAt:   tspb.New(info.ModTime()),
		DownloadUrl: fmt.Sprintf("/files/%s", req.Path),
	}

	return &v1.GetFileResponse{File: file}, nil
}

// ListFiles lists files in a directory
func (s *FileStoreService) ListFiles(ctx context.Context, req *v1.ListFilesRequest) (resp *v1.ListFilesResponse, err error) {
	resp = &v1.ListFilesResponse{
		Items: []*v1.File{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}

	// Default to root if path is empty
	dirPath := req.Path
	if dirPath == "" {
		dirPath = "."
	}

	fullPath, err := s.resolvePathOrRoot(dirPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %s", req.Path)
		}
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", req.Path)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := entry.Name()
		if req.Path != "" && req.Path != "." {
			relativePath = filepath.Join(req.Path, entry.Name())
		}

		file := &v1.File{
			Path:        relativePath,
			FileSize:    uint64(entryInfo.Size()),
			UpdatedAt:   tspb.New(entryInfo.ModTime()),
			DownloadUrl: fmt.Sprintf("/files/%s", relativePath),
		}
		resp.Items = append(resp.Items, file)
	}

	resp.Pagination.TotalResults = int32(len(resp.Items))
	return resp, nil
}
