//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"strings"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/services/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// MaxFileSize is the maximum allowed file size (5MB)
	MaxFileSize = 5 * 1024 * 1024
)

// AllowedContentTypes is the list of content types allowed for file uploads
var AllowedContentTypes = map[string]bool{
	"image/png":     true,
	"image/svg+xml": true,
}

// FileStoreValidator provides validation and authorization for file operations
type FileStoreValidator struct {
	ClientMgr *ClientMgr
}

// NewFileStoreValidator creates a new FileStoreValidator
func NewFileStoreValidator(clientMgr *ClientMgr) *FileStoreValidator {
	return &FileStoreValidator{
		ClientMgr: clientMgr,
	}
}

// ValidatePutFile validates a PutFile request including authorization
func (v *FileStoreValidator) ValidatePutFile(ctx context.Context, req *v1.PutFileRequest) error {
	if req.File == nil {
		return status.Error(codes.InvalidArgument, "file is required")
	}
	if req.File.Path == "" {
		return status.Error(codes.InvalidArgument, "file path is required")
	}

	// Check content type
	if err := v.validateContentType(req.File.ContentType); err != nil {
		return err
	}

	// Check file size
	if err := v.validateFileSize(len(req.Content)); err != nil {
		return err
	}

	// Check authorization (only for external calls)
	if err := v.authorizePathModification(ctx, req.File.Path); err != nil {
		return err
	}

	return nil
}

// ValidateDeleteFile validates a DeleteFile request including authorization
func (v *FileStoreValidator) ValidateDeleteFile(ctx context.Context, req *v1.DeleteFileRequest) error {
	if req.Path == "" {
		return status.Error(codes.InvalidArgument, "path is required")
	}

	// Check authorization (only for external calls)
	if err := v.authorizePathModification(ctx, req.Path); err != nil {
		return err
	}

	return nil
}

// validateContentType checks if the content type is allowed
func (v *FileStoreValidator) validateContentType(contentType string) error {
	if contentType == "" {
		return status.Error(codes.InvalidArgument, "content type is required")
	}

	if !AllowedContentTypes[contentType] {
		allowed := make([]string, 0, len(AllowedContentTypes))
		for ct := range AllowedContentTypes {
			allowed = append(allowed, ct)
		}
		return status.Errorf(codes.InvalidArgument,
			"content type %q is not allowed; allowed types: %s",
			contentType, strings.Join(allowed, ", "))
	}

	return nil
}

// validateFileSize checks if the file size is within limits
func (v *FileStoreValidator) validateFileSize(size int) error {
	if size > MaxFileSize {
		return status.Errorf(codes.InvalidArgument,
			"file size %d bytes exceeds maximum allowed size of %d bytes",
			size, MaxFileSize)
	}
	return nil
}

// authorizePathModification checks if the user is authorized to modify files at the given path
func (v *FileStoreValidator) authorizePathModification(ctx context.Context, path string) error {
	// Get user ID from context
	userID := authz.GetUserIDFromContext(ctx)

	// Empty userID means internal service call (e.g., screenshot indexer)
	// These bypass authorization checks
	if userID == "" {
		return nil
	}

	// Parse the path to extract resource type and ID
	// Expected format: screenshots/{kind}/{id}/{filename}
	// e.g., screenshots/game/abc123/default.png
	//       screenshots/world/xyz789/modern.svg
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return status.Errorf(codes.PermissionDenied,
			"invalid path format: expected screenshots/{kind}/{id}/...")
	}

	if parts[0] != "screenshots" {
		return status.Errorf(codes.PermissionDenied,
			"only screenshot paths are allowed via API")
	}

	kind := parts[1]
	resourceID := parts[2]

	switch kind {
	case "game":
		return v.verifyGameOwnership(ctx, userID, resourceID)
	case "world":
		return v.verifyWorldOwnership(ctx, userID, resourceID)
	default:
		return status.Errorf(codes.PermissionDenied,
			"unknown resource type: %s", kind)
	}
}

// verifyGameOwnership checks if the user owns the specified game
func (v *FileStoreValidator) verifyGameOwnership(ctx context.Context, userID, gameID string) error {
	if v.ClientMgr == nil {
		return status.Error(codes.Internal, "client manager not configured")
	}

	client := v.ClientMgr.GetGamesSvcClient()
	if client == nil {
		return status.Error(codes.Internal, "games service client not available")
	}

	resp, err := client.GetGame(ctx, &v1.GetGameRequest{Id: gameID})
	if err != nil {
		return status.Errorf(codes.NotFound, "game not found: %s", gameID)
	}

	if resp.Game == nil {
		return status.Errorf(codes.NotFound, "game not found: %s", gameID)
	}

	if resp.Game.CreatorId != userID {
		return status.Errorf(codes.PermissionDenied,
			"you do not have permission to modify screenshots for this game")
	}

	return nil
}

// verifyWorldOwnership checks if the user owns the specified world
func (v *FileStoreValidator) verifyWorldOwnership(ctx context.Context, userID, worldID string) error {
	if v.ClientMgr == nil {
		return status.Error(codes.Internal, "client manager not configured")
	}

	client := v.ClientMgr.GetWorldsSvcClient()
	if client == nil {
		return status.Error(codes.Internal, "worlds service client not available")
	}

	resp, err := client.GetWorld(ctx, &v1.GetWorldRequest{Id: worldID})
	if err != nil {
		return status.Errorf(codes.NotFound, "world not found: %s", worldID)
	}

	if resp.World == nil {
		return status.Errorf(codes.NotFound, "world not found: %s", worldID)
	}

	if resp.World.CreatorId != userID {
		return status.Errorf(codes.PermissionDenied,
			"you do not have permission to modify screenshots for this world")
	}

	return nil
}
