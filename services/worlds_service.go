package services

import (
	"context"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

type WorldsService interface {
	// *
	// Create a new world
	CreateWorld(context.Context, *v1.CreateWorldRequest) (*v1.CreateWorldResponse, error)
	// *
	// Batch get multiple worlds by ID
	GetWorlds(context.Context, *v1.GetWorldsRequest) (*v1.GetWorldsResponse, error)
	// ListWorlds returns all available worlds
	ListWorlds(context.Context, *v1.ListWorldsRequest) (*v1.ListWorldsResponse, error)
	// GetWorld returns a specific world with metadata
	GetWorld(context.Context, *v1.GetWorldRequest) (*v1.GetWorldResponse, error)
	// *
	// Delete a particular world
	DeleteWorld(context.Context, *v1.DeleteWorldRequest) (*v1.DeleteWorldResponse, error)
	// GetWorld returns a specific world with metadata
	UpdateWorld(context.Context, *v1.UpdateWorldRequest) (*v1.UpdateWorldResponse, error)
}

type BaseWorldsService struct {
	Self WorldsService // The actual implementation
}
