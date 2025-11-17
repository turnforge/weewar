//go:build !wasm
// +build !wasm

package gormbe

import (
	"context"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1services "github.com/turnforge/weewar/gen/go/weewar/v1/services"
	v1gorm "github.com/turnforge/weewar/gen/gorm"
	"github.com/turnforge/weewar/services"
	"gorm.io/gorm"
)

// IndexerService implements the GamesService gRPC interface
type IndexerService struct {
	services.BaseIndexerService
	WorldsService v1services.WorldsServiceServer
	storage       *gorm.DB
	MaxPageSize   int
}

// NewGamesService creates a new GamesService implementation for server mode
func NewIndexerService(db *gorm.DB) *IndexerService {
	db.AutoMigrate(&v1gorm.IndexRecordsLROGORM{})
	db.AutoMigrate(&v1gorm.IndexStateGORM{})

	db.AutoMigrate(&GenId{}) // Move this to its own genid service?

	service := &IndexerService{
		BaseIndexerService: services.BaseIndexerService{},
		storage:            db,
		MaxPageSize:        1000,
	}
	service.Self = service

	return service
}

func (i *IndexerService) GetIndexStates(ctx context.Context, req *v1.GetIndexStatesRequest) (resp *v1.GetIndexStatesResponse, err error) {
	ctx, span := Tracer.Start(ctx, "GetIndexState")
	defer span.End()
	// curr, _ := s.DB.GetIndexState(ctx, req.Id)

	resp = &v1.GetIndexStatesResponse{}
	var out []*v1gorm.IndexStateGORM
	query := i.storage.Where("entity_type = ?", req.EntityType).Where("entity_id = ?", req.EntityId)
	err = query.Find(&out).Error
	if err != nil {
		resp.States = map[string]*v1.IndexState{}
		for _, input := range out {
			output, err := v1gorm.IndexStateFromIndexStateGORM(nil, input, nil)
			if err == nil {
				resp.States[input.IndexType] = output
			}
		}
	}
	return
}

func (i *IndexerService) ListIndexStates(ctx context.Context, req *v1.ListIndexStatesRequest) (resp *v1.ListIndexStatesResponse, err error) {
	ctx, span := Tracer.Start(ctx, "ListIndexStates")
	defer span.End()

	var out []*v1gorm.IndexStateGORM
	query := i.storage.Where("entity_type = ?", req.EntityType)
	if req.UpdatedBefore != nil {
		query = query.Where("indexed_at lte ?", req.UpdatedBefore)
	}
	if req.UpdatedAfter != nil {
		query = query.Where("indexed_at gte ?", req.UpdatedAfter)
	}
	if len(req.IndexTypes) > 0 {
		query = query.Where("index_type in ?", req.IndexTypes)
	}
	if req.OrderBy == "id" {
		query = query.Order("entity_id asc")
	} else {
		query = query.Order("indexed_at asc")
	}
	if req.Count <= 0 {
		req.Count = 1000
	}
	query = query.Limit(req.Count)
	err = query.Find(&out).Error

	resp = &v1.ListIndexStatesResponse{}
	if err != nil {
		return
	}
	for _, input := range out {
		output, err := v1gorm.IndexStateFromIndexStateGORM(nil, input, nil)
		if err == nil {
			resp.Items = append(resp.Items, output)
		}
	}
	return
}

func (i *IndexerService) DeleteIndexStates(ctx context.Context, req *v1.DeleteIndexStatesRequest) (resp *v1.DeleteIndexStatesResponse, err error) {
	ctx, span := Tracer.Start(ctx, "DeleteIndexState")
	defer span.End()
	resp = &v1.DeleteIndexStatesResponse{}
	// s.DB.DeleteIndexState(ctx, req.Id)
	query := i.storage.Where("entity_type = ? and entity_id = ?", req.EntityType, req.EntityId)
	if len(req.IndexTypes) > 0 {
		query = query.Where("index_type in ?", req.IndexTypes)
	}
	err = query.Delete(&v1gorm.IndexStateGORM{}).Error
	return
}

func (b *IndexerService) CreateIndexState(ctx context.Context, req *v1.CreateIndexStateRequest) (resp *v1.CreateIndexStateResponse, err error) {
	ctx, span := Tracer.Start(ctx, "CreateIndexState")
	defer span.End()
	/*
		req.CreatorId = GetAuthedUser(ctx)
		if req.IndexState.Base.CreatorId == "" {
			Logger.InfoContext(ctx, "User is not authenticated to create a topic.")
			return nil, status.Error(codes.PermissionDenied, "User is not authenticated to create a topic.")
		}
	*/
	indexState := req.IndexState

	dbIndexState := IndexStateFromProto(indexState)
	err = s.DB.SaveIndexState(ctx, dbIndexState)
	if err == nil {
		resp = &protos.CreateIndexStateResponse{
			IndexState: IndexStateToProto(dbIndexState),
		}
		// Increatement our indexState
		indexStateCnt.Add(ctx, 1)
	}
	return resp, err
}
