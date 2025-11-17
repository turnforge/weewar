package services

import (
	"context"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1s "github.com/turnforge/weewar/gen/go/weewar/v1/services"
)

type IndexerService interface {
	v1s.IndexerServiceServer
}

type BaseIndexerService struct {
	Self IndexerService // The actual implementation
	v1s.UnimplementedIndexerServiceServer
}

func (b *BaseIndexerService) GetIndexStates(contextcontext *v1.GetIndexStatesRequest) (resp *v1.GetIndexStatesResponse, err error) {
	return
}

func (b *BaseIndexerService) DeleteIndexStates(contextcontext *v1.DeleteIndexStatesRequest) (resp *v1.DeleteIndexStatesResponse, err error) {
	return
}

func (b *BaseIndexerService) CreateIndexRecordsLRO(context.Context, *v1.CreateIndexRecordsLRORequest) (resp *v1.CreateIndexRecordsLROResponse, err error) {
	return
}

func (b *BaseIndexerService) GetIndexRecordsLRO(context.Context, *v1.GetIndexRecordsLRORequest) (resp *v1.GetIndexRecordsLROResponse, err error) {
	return
}

func (b *BaseIndexerService) UpdateIndexRecordsLRO(context.Context, *v1.UpdateIndexRecordsLRORequest) (resp *v1.UpdateIndexRecordsLROResponse, err error) {
	return
}
