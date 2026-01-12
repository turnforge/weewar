package services

import (
	"context"
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/web/assets/themes"
)

// Data-Only panel implementations

type PanelBase struct {
	Theme       themes.Theme
	RulesEngine *v1.RulesEngine
}

func (p *PanelBase) SetTheme(t themes.Theme) {
	p.Theme = t
}

func (p *PanelBase) SetRulesEngine(r *v1.RulesEngine) {
	p.RulesEngine = r
}

// BaseGameState is a non-UI implementation of GameState interface
// Used for CLI and testing - stores game state without rendering
type BaseGameState struct {
	Game  *v1.Game
	State *v1.GameState
}

func (b *BaseGameState) SetGameState(_ context.Context, req *v1.SetGameStateRequest) (*v1.SetGameStateResponse, error) {
	b.Game = req.Game
	b.State = req.State
	return nil, nil
}

func (b *BaseGameState) SetUnitAt(_ context.Context, req *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error) {
	if b.State == nil || b.State.WorldData == nil {
		return nil, fmt.Errorf("game state not initialized")
	}

	// Initialize map if needed
	if b.State.WorldData.UnitsMap == nil {
		b.State.WorldData.UnitsMap = make(map[string]*v1.Unit)
	}

	// Set unit at coordinate using map-based storage
	key := lib.CoordKey(req.Q, req.R)
	b.State.WorldData.UnitsMap[key] = req.Unit

	return nil, nil
}

func (b *BaseGameState) RemoveUnitAt(_ context.Context, req *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error) {
	if b.State == nil || b.State.WorldData == nil {
		return nil, fmt.Errorf("game state not initialized")
	}

	// Remove unit at coordinate using map-based storage
	if b.State.WorldData.UnitsMap != nil {
		key := lib.CoordKey(req.Q, req.R)
		delete(b.State.WorldData.UnitsMap, key)
	}

	return nil, nil
}

func (b *BaseGameState) UpdateGameStatus(_ context.Context, req *v1.UpdateGameStatusRequest) (*v1.UpdateGameStatusResponse, error) {
	if b.State == nil {
		return nil, fmt.Errorf("game state not initialized")
	}

	b.State.CurrentPlayer = req.CurrentPlayer
	b.State.TurnCounter = req.TurnCounter

	return nil, nil
}

type BaseUnitPanel struct {
	PanelBase
	Unit *v1.Unit
}

type BaseTilePanel struct {
	PanelBase
	Tile *v1.Tile
}

type BaseGameScene struct {
	PanelBase
	CurrentPathsRequest      *v1.ShowPathRequest
	CurrentHighlightsRequest *v1.ShowHighlightsRequest
}

func (b *BaseGameScene) ClearPaths(context.Context) {
	b.CurrentPathsRequest = nil
}

func (b *BaseGameScene) ClearHighlights(_ context.Context, req *v1.ClearHighlightsRequest) {
	// Only clear CurrentHighlightsRequest if clearing all or clearing specific interactive types
	if req == nil || len(req.Types) == 0 {
		b.CurrentHighlightsRequest = nil
	}
}

func (b *BaseGameScene) ShowPath(_ context.Context, p *v1.ShowPathRequest) {
	b.CurrentPathsRequest = p
}

func (b *BaseGameScene) ShowHighlights(_ context.Context, h *v1.ShowHighlightsRequest) {
	b.CurrentHighlightsRequest = h
}

// Animation methods - no-ops for CLI
func (b *BaseGameScene) MoveUnit(_ context.Context, _ *v1.MoveUnitRequest) (*v1.MoveUnitResponse, error) {
	return &v1.MoveUnitResponse{}, nil
}

func (b *BaseGameScene) ShowAttackEffect(_ context.Context, _ *v1.ShowAttackEffectRequest) (*v1.ShowAttackEffectResponse, error) {
	return &v1.ShowAttackEffectResponse{}, nil
}

func (b *BaseGameScene) ShowHealEffect(_ context.Context, _ *v1.ShowHealEffectRequest) (*v1.ShowHealEffectResponse, error) {
	return &v1.ShowHealEffectResponse{}, nil
}

func (b *BaseGameScene) ShowCaptureEffect(_ context.Context, _ *v1.ShowCaptureEffectRequest) (*v1.ShowCaptureEffectResponse, error) {
	return &v1.ShowCaptureEffectResponse{}, nil
}

func (b *BaseGameScene) SetUnitAt(_ context.Context, _ *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error) {
	return &v1.SetUnitAtResponse{}, nil
}

func (b *BaseGameScene) RemoveUnitAt(_ context.Context, _ *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error) {
	return &v1.RemoveUnitAtResponse{}, nil
}

type BaseTurnOptionsPanel struct {
	BaseUnitPanel
	Options *v1.GetOptionsAtResponse
}

func (b *BaseTurnOptionsPanel) CurrentOptions() *v1.GetOptionsAtResponse {
	return b.Options
}

func (b *BaseTurnOptionsPanel) SetCurrentUnit(_ context.Context, unit *v1.Unit, options *v1.GetOptionsAtResponse) {
	b.Unit = unit
	if options == nil {
		options = &v1.GetOptionsAtResponse{}
	}
	b.Options = options
}

func (b *BaseUnitPanel) CurrentUnit() *v1.Unit {
	return b.Unit
}

