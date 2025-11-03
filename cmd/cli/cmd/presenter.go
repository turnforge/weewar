package cmd

import (
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
	"github.com/panyam/turnengine/games/weewar/services/fsbe"
	"github.com/panyam/turnengine/games/weewar/services/singleton"
)

// PresenterContext holds the presenter and associated panels for CLI operations
type PresenterContext struct {
	Presenter          *services.GameViewPresenter
	GameState          *services.BaseGameState
	TurnOptions        *services.BaseTurnOptionsPanel
	BuildOptions       *services.BaseBuildOptionsModal
	UnitStats          *services.BaseUnitPanel
	TerrainStats       *services.BaseTilePanel
	DamageDistribution *services.BaseUnitPanel
	GameScene          *services.BaseGameScene
	GameID             string
	FSService          *fsbe.FSGamesService
}

// createPresenter loads a game from disk into an in-memory presenter
func createPresenter(gameID string) (*PresenterContext, error) {
	ctx := context.Background()

	// Load game from disk using FSGamesService
	fsService := fsbe.NewFSGamesService("")
	gameResp, err := fsService.GetGame(ctx, &v1.GetGameRequest{Id: gameID})
	if err != nil {
		return nil, fmt.Errorf("failed to load game %s: %w", gameID, err)
	}

	// Always use SingletonGamesService for in-memory operations
	singletonService := singleton.NewSingletonGamesService()
	singletonService.SingletonGame = gameResp.Game
	singletonService.SingletonGameState = gameResp.State
	singletonService.SingletonGameMoveHistory = gameResp.History

	// Create presenter (already initializes RulesEngine and Theme)
	presenter := services.NewGameViewPresenter()
	presenter.GamesService = singletonService

	// Create base panels (data-only, no HTML rendering)
	gameState := &services.BaseGameState{
		Game:  gameResp.Game,
		State: gameResp.State,
	}
	turnOptions := &services.BaseTurnOptionsPanel{}
	buildOptions := &services.BaseBuildOptionsModal{}
	unitStats := &services.BaseUnitPanel{}
	terrainStats := &services.BaseTilePanel{}
	damageDistribution := &services.BaseUnitPanel{}
	gameScene := &services.BaseGameScene{}

	// Wire up presenter with base panels
	presenter.GameState = gameState
	presenter.TurnOptionsPanel = turnOptions
	presenter.BuildOptionsModal = buildOptions
	presenter.UnitStatsPanel = unitStats
	presenter.TerrainStatsPanel = terrainStats
	presenter.DamageDistributionPanel = damageDistribution
	presenter.GameScene = gameScene

	return &PresenterContext{
		Presenter:          presenter,
		GameState:          gameState,
		TurnOptions:        turnOptions,
		BuildOptions:       buildOptions,
		UnitStats:          unitStats,
		TerrainStats:       terrainStats,
		DamageDistribution: damageDistribution,
		GameScene:          gameScene,
		GameID:             gameID,
		FSService:          fsService,
	}, nil
}

// savePresenterState persists the current presenter state back to disk
func savePresenterState(pc *PresenterContext, dryrun bool) error {
	if dryrun {
		if isVerbose() {
			fmt.Println("[VERBOSE] Dryrun mode: skipping save to disk")
		}
		return nil
	}

	ctx := context.Background()
	pc, game, gameState, gameHistory, _, err := GetGame()
	if err != nil {
		panic(err)
	}

	// Save game state back to disk using FSGamesService
	_, err = pc.FSService.UpdateGame(ctx, &v1.UpdateGameRequest{
		GameId:     pc.GameID,
		NewGame:    game,
		NewState:   gameState,
		NewHistory: gameHistory,
	})

	if err != nil {
		return fmt.Errorf("failed to save game state: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Game state saved to disk for game %s\n", pc.GameID)
	}

	return nil
}
