package cmd

import (
	"context"
	"fmt"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/services"
	"github.com/turnforge/weewar/services/connectclient"
	"github.com/turnforge/weewar/services/fsbe"
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
	CompactSummaryCard *services.BaseCompactSummaryCardPanel
	GameScene          *services.BaseGameScene
	GameID             string
	GamesService       services.GamesService // The underlying service (Connect or Singleton)
	IsRemote           bool                  // Whether using remote server
}

// createPresenter loads a game and creates a presenter
// If serverURL is provided, connects to remote server via Connect protocol
// Otherwise, uses local FSGamesService for file-based storage
func createPresenter(gameID string) (*PresenterContext, error) {
	ctx := context.Background()
	serverURL := getServerURL()

	var gamesService services.GamesService
	var isRemote bool

	if serverURL != "" {
		// Use Connect client to connect to remote server
		gamesService = connectclient.NewConnectGamesClient(serverURL)
		isRemote = true
		if isVerbose() {
			fmt.Printf("[VERBOSE] Connecting to server: %s\n", serverURL)
		}
	} else {
		// Use FSGamesService for local file-based storage
		gamesService = fsbe.NewFSGamesService("", nil)
		isRemote = false
		if isVerbose() {
			fmt.Println("[VERBOSE] Using local file storage")
		}
	}

	// Load game data
	gameResp, err := gamesService.GetGame(ctx, &v1.GetGameRequest{Id: gameID})
	if err != nil {
		return nil, fmt.Errorf("failed to load game %s: %w", gameID, err)
	}

	// Create presenter (already initializes RulesEngine and Theme)
	presenter := services.NewGameViewPresenter()
	presenter.GamesService = gamesService

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
	compactSummaryCard := &services.BaseCompactSummaryCardPanel{}
	gameScene := &services.BaseGameScene{}

	// Wire up presenter with base panels
	presenter.GameState = gameState
	presenter.TurnOptionsPanel = turnOptions
	presenter.BuildOptionsModal = buildOptions
	presenter.UnitStatsPanel = unitStats
	presenter.TerrainStatsPanel = terrainStats
	presenter.DamageDistributionPanel = damageDistribution
	presenter.CompactSummaryCardPanel = compactSummaryCard
	presenter.GameScene = gameScene

	return &PresenterContext{
		Presenter:          presenter,
		GameState:          gameState,
		TurnOptions:        turnOptions,
		BuildOptions:       buildOptions,
		UnitStats:          unitStats,
		TerrainStats:       terrainStats,
		DamageDistribution: damageDistribution,
		CompactSummaryCard: compactSummaryCard,
		GameScene:          gameScene,
		GameID:             gameID,
		GamesService:       gamesService,
		IsRemote:           isRemote,
	}, nil
}

// savePresenterState persists the current presenter state
// For remote mode, saves via Connect client to server
// For local mode, saves via the games service
func savePresenterState(pc *PresenterContext, dryrun bool) error {
	if dryrun {
		if isVerbose() {
			fmt.Println("[VERBOSE] Dryrun mode: skipping save")
		}
		return nil
	}

	ctx := context.Background()
	pc, game, gameState, gameHistory, _, err := GetGame()
	if err != nil {
		panic(err)
	}

	// Save game state using the games service (works for both local and remote)
	_, err = pc.GamesService.UpdateGame(ctx, &v1.UpdateGameRequest{
		GameId:     pc.GameID,
		NewGame:    game,
		NewState:   gameState,
		NewHistory: gameHistory,
	})

	if err != nil {
		return fmt.Errorf("failed to save game state: %w", err)
	}

	if isVerbose() {
		if pc.IsRemote {
			fmt.Printf("[VERBOSE] Game state saved to server for game %s\n", pc.GameID)
		} else {
			fmt.Printf("[VERBOSE] Game state saved locally for game %s\n", pc.GameID)
		}
	}

	return nil
}
