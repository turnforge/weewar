package singleton

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
)

// Test that we can load a game and get options for units
func TestSingletonGamesService_GetOptionsAt(t *testing.T) {
	// Load a real test world using the existing test utility
	homeDir, _ := os.UserHomeDir()
	worldsDir := filepath.Join(homeDir, "dev-app-data", "weewar", "storage", "worlds")

	world, gameState, err := services.LoadTestWorldFromStorage(worldsDir, "32112070")
	if err != nil {
		t.Skipf("Skipping test - world data not available: %v", err)
	}

	// Create game
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	// Create service
	gamesService := NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.SingletonGameMoveHistory = &v1.GameMoveHistory{}
	gamesService.Self = gamesService

	// Store the world in gameState's WorldData so GetRuntimeGame can use it
	rtGame, err := gamesService.GetRuntimeGame(game, gameState)
	if err != nil {
		t.Fatalf("GetRuntimeGame failed: %v", err)
	}

	// Use the loaded world
	rtGame.World = world

	// Find a unit in the world
	var testUnit *v1.Unit
	var testCoord services.AxialCoord

	for coord, unit := range world.UnitsByCoord() {
		testUnit = unit
		testCoord = coord
		break
	}

	if testUnit == nil {
		t.Skip("No units found in test world, skipping")
	}

	ctx := context.Background()

	// Get options for this unit
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		Q: int32(testCoord.Q),
		R: int32(testCoord.R),
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Verify we got a response
	if resp == nil {
		t.Fatal("GetOptionsAt returned nil response")
	}

	// We expect at least some options for a unit (move, attack, or capture)
	t.Logf("Unit at (%d,%d) has %d options", testCoord.Q, testCoord.R, len(resp.Options))

	// Verify option types
	for i, opt := range resp.Options {
		if moveOpt := opt.GetMove(); moveOpt != nil {
			t.Logf("  Option %d: Move to (%d,%d)", i, moveOpt.ToQ, moveOpt.ToR)
			if moveOpt.ReconstructedPath != nil {
				t.Logf("    Path has %d edges", len(moveOpt.ReconstructedPath.Edges))
			}
		} else if attackOpt := opt.GetAttack(); attackOpt != nil {
			t.Logf("  Option %d: Attack at (%d,%d)", i, attackOpt.DefenderQ, attackOpt.DefenderR)
		} else if captureOpt := opt.GetCapture(); captureOpt != nil {
			t.Logf("  Option %d: Capture at (%d,%d)", i, captureOpt.Q, captureOpt.R)
		}
	}
}

// Test getting options for empty tile returns empty
func TestSingletonGamesService_GetOptionsAt_EmptyTile(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	worldsDir := filepath.Join(homeDir, "dev-app-data", "weewar", "storage", "worlds")

	_, gameState, err := services.LoadTestWorldFromStorage(worldsDir, "32112070")
	if err != nil {
		t.Skipf("Skipping test - world data not available: %v", err)
	}

	game := &v1.Game{Id: "test-game"}

	gamesService := NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.Self = gamesService

	// Find an empty tile (try coordinates unlikely to have units)
	emptyQ, emptyR := int32(100), int32(100)

	ctx := context.Background()
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		Q: emptyQ,
		R: emptyR,
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Should return only end-turn option (1 option) for empty tile
	if len(resp.Options) > 1 {
		t.Errorf("Expected at most 1 option (end-turn) for empty tile, got %d", len(resp.Options))
	}
}

// Test GetRuntimeGame creates proper world
func TestSingletonGamesService_GetRuntimeGame(t *testing.T) {
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		WorldData:     &v1.WorldData{},
	}

	gamesService := NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.Self = gamesService

	rtGame, err := gamesService.GetRuntimeGame(game, gameState)
	if err != nil {
		t.Fatalf("GetRuntimeGame failed: %v", err)
	}

	// Verify runtime game structure
	if rtGame.Game != game {
		t.Error("Runtime game should reference the game proto")
	}

	if rtGame.GameState != gameState {
		t.Error("Runtime game should reference the game state")
	}

	if rtGame.World == nil {
		t.Error("Runtime game should have a world")
	}

	unitCount := rtGame.World.NumUnits()
	t.Logf("Runtime game has %d units", unitCount)
}
