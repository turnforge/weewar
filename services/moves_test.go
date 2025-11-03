package services

import (
	"context"
	"testing"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupTest(t *testing.T, nq, nr int, units []*v1.Unit) *SingletonGamesServiceImpl {
	// 1. Create test world with 3 units
	protoWorld := &v1.WorldData{} // Empty world data for test
	world := NewWorld("test", protoWorld)
	// Add some tiles for movement
	for q := range nq {
		for r := range nr {
			coord := AxialCoord{Q: q, R: r}
			tile := NewTile(coord, 1) // Grass terrain
			world.AddTile(tile)
		}
	}

	for _, unit := range units {
		world.AddUnit(unit)
	}

	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create game and state objects for NewGame
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	// Create runtime game
	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Set current player to 1 for move validation
	rtGame.CurrentPlayer = 1

	t.Logf("Game setup: CurrentPlayer=%d, TurnCounter=%d", rtGame.CurrentPlayer, rtGame.TurnCounter)

	// Debug: Check if destination tile exists
	destTile := world.TileAt(AxialCoord{Q: 2, R: 3})
	if destTile == nil {
		t.Logf("WARNING: No tile at destination (2,3)")
	} else {
		t.Logf("Destination tile (2,3) exists: type=%d", destTile.TileType)
	}
	// Create SingletonGamesService and set up singleton data
	singletonService := NewSingletonGamesServiceImpl()

	// Set up the singleton objects (reuse the ones we created)
	singletonService.SingletonGame = game

	singletonService.SingletonGameState = gameState
	singletonService.SingletonGameState.WorldData = convertRuntimeWorldToProto(world)
	singletonService.SingletonGameState.UpdatedAt = timestamppb.Now()

	singletonService.SingletonGameMoveHistory = &v1.GameMoveHistory{
		Groups: []*v1.GameMoveGroup{},
	}

	singletonService.RuntimeGame = rtGame

	// Verify initial state: N units
	if rtGame.World.NumUnits() != int32(len(units)) {
		t.Fatalf("Expected 3 units initially, got %d", rtGame.World.NumUnits())
	}

	t.Logf("Initial state - units: %d", rtGame.World.NumUnits())
	for coord, unit := range rtGame.World.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d type=%d", coord.Q, coord.R, unit.Player, unit.UnitType)
	}

	return singletonService
}

