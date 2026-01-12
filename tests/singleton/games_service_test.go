package singleton

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/services"
	"github.com/turnforge/lilbattle/services/singleton"
	"github.com/turnforge/lilbattle/tests"
)

// Test that we can load a game and get options for units
func TestSingletonGamesService_GetOptionsAt(t *testing.T) {
	// Load a real test world using the existing test utility
	homeDir, _ := os.UserHomeDir()
	worldsDir := filepath.Join(homeDir, "dev-app-data", "lilbattle", "storage", "worlds")

	world, gameState, err := tests.LoadTestWorldFromStorage(worldsDir, "32112070")
	if err != nil {
		t.Skipf("Skipping test - world data not available: %v", err)
	}

	// Create game
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	// Create service
	gamesService := singleton.NewSingletonGamesService()
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
		Pos: &v1.Position{
			Q: int32(testCoord.Q),
			R: int32(testCoord.R),
		},
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
			t.Logf("  Option %d: Move to (%d,%d)", i, moveOpt.To.Q, moveOpt.To.R)
			if moveOpt.ReconstructedPath != nil {
				t.Logf("    Path has %d edges", len(moveOpt.ReconstructedPath.Edges))
			}
		} else if attackOpt := opt.GetAttack(); attackOpt != nil {
			t.Logf("  Option %d: Attack at (%d,%d)", i, attackOpt.Defender.Q, attackOpt.Defender.R)
		} else if captureOpt := opt.GetCapture(); captureOpt != nil {
			t.Logf("  Option %d: Capture at (%d,%d)", i, captureOpt.Pos.Q, captureOpt.Pos.R)
		}
	}
}

// Test getting options for empty tile returns empty
func TestSingletonGamesService_GetOptionsAt_EmptyTile(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	worldsDir := filepath.Join(homeDir, "dev-app-data", "lilbattle", "storage", "worlds")

	_, gameState, err := tests.LoadTestWorldFromStorage(worldsDir, "32112070")
	if err != nil {
		t.Skipf("Skipping test - world data not available: %v", err)
	}

	game := &v1.Game{Id: "test-game"}

	gamesService := singleton.NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.Self = gamesService

	// Find an empty tile (try coordinates unlikely to have units)
	emptyQ, emptyR := int32(100), int32(100)

	ctx := context.Background()
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		Pos: &v1.Position{
			Q: int32(emptyQ),
			R: int32(emptyR),
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Should return only end-turn option (1 option) for empty tile
	if len(resp.Options) > 1 {
		t.Errorf("Expected at most 1 option (end-turn) for empty tile, got %d", len(resp.Options))
	}
}

// Test that build options are not returned when a unit is on a base tile
func TestSingletonGamesService_GetOptionsAt_NoBuildWhenUnitOnTile(t *testing.T) {
	// Create a minimal game state with a base tile and a unit on the same position
	testQ, testR := int32(0), int32(0)

	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Settings: &v1.GameSettings{},
		},
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 1000}, // Plenty of coins to afford any unit
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0": {
					Q:        testQ,
					R:        testR,
					TileType: 1, // Base (buildable tile)
					Player:   1, // Owned by current player
				},
			},
			UnitsMap: map[string]*v1.Unit{
				"0,0": {
					Q:               testQ,
					R:               testR,
					Player:          1,
					UnitType:        1, // Trooper
					AvailableHealth: 10,
					DistanceLeft:    3,
				},
			},
		},
	}

	gamesService := singleton.NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.SingletonGameMoveHistory = &v1.GameMoveHistory{}
	gamesService.Self = gamesService

	ctx := context.Background()

	// Get options at the tile position (where unit is)
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: int32(testQ),
			R: int32(testR),
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Count build options - there should be none since unit is on the tile
	buildOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetBuild() != nil {
			buildOptionCount++
		}
	}

	if buildOptionCount > 0 {
		t.Errorf("Expected 0 build options when unit is on tile, got %d", buildOptionCount)
	}

	t.Logf("Got %d total options, %d build options (expected 0 build options)", len(resp.Options), buildOptionCount)
}

// Test that build options ARE returned when no unit is on a base tile
func TestSingletonGamesService_GetOptionsAt_BuildWhenNoUnitOnTile(t *testing.T) {
	// Create a minimal game state with a base tile and NO unit on it
	testQ, testR := int32(0), int32(0)

	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Settings: &v1.GameSettings{},
		},
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 1000}, // Plenty of coins to afford any unit
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0": {
					Q:        testQ,
					R:        testR,
					TileType: 1, // Base (buildable tile)
					Player:   1, // Owned by current player
				},
			},
			UnitsMap: map[string]*v1.Unit{}, // No units
		},
	}

	gamesService := singleton.NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.SingletonGameMoveHistory = &v1.GameMoveHistory{}
	gamesService.Self = gamesService

	ctx := context.Background()

	// Get options at the tile position (no unit)
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: int32(testQ),
			R: int32(testR),
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Count build options - there should be some since no unit is on the tile
	buildOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetBuild() != nil {
			buildOptionCount++
		}
	}

	if buildOptionCount == 0 {
		t.Errorf("Expected build options when no unit is on base tile, got 0")
	}

	t.Logf("Got %d total options, %d build options", len(resp.Options), buildOptionCount)
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

	gamesService := singleton.NewSingletonGamesService()
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
