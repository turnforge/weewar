package services

import (
	"context"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/web/assets/themes"
)

type BasePanel interface {
	SetTheme(t themes.Theme)
	SetRulesEngine(t *v1.RulesEngine)
}

type GameState interface {
	SetGameState(context.Context, *v1.SetGameStateRequest) (*v1.SetGameStateResponse, error)
	RemoveUnitAt(context.Context, *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error)
	SetUnitAt(context.Context, *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error)
	UpdateGameStatus(context.Context, *v1.UpdateGameStatusRequest) (*v1.UpdateGameStatusResponse, error)
}

type TurnOptionsPanel interface {
	BasePanel
	CurrentOptions() *v1.GetOptionsAtResponse
	CurrentUnit() *v1.Unit
	SetCurrentUnit(context.Context, *v1.Unit, *v1.GetOptionsAtResponse)
}

type UnitStatsPanel interface {
	BasePanel
	CurrentUnit() *v1.Unit
	SetCurrentUnit(context.Context, *v1.Unit)
}

type DamageDistributionPanel interface {
	BasePanel
	CurrentUnit() *v1.Unit
	SetCurrentUnit(context.Context, *v1.Unit)
}

type TerrainStatsPanel interface {
	BasePanel
	CurrentTile() *v1.Tile
	SetCurrentTile(context.Context, *v1.Tile)
}

type BuildOptionsModal interface {
	BasePanel
	Show(context.Context, *v1.Tile, []*v1.BuildUnitAction, int32)
	Hide(context.Context)
}

type GameScene interface {
	BasePanel
	ClearPaths(context.Context)
	ClearHighlights(context.Context, *v1.ClearHighlightsRequest)
	ShowPath(context.Context, *v1.ShowPathRequest)
	ShowHighlights(context.Context, *v1.ShowHighlightsRequest)
	// Animation methods
	MoveUnit(context.Context, *v1.MoveUnitRequest) (*v1.MoveUnitResponse, error)
	ShowAttackEffect(context.Context, *v1.ShowAttackEffectRequest) (*v1.ShowAttackEffectResponse, error)
	ShowHealEffect(context.Context, *v1.ShowHealEffectRequest) (*v1.ShowHealEffectResponse, error)
	ShowCaptureEffect(context.Context, *v1.ShowCaptureEffectRequest) (*v1.ShowCaptureEffectResponse, error)
	SetUnitAt(context.Context, *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error)
	RemoveUnitAt(context.Context, *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error)
}

type GameViewPresenterImpl interface {
	v1.GameViewPresenterServer
}

type BaseGameViewPresenterImpl struct {
	v1.UnimplementedGameViewPresenterServer
}
