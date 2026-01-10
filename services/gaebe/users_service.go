//go:build !wasm
// +build !wasm

package gaebe

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1ds "github.com/turnforge/weewar/gen/datastore"
	v1dal "github.com/turnforge/weewar/gen/datastore/dal"
	"github.com/turnforge/weewar/services"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// UsersService implements the UsersService for Google Datastore
type UsersService struct {
	services.BaseUsersService
	client      *datastore.Client
	namespace   string
	MaxPageSize int
	UserDAL     v1dal.UserDatastoreDAL
}

// NewUsersService creates a new Datastore-backed UsersService
func NewUsersService(client *datastore.Client, namespace string) *UsersService {
	service := &UsersService{
		client:      client,
		namespace:   namespace,
		MaxPageSize: 1000,
	}
	service.UserDAL.Namespace = namespace
	service.Self = service
	service.StorageProvider = service
	service.InitializeCache()

	return service
}

// LoadUser implements UserStorageProvider
func (s *UsersService) LoadUser(ctx context.Context, id string) (*v1.User, error) {
	key := NamespacedKey("User", id, s.namespace)
	userDs, err := s.UserDAL.Get(ctx, s.client, key)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, services.ErrUserNotFound
		}
		return nil, err
	}
	return v1ds.UserDatastoreToUser(userDs, nil, nil)
}

// ListAllUsers implements UserStorageProvider
func (s *UsersService) ListAllUsers(ctx context.Context) ([]*v1.User, error) {
	userDsList, err := s.UserDAL.List(ctx, s.client, s.MaxPageSize, nil)
	if err != nil {
		return nil, err
	}

	users := make([]*v1.User, 0, len(userDsList))
	for _, uds := range userDsList {
		user, err := v1ds.UserDatastoreToUser(uds, nil, nil)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// SaveUser implements UserStorageProvider
func (s *UsersService) SaveUser(ctx context.Context, id string, user *v1.User) error {
	key := NamespacedKey("User", id, s.namespace)
	userDs, err := v1ds.UserToUserDatastore(user, nil, nil)
	if err != nil {
		return err
	}
	return s.UserDAL.Save(ctx, s.client, key, userDs)
}

// DeleteFromStorage implements UserStorageProvider
func (s *UsersService) DeleteFromStorage(ctx context.Context, id string) error {
	key := NamespacedKey("User", id, s.namespace)
	return s.UserDAL.Delete(ctx, s.client, key)
}

// UserExists implements UserStorageProvider
func (s *UsersService) UserExists(ctx context.Context, id string) bool {
	key := NamespacedKey("User", id, s.namespace)
	_, err := s.UserDAL.Get(ctx, s.client, key)
	return err == nil
}

// CreateUser creates a new user profile
func (s *UsersService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	resp := &v1.CreateUserResponse{}
	if req.User == nil {
		return nil, fmt.Errorf("user data is required")
	}

	// Generate ID if not provided
	if req.User.Id == "" {
		req.User.Id = GenShortId()
	}

	// Check if user already exists
	if s.UserExists(ctx, req.User.Id) {
		suggestedId := req.User.Id + "-" + GenShortId()[:4]
		resp.FieldErrors = map[string]string{
			"id": suggestedId,
		}
		return resp, nil
	}

	now := time.Now()
	req.User.CreatedAt = tspb.New(now)
	req.User.UpdatedAt = tspb.New(now)

	if err := s.SaveUser(ctx, req.User.Id, req.User); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	resp.User = req.User
	return resp, nil
}

// UpdateUser updates an existing user profile
func (s *UsersService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
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
