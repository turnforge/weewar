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

	// Our various views here that can be "controlled" by the presenter
	TurnOptionsPanel        *TurnOptionsPanel
	PhaserGameScene         *PhaserGameScene
	GameLogPanel            *GameLogPanel
	UnitStatsPanel          *UnitStatsPanel
	DamageDistributionPanel *DamageDistributionPanel
	TerrainStatsPanel       *TerrainStatsPanel
}

func (s *BaseGameViewPresenterServiceImpl) TileClicked(ctx context.Context, req *v1.TileClickedRequest) (resp *v1.TileClickedResponse, err error) {
	resp = &v1.TileClickedResponse{}
	return resp, err
}

func (s *BaseGameViewPresenterServiceImpl) TurnOptionClicked(ctx context.Context, req *v1.TurnOptionClickedRequest) (resp *v1.TurnOptionClickedResponse, err error) {
	resp = &v1.TurnOptionClickedResponse{}
	return resp, err
}

// Our various view interfaces that the service will call
type GameLogPanel interface {
	Log(message string)
}

type TurnOptionsPanel interface {
	SetTurnOptions()
}

type UnitStatsPanel interface {
	SetUnit(*v1.RulesEngine, *v1.Unit)
}

type TerrainStatsPanel interface {
	SetTerrain(*v1.RulesEngine, *v1.Unit)
}

type DamageDistributionPanel interface {
	SetUnit(*v1.RulesEngine, *v1.Unit)
}

type PhaserGameScene interface {
	SetUnit(*v1.RulesEngine, *v1.Unit)
}