func (b *BaseUnitPanel) SetCurrentUnit(_ context.Context, u *v1.Unit) {
	b.Unit = u
}

func (b *BaseTilePanel) CurrentTile() *v1.Tile {
	return b.Tile
}

func (b *BaseTilePanel) SetCurrentTile(_ context.Context, u *v1.Tile) {
	b.Tile = u
}

type BaseBuildOptionsModal struct {
	PanelBase
	BuildOptions []*v1.BuildUnitAction
	Tile         *v1.Tile
	PlayerCoins  int32
}

func (b *BaseBuildOptionsModal) Show(_ context.Context, tile *v1.Tile, buildOptions []*v1.BuildUnitAction, playerCoins int32) {
	b.Tile = tile
	b.BuildOptions = buildOptions
	b.PlayerCoins = playerCoins
}

func (b *BaseBuildOptionsModal) Hide(_ context.Context) {
	b.Tile = nil
	b.BuildOptions = nil
	b.PlayerCoins = 0
}

type BaseCompactSummaryCardPanel struct {
	PanelBase
	Tile *v1.Tile
	Unit *v1.Unit
}

func (b *BaseCompactSummaryCardPanel) SetCurrentData(_ context.Context, tile *v1.Tile, unit *v1.Unit) {
	b.Tile = tile
	b.Unit = unit
}

// PlayerStats holds computed stats for a player (bases, units counts)
type PlayerStats struct {
	Bases int32
	Units int32
}

// BaseGameStatePanel is a non-UI implementation of GameStatePanel
type BaseGameStatePanel struct {
	PanelBase
	Game                *v1.Game
	State               *v1.GameState
	PlayerStats         map[int32]*PlayerStats
	CurrentPlayerCoins  int32
	CurrentPlayerIncome int32
	IncomeBreakdown     string
}

// Update refreshes the panel with current game state
func (b *BaseGameStatePanel) Update(_ context.Context, game *v1.Game, state *v1.GameState) {
	b.Game = game
	b.State = state
	b.ComputePlayerStats()
	b.ComputeCurrentPlayerIncome()
}

// ComputePlayerStats calculates bases and units count per player from world data
func (b *BaseGameStatePanel) ComputePlayerStats() {
	b.PlayerStats = make(map[int32]*PlayerStats)

	if b.State == nil || b.State.WorldData == nil {
		return
	}

	// Count tiles (bases) per player
	for _, tile := range b.State.WorldData.TilesMap {
		if tile.Player > 0 {
			if b.PlayerStats[tile.Player] == nil {
				b.PlayerStats[tile.Player] = &PlayerStats{}
			}
			b.PlayerStats[tile.Player].Bases++
		}
	}

	// Count units per player
	for _, unit := range b.State.WorldData.UnitsMap {
		if unit.Player > 0 {
			if b.PlayerStats[unit.Player] == nil {
				b.PlayerStats[unit.Player] = &PlayerStats{}
			}
			b.PlayerStats[unit.Player].Units++
		}
	}
}

// ComputeCurrentPlayerIncome calculates income for the current player
func (b *BaseGameStatePanel) ComputeCurrentPlayerIncome() {
	b.CurrentPlayerCoins = 0
	b.CurrentPlayerIncome = 0
	b.IncomeBreakdown = ""

	if b.Game == nil || b.Game.Config == nil || b.State == nil {
		return
	}

	currentPlayer := b.State.CurrentPlayer

	// Get current player's coins (from GameState.PlayerStates)
	if playerState := b.State.PlayerStates[currentPlayer]; playerState != nil {
		b.CurrentPlayerCoins = playerState.Coins
	}

	// Calculate income from owned tiles
	if b.State.WorldData == nil {
		return
	}

	incomeConfig := b.Game.Config.IncomeConfigs
	baseCounts := make(map[int32]int32) // tileType -> count

	for _, tile := range b.State.WorldData.TilesMap {
		if tile.Player == currentPlayer {
			tileIncome := lib.GetTileIncomeFromConfig(tile.TileType, incomeConfig)
			if tileIncome > 0 {
				b.CurrentPlayerIncome += tileIncome
				baseCounts[tile.TileType]++
			}
		}
	}

	// Build income breakdown string
	b.IncomeBreakdown = b.buildIncomeBreakdown(baseCounts, incomeConfig)
}

// GetPlayerColorHex returns the primary hex color for a player from the theme
func (b *BaseGameStatePanel) GetPlayerColorHex(playerId int32) string {
	if b.Theme == nil {
		panic("Theme not set on BaseGameStatePanel")
	}
	color := b.Theme.GetPlayerColor(playerId)
	if color != nil && color.Primary != "" {
		return color.Primary
	}
	return "#888888"
}

// buildIncomeBreakdown creates a human-readable income breakdown string
func (b *BaseGameStatePanel) buildIncomeBreakdown(baseCounts map[int32]int32, incomeConfig *v1.IncomeConfig) string {
	if len(baseCounts) == 0 {
		return ""
	}

	parts := []string{}
	for tileType, count := range baseCounts {
		income := lib.GetTileIncomeFromConfig(tileType, incomeConfig)
		if count == 1 {
			parts = append(parts, fmt.Sprintf("%d", income))
		} else {
			parts = append(parts, fmt.Sprintf("%d x %d", count, income))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " + "
		}
		result += part
	}
	return result
}
