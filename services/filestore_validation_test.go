//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"testing"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"google.golang.org/grpc/metadata"
)

// contextWithUserID creates a context with the user ID set in gRPC metadata.
// This simulates what the auth interceptor does in production.
func contextWithUserID(userID string) context.Context {
	md := metadata.Pairs("x-user-id", userID)
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestValidateContentType(t *testing.T) {
	validator := NewFileStoreValidator(nil)

	tests := []struct {
		name        string
		contentType string
		wantErr     bool
	}{
		{
			name:        "valid PNG",
			contentType: "image/png",
			wantErr:     false,
		},
		{
			name:        "valid SVG",
			contentType: "image/svg+xml",
			wantErr:     false,
		},
		{
			name:        "invalid JPEG",
			contentType: "image/jpeg",
			wantErr:     true,
		},
		{
			name:        "invalid text",
			contentType: "text/plain",
			wantErr:     true,
		},
		{
			name:        "invalid octet-stream",
			contentType: "application/octet-stream",
			wantErr:     true,
		},
		{
			name:        "empty content type",
			contentType: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateContentType(tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContentType(%q) error = %v, wantErr %v", tt.contentType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	validator := NewFileStoreValidator(nil)

	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{
			name:    "small file",
			size:    1024,
			wantErr: false,
		},
		{
			name:    "exactly at limit",
			size:    MaxFileSize,
			wantErr: false,
		},
		{
			name:    "over limit by 1",
			size:    MaxFileSize + 1,
			wantErr: true,
		},
		{
			name:    "empty file",
			size:    0,
			wantErr: false,
		},
		{
			name:    "way over limit",
			size:    MaxFileSize * 2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFileSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFileSize(%d) error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
		})
	}
}

func TestAuthorizePathModification_InternalCalls(t *testing.T) {
	validator := NewFileStoreValidator(nil)

	// Internal calls (empty user ID) should always pass
	ctx := context.Background()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "internal call to game screenshot",
			path:    "screenshots/game/abc123/default.png",
			wantErr: false,
		},
		{
			name:    "internal call to world screenshot",
			path:    "screenshots/world/xyz789/modern.svg",
			wantErr: false,
		},
		{
			name:    "internal call to any path",
			path:    "some/random/path.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.authorizePathModification(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("authorizePathModification(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestAuthorizePathModification_ExternalCalls(t *testing.T) {
	validator := NewFileStoreValidator(nil)

	// External calls (with user ID) need proper path format
	ctx := contextWithUserID("user123")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "invalid path - too short",
			path:    "screenshots/game",
			wantErr: true,
		},
		{
			name:    "invalid path - not screenshots",
			path:    "other/game/abc123/file.png",
			wantErr: true,
		},
		{
			name:    "invalid resource type",
			path:    "screenshots/invalid/abc123/file.png",
			wantErr: true,
		},
		// Note: valid paths will fail because ClientMgr is nil
		// Those are tested in integration tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.authorizePathModification(ctx, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("authorizePathModification(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePutFile_BasicValidation(t *testing.T) {
	validator := NewFileStoreValidator(nil)
	ctx := context.Background() // Internal call

	tests := []struct {
		name    string
		req     *v1.PutFileRequest
		wantErr bool
	}{
		{
			name:    "nil file",
			req:     &v1.PutFileRequest{File: nil},
			wantErr: true,
		},
		{
			name: "empty path",
			req: &v1.PutFileRequest{
				File: &v1.File{Path: "", ContentType: "image/png"},
			},
			wantErr: true,
		},
		{
			name: "invalid content type",
			req: &v1.PutFileRequest{
				File:    &v1.File{Path: "screenshots/game/abc/test.jpg", ContentType: "image/jpeg"},
				Content: make([]byte, 100),
			},
			wantErr: true,
		},
		{
			name: "file too large",
			req: &v1.PutFileRequest{
				File:    &v1.File{Path: "screenshots/game/abc/test.png", ContentType: "image/png"},
				Content: make([]byte, MaxFileSize+1),
			},
			wantErr: true,
		},
		{
			name: "valid request (internal call)",
			req: &v1.PutFileRequest{
				File:    &v1.File{Path: "screenshots/game/abc/test.png", ContentType: "image/png"},
				Content: make([]byte, 1024),
			},
			wantErr: false,
		},
		{
			name: "valid SVG request (internal call)",
			req: &v1.PutFileRequest{
				File:    &v1.File{Path: "screenshots/world/xyz/map.svg", ContentType: "image/svg+xml"},
				Content: []byte("<svg></svg>"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePutFile(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePutFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDeleteFile_BasicValidation(t *testing.T) {
	validator := NewFileStoreValidator(nil)
	ctx := context.Background() // Internal call

	tests := []struct {
		name    string
		req     *v1.DeleteFileRequest
		wantErr bool
	}{
		{
			name:    "empty path",
			req:     &v1.DeleteFileRequest{Path: ""},
			wantErr: true,
		},
		{
			name:    "valid path (internal call)",
			req:     &v1.DeleteFileRequest{Path: "screenshots/game/abc/test.png"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDeleteFile(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeleteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