// Test that reproduces the unit duplication bug using real ProcessMoves with SingletonGamesService
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
			// TurnCounter removed: Units will be lazily topped-up when accessed
		},
		{
			Q:               3,
			R:               4,
			Player:          1,
			UnitType:        1,
			AvailableHealth: 100,
			DistanceLeft:    3,
			// TurnCounter removed: Units will be lazily topped-up when accessed
		},
		{
			Q:               0,
			R:               0,
			Player:          2,
			UnitType:        1,
			AvailableHealth: 100,
			DistanceLeft:    3,
			// TurnCounter removed: Units will be lazily topped-up when accessed
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
	if rtGame.World.UnitAt(AxialCoord{Q: 1, R: 2}) != nil {
		t.Error("Unit still found at old position (1,2)")
	}
	if rtGame.World.UnitAt(AxialCoord{Q: 1, R: 1}) == nil {
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
func convertRuntimeWorldToProto(world *World) *v1.WorldData {
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

// TestProcessEndTurnIncome tests that income is calculated correctly based on owned terrain types
func TestProcessEndTurnIncome(t *testing.T) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with different income-generating tiles
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add tiles for player 1
	// Land Base (ID 1): income 100
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, 1)
	baseTile.Player = 1
	world.AddTile(baseTile)

	// Naval Base (ID 2): income 150
	harborTile := NewTile(AxialCoord{Q: 1, R: 0}, 2)
	harborTile.Player = 1
	world.AddTile(harborTile)

	// Airport (ID 3): income 200
	airportTile := NewTile(AxialCoord{Q: 2, R: 0}, 3)
	airportTile.Player = 1
	world.AddTile(airportTile)

	// Missile Silo (ID 16): income 300
	siloTile := NewTile(AxialCoord{Q: 3, R: 0}, 16)
	siloTile.Player = 1
	world.AddTile(siloTile)

	// Non-income tile for player 1 (Grass)
	grassTile := NewTile(AxialCoord{Q: 4, R: 0}, 4)
	grassTile.Player = 1
	world.AddTile(grassTile)

	// Land Base for player 2
	player2Base := NewTile(AxialCoord{Q: 5, R: 0}, 1)
	player2Base.Player = 2
	world.AddTile(player2Base)

	// Create game configuration with initial coins
	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, Coins: 500, StartingCoins: 500},
			{PlayerId: 2, Coins: 300, StartingCoins: 300},
		},
	}

	game := &v1.Game{
		Id:     "test-game",
		Name:   "Test Game",
		Config: gameConfig,
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn for player 1
	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	result, err := processor.ProcessEndTurn(rtGame, move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify income was calculated correctly
	// Expected income for player 1: 100 (base) + 150 (harbor) + 200 (airport) + 300 (silo) = 750
	expectedIncome := int32(750)
	expectedCoins := int32(500) + expectedIncome // 500 starting + 750 income = 1250

	// Check player coins were updated
	player1 := rtGame.Config.Players[0]
	if player1.Coins != expectedCoins {
		t.Errorf("Player 1 coins after end turn: got %d, want %d (initial 500 + income %d)",
			player1.Coins, expectedCoins, expectedIncome)
	}

	// Verify CoinsChangedChange was recorded
	hasCoinsChange := false
	for _, change := range result.Changes {
		if coinsChange := change.GetCoinsChanged(); coinsChange != nil {
			hasCoinsChange = true
			if coinsChange.PlayerId != 1 {
				t.Errorf("CoinsChangedChange player: got %d, want 1", coinsChange.PlayerId)
			}
			if coinsChange.PreviousCoins != 500 {
				t.Errorf("Previous coins: got %d, want 500", coinsChange.PreviousCoins)
			}
			if coinsChange.NewCoins != expectedCoins {
				t.Errorf("New coins: got %d, want %d", coinsChange.NewCoins, expectedCoins)
			}
		}
	}

	if !hasCoinsChange {
		t.Error("Expected CoinsChangedChange in result, but not found")
	}

	// Verify current player changed to player 2
	if rtGame.CurrentPlayer != 2 {
		t.Errorf("Current player after end turn: got %d, want 2", rtGame.CurrentPlayer)
	}
}

// TestProcessEndTurnNoIncome tests end turn with no income-generating tiles
func TestProcessEndTurnNoIncome(t *testing.T) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with only non-income tiles
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add only grass tiles for player 1 (no income)
	grassTile1 := NewTile(AxialCoord{Q: 0, R: 0}, 4)
	grassTile1.Player = 1
	world.AddTile(grassTile1)

	grassTile2 := NewTile(AxialCoord{Q: 1, R: 0}, 4)
	grassTile2.Player = 1
	world.AddTile(grassTile2)

	// Create game configuration
	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, Coins: 200, StartingCoins: 200},
			{PlayerId: 2, Coins: 300, StartingCoins: 300},
		},
	}

	game := &v1.Game{
		Id:     "test-game",
		Name:   "Test Game",
		Config: gameConfig,
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn for player 1
	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	result, err := processor.ProcessEndTurn(rtGame, move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify no income was added (coins stay the same)
	player1 := rtGame.Config.Players[0]
	if player1.Coins != 200 {
		t.Errorf("Player 1 coins after end turn: got %d, want 200 (no income generated)", player1.Coins)
	}

	// Verify no CoinsChangedChange was recorded (since income was 0)
	for _, change := range result.Changes {
		if coinsChange := change.GetCoinsChanged(); coinsChange != nil {
			t.Errorf("Expected no CoinsChangedChange when income is 0, but got change with %d income",
				coinsChange.NewCoins-coinsChange.PreviousCoins)
		}
	}
}

// TestProcessEndTurnMultipleSameType tests income from multiple bases of the same type
func TestProcessEndTurnMultipleSameType(t *testing.T) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with multiple land bases
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add 5 land bases for player 1 (each generates 100 income)
	for i := 0; i < 5; i++ {
		baseTile := NewTile(AxialCoord{Q: i, R: 0}, 1)
		baseTile.Player = 1
		world.AddTile(baseTile)
	}

	// Create game configuration
	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, Coins: 1000, StartingCoins: 1000},
			{PlayerId: 2, Coins: 500, StartingCoins: 500},
		},
	}

	game := &v1.Game{
		Id:     "test-game",
		Name:   "Test Game",
		Config: gameConfig,
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn
	processor := &MoveProcessor{}
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	result, err := processor.ProcessEndTurn(rtGame, move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify income: 5 bases * 100 = 500
	expectedIncome := int32(500)
	expectedCoins := int32(1000) + expectedIncome

	player1 := rtGame.Config.Players[0]
	if player1.Coins != expectedCoins {
		t.Errorf("Player 1 coins after end turn: got %d, want %d (1000 + 500 income)",
			player1.Coins, expectedCoins)
	}

	// Verify CoinsChangedChange
	hasCoinsChange := false
	for _, change := range result.Changes {
		if coinsChange := change.GetCoinsChanged(); coinsChange != nil {
			hasCoinsChange = true
			actualIncome := coinsChange.NewCoins - coinsChange.PreviousCoins
			if actualIncome != expectedIncome {
				t.Errorf("Income: got %d, want %d", actualIncome, expectedIncome)
			}
		}
	}

	if !hasCoinsChange {
		t.Error("Expected CoinsChangedChange in result")
	}
}
