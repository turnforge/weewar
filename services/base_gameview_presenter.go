package services

import (
	"context"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type GameViewPresenterServiceImpl interface {
	v1.GameViewPresenterServiceServer
}

type BaseGameViewPresenterServiceImpl struct {
	v1.UnimplementedGameViewPresenterServiceServer
	GamesService GamesServiceImpl
}

func (s *BaseGameViewPresenterServiceImpl) SceneClicked(ctx context.Context, req *v1.SceneClickedRequest) (resp *v1.SceneClickedResponse, err error) {
	resp = &v1.SceneClickedResponse{}
	return resp, err
}

func (s *BaseGameViewPresenterServiceImpl) TurnOptionClicked(ctx context.Context, req *v1.TurnOptionClickedRequest) (resp *v1.TurnOptionClickedResponse, err error) {
	resp = &v1.TurnOptionClickedResponse{}
	return resp, err
}
