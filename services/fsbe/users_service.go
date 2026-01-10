//go:build !wasm
// +build !wasm

package fsbe

import (
	"context"
	"fmt"
	"time"

	"github.com/turnforge/turnengine/engine/storage"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/services"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var USERS_STORAGE_DIR = ""

// FSUsersService implements the UsersService using filesystem storage
type FSUsersService struct {
	services.BaseUsersService
	storage *storage.FileStorage
}

// NewFSUsersService creates a new FSUsersService implementation
func NewFSUsersService(storageDir string) *FSUsersService {
	if storageDir == "" {
		if USERS_STORAGE_DIR == "" {
			USERS_STORAGE_DIR = DevDataPath("storage/users")
		}
		storageDir = USERS_STORAGE_DIR
	}
	service := &FSUsersService{storage: storage.NewFileStorage(storageDir)}
	service.Self = service
	service.StorageProvider = service
	service.InitializeCache()
	return service
}

// LoadUser implements UserStorageProvider
func (s *FSUsersService) LoadUser(ctx context.Context, id string) (*v1.User, error) {
	return storage.LoadFSArtifact[*v1.User](s.storage, id, "user")
}

// ListAllUsers implements UserStorageProvider
func (s *FSUsersService) ListAllUsers(ctx context.Context) ([]*v1.User, error) {
	return storage.ListFSEntities[*v1.User](s.storage, nil)
}

// SaveUser implements UserStorageProvider
func (s *FSUsersService) SaveUser(ctx context.Context, id string, user *v1.User) error {
	return s.storage.SaveArtifact(id, "user", user)
}

// DeleteFromStorage implements UserStorageProvider
func (s *FSUsersService) DeleteFromStorage(ctx context.Context, id string) error {
	return s.storage.DeleteEntity(id)
}

// UserExists implements UserStorageProvider
func (s *FSUsersService) UserExists(ctx context.Context, id string) bool {
	_, err := s.LoadUser(ctx, id)
	return err == nil
}

// CreateUser creates a new user profile
func (s *FSUsersService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	resp := &v1.CreateUserResponse{}
	if req.User == nil {
		return nil, fmt.Errorf("user data is required")
	}

	userId, err := s.storage.CreateEntity(req.User.Id)
	if err != nil {
		// Check if this is an ID conflict
		if req.User.Id != "" {
			suggestedId := req.User.Id + "-" + shortRandSuffix()
			resp.FieldErrors = map[string]string{
				"id": suggestedId,
			}
			return resp, nil
		}
		return nil, err
	}
	req.User.Id = userId

	now := time.Now()
	req.User.CreatedAt = tspb.New(now)
	req.User.UpdatedAt = tspb.New(now)

	if err := s.SaveUser(ctx, userId, req.User); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	resp.User = req.User
	return resp, nil
}

// UpdateUser updates an existing user profile
func (s *FSUsersService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	if req.User == nil || req.User.Id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Load existing user
	existingUser, err := s.LoadUser(ctx, req.User.Id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if req.User.Name != "" {
		existingUser.Name = req.User.Name
	}
	if req.User.Description != "" {
		existingUser.Description = req.User.Description
	}
	if req.User.Email != "" {
		existingUser.Email = req.User.Email
	}
	if req.User.ImageUrl != "" {
		existingUser.ImageUrl = req.User.ImageUrl
	}
	if req.User.Tags != nil {
		existingUser.Tags = req.User.Tags
	}
	if req.User.Extras != nil {
		existingUser.Extras = req.User.Extras
	}
	existingUser.UpdatedAt = tspb.New(time.Now())

	if err := s.SaveUser(ctx, req.User.Id, existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache
	if s.CacheEnabled {
		s.cacheMu.Lock()
		s.userCache[req.User.Id] = existingUser
		s.cacheMu.Unlock()
	}

	return &v1.UpdateUserResponse{User: existingUser}, nil
}
