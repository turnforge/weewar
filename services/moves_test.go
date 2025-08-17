package services

import (
	"context"
	"testing"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupTest(t *testing.T, nq, nr int, units []*v1.Unit) *WasmGamesServiceImpl {
	// 1. Create test world with 3 units
	world := weewar.NewWorld("test")
	// Add some tiles for movement
	for q := range nq {
		for r := range nr {
			coord := weewar.AxialCoord{Q: q, R: r}
			tile := weewar.NewTile(coord, 1) // Grass terrain
			world.AddTile(tile)
		}
	}

	for _, unit := range units {
		world.AddUnit(unit)
	}

	// Load rules engine
	rulesEngine, err := weewar.LoadRulesEngineFromFile(weewar.DevDataPath("data/rules-data.json"))
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create runtime game
	rtGame, err := weewar.NewGame(world, rulesEngine, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	
	// Set current player to 1 for move validation
	rtGame.CurrentPlayer = 1
	
	t.Logf("Game setup: CurrentPlayer=%d, TurnCounter=%d", rtGame.CurrentPlayer, rtGame.TurnCounter)
	
	// Debug: Check if destination tile exists
	destTile := world.TileAt(weewar.AxialCoord{Q: 2, R: 3})
	if destTile == nil {
		t.Logf("WARNING: No tile at destination (2,3)")
	} else {
		t.Logf("Destination tile (2,3) exists: type=%d", destTile.TileType)
	}
	// Create WasmGamesService and set up singleton data
	wasmService := NewWasmGamesServiceImpl()

	// Set up the singleton objects
	wasmService.SingletonGame = &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	wasmService.SingletonGameState = &v1.GameState{
		WorldData: convertRuntimeWorldToProto(world),
		UpdatedAt: timestamppb.Now(),
	}

	wasmService.SingletonGameMoveHistory = &v1.GameMoveHistory{
		Groups: []*v1.GameMoveGroup{},
	}

	wasmService.RuntimeGame = rtGame

	// Verify initial state: N units
	if rtGame.World.NumUnits() != int32(len(units)) {
		t.Fatalf("Expected 3 units initially, got %d", rtGame.World.NumUnits())
	}

	t.Logf("Initial state - units: %d", rtGame.World.NumUnits())
	for coord, unit := range rtGame.World.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d type=%d", coord.Q, coord.R, unit.Player, unit.UnitType)
	}

	return wasmService
}

// Test that reproduces the unit duplication bug using real ProcessMoves with WasmGamesService
func TestProcessMovesNoDuplication(t *testing.T) {
	// Add 3 test units
	units := []*v1.Unit{
		{
			Q:               1,
			R:               2,
			Player:          1,
			UnitType:        1,
			AvailableHealth: 100,
			DistanceLeft:    3,
			TurnCounter:     1,
		},
		{
			Q:               3,
			R:               4,
			Player:          1,
			UnitType:        1,
			AvailableHealth: 100,
			DistanceLeft:    3,
			TurnCounter:     1,
		},
		{
			Q:               0,
			R:               0,
			Player:          2,
			UnitType:        1,
			AvailableHealth: 100,
			DistanceLeft:    3,
			TurnCounter:     1,
		}}

	svc := setupTest(t, 5, 5, units)

	// 2. Call service.ProcessMoves with test move combo
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_MoveUnit{
			MoveUnit: &v1.MoveUnitAction{
				FromQ: 1,
				FromR: 2,
				ToQ:   1,
				ToR:   1,
			},
		},
	}

	req := &v1.ProcessMovesRequest{
		GameId: "test-game",
		Moves:  []*v1.GameMove{move},
	}

	// Call the REAL ProcessMoves method
	resp, err := svc.ProcessMoves(context.Background(), req)
	if err != nil {
		t.Fatalf("ProcessMoves failed: %v", err)
	}

	rtGame := svc.RuntimeGame

	// CRITICAL TEST: Verify we still have exactly 3 units (no duplication)
	finalUnitCount := rtGame.World.NumUnits()
	if finalUnitCount != 3 {
		t.Errorf("UNIT DUPLICATION BUG: Expected 3 units after move, got %d", finalUnitCount)

		// Debug: list all units
		t.Logf("Final units:")
		for coord, unit := range rtGame.World.UnitsByCoord() {
			t.Logf("  Unit at (%d,%d) player=%d type=%d", coord.Q, coord.R, unit.Player, unit.UnitType)
		}
	}

	// Verify the unit moved correctly
	if rtGame.World.UnitAt(weewar.AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position (1,2)")
	}
	if rtGame.World.UnitAt(weewar.AxialCoord{Q: 1, R: 1}) == nil {
		t.Error("Unit not found at new position (1,1)")
	}

	// Verify ProcessMoves response
	if resp == nil {
		t.Fatal("ProcessMoves response is nil")
	}
	if len(resp.MoveResults) == 0 {
		t.Error("Expected move results in response")
	}

	t.Logf("ProcessMoves completed successfully with %d move results", len(resp.MoveResults))
}

// Helper function to convert runtime World to proto (simplified version)
func convertRuntimeWorldToProto(world *weewar.World) *v1.WorldData {
	worldData := &v1.WorldData{
		Units: []*v1.Unit{},
		Tiles: []*v1.Tile{},
	}

	for _, unit := range world.UnitsByCoord() {
		worldData.Units = append(worldData.Units, unit)
	}

	for _, tile := range world.TilesByCoord() {
		worldData.Tiles = append(worldData.Tiles, tile)
	}

	return worldData
}
