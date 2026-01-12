package tests

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services/singleton"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupTest(t *testing.T, nq, nr int, units []*v1.Unit) *singleton.SingletonGamesService {
	// 1. Create test world with 3 units
	protoWorld := &v1.WorldData{} // Empty world data for test
	world := lib.NewWorld("test", protoWorld)
	// Add some tiles for movement
	for q := range nq {
		for r := range nr {
			coord := lib.AxialCoord{Q: q, R: r}
			tile := lib.NewTile(coord, 1) // Grass terrain
			world.AddTile(tile)
		}
	}

	for _, unit := range units {
		world.AddUnit(unit)
	}

	// Load rules engine
	rulesEngine, err := lib.LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create game and state objects for NewGame
	// Include player config with user_id for authorization
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: TestUserID},
				{PlayerId: 2, UserId: "player-2"},
			},
		},
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
	singletonService := singleton.NewSingletonGamesService()

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
			AvailableHealth: 10,
			DistanceLeft:    3,
			// TurnCounter removed: Units will be lazily topped-up when accessed
		},
		{
			Q:               3,
			R:               4,
			Player:          1,
			UnitType:        1,
			AvailableHealth: 10,
			DistanceLeft:    3,
			// TurnCounter removed: Units will be lazily topped-up when accessed
		},
		{
			Q:               0,
			R:               0,
			Player:          2,
			UnitType:        1,
			AvailableHealth: 10,
			DistanceLeft:    3,
			// TurnCounter removed: Units will be lazily topped-up when accessed
		}}

	svc := setupTest(t, 5, 5, units)

	// 2. Call service.ProcessMoves with test move combo
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_MoveUnit{
			MoveUnit: &v1.MoveUnitAction{
				From: &v1.Position{Q: 1, R: 2},
				To:   &v1.Position{Q: 1, R: 1},
			},
		},
	}

	req := &v1.ProcessMovesRequest{
		GameId: "test-game",
		Moves:  []*v1.GameMove{move},
	}

	// Call the REAL ProcessMoves method with authenticated context
	resp, err := svc.ProcessMoves(AuthenticatedContext(), req)
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
	if len(resp.Moves) == 0 {
		t.Error("Expected move results in response")
	}

	t.Logf("ProcessMoves completed successfully with %d move results", len(resp.Moves))
}

// Helper function to convert runtime World to proto (simplified version)
func convertRuntimeWorldToProto(world *World) *v1.WorldData {
	worldData := &v1.WorldData{
		UnitsMap: map[string]*v1.Unit{},
		TilesMap: map[string]*v1.Tile{},
	}

	for _, unit := range world.UnitsByCoord() {
		worldData.UnitsMap[lib.CoordKey(unit.Q, unit.R)] = unit
	}

	for _, tile := range world.TilesByCoord() {
		worldData.TilesMap[lib.CoordKey(tile.Q, tile.R)] = tile
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
			{PlayerId: 1, StartingCoins: 500},
			{PlayerId: 2, StartingCoins: 300},
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
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
			2: {Coins: 300, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn for player 1
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	err = rtGame.ProcessEndTurn(move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify income was calculated correctly
	// Expected income for player 1: 100 (base) + 150 (harbor) + 200 (airport) + 300 (silo) = 750
	expectedIncome := int32(750)
	expectedCoins := int32(500) + expectedIncome // 500 starting + 750 income = 1250

	// Check player coins were updated (from GameState.PlayerStates)
	player1Coins := rtGame.GameState.PlayerStates[1].Coins
	if player1Coins != expectedCoins {
		t.Errorf("Player 1 coins after end turn: got %d, want %d (initial 500 + income %d)",
			player1Coins, expectedCoins, expectedIncome)
	}

	// Verify CoinsChangedChange was recorded
	hasCoinsChange := false
	for _, change := range move.Changes {
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
			{PlayerId: 1, StartingCoins: 200},
			{PlayerId: 2, StartingCoins: 300},
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
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 200, IsActive: true},
			2: {Coins: 300, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn for player 1
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	err = rtGame.ProcessEndTurn(move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify no income was added (coins stay the same)
	player1Coins := rtGame.GameState.PlayerStates[1].Coins
	if player1Coins != 200 {
		t.Errorf("Player 1 coins after end turn: got %d, want 200 (no income generated)", player1Coins)
	}

	// Verify no CoinsChangedChange was recorded (since income was 0)
	for _, change := range move.Changes {
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
	for i := range 5 {
		baseTile := NewTile(AxialCoord{Q: i, R: 0}, 1)
		baseTile.Player = 1
		world.AddTile(baseTile)
	}

	// Create game configuration
	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, StartingCoins: 1000},
			{PlayerId: 2, StartingCoins: 500},
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
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 1000, IsActive: true},
			2: {Coins: 500, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	err = rtGame.ProcessEndTurn(move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Verify income: 5 bases * 100 = 500
	expectedIncome := int32(500)
	expectedCoins := int32(1000) + expectedIncome

	player1Coins := rtGame.GameState.PlayerStates[1].Coins
	if player1Coins != expectedCoins {
		t.Errorf("Player 1 coins after end turn: got %d, want %d (1000 + 500 income)",
			player1Coins, expectedCoins)
	}

	// Verify CoinsChangedChange
	hasCoinsChange := false
	for _, change := range move.Changes {
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

// TestProcessEndTurnCustomIncomeConfig tests income calculation using custom IncomeConfig values
func TestProcessEndTurnCustomIncomeConfig(t *testing.T) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create a world with different income-generating tiles
	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add tiles for player 1:
	// Land Base (ID 1)
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, 1)
	baseTile.Player = 1
	world.AddTile(baseTile)

	// Naval Base (ID 2)
	harborTile := NewTile(AxialCoord{Q: 1, R: 0}, 2)
	harborTile.Player = 1
	world.AddTile(harborTile)

	// Airport (ID 3)
	airportTile := NewTile(AxialCoord{Q: 2, R: 0}, 3)
	airportTile.Player = 1
	world.AddTile(airportTile)

	// Create game configuration with CUSTOM income values (different from defaults)
	customIncomeConfig := &v1.IncomeConfig{
		StartingCoins:     1000,
		GameIncome:        50,  // 50 coins just for being in game
		LandbaseIncome:    200, // Custom: 200 instead of default 100
		NavalbaseIncome:   300, // Custom: 300 instead of default 150
		AirportbaseIncome: 400, // Custom: 400 instead of default 200
	}

	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, StartingCoins: 500},
			{PlayerId: 2, StartingCoins: 300},
		},
		IncomeConfigs: customIncomeConfig,
	}

	game := &v1.Game{
		Id:     "test-game",
		Name:   "Test Game",
		Config: gameConfig,
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
			2: {Coins: 300, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	// Process end turn for player 1
	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	err = rtGame.ProcessEndTurn(move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Calculate expected income using CUSTOM config values:
	// 200 (land base) + 300 (naval base) + 400 (airport) + 50 (game income) = 950
	expectedIncome := int32(950)
	expectedCoins := int32(500) + expectedIncome // 500 starting + 950 income = 1450

	// Check player coins were updated (from GameState.PlayerStates)
	player1Coins := rtGame.GameState.PlayerStates[1].Coins
	if player1Coins != expectedCoins {
		t.Errorf("Player 1 coins after end turn: got %d, want %d (initial 500 + custom income %d)",
			player1Coins, expectedCoins, expectedIncome)
	}

	// Verify CoinsChangedChange was recorded with correct values
	hasCoinsChange := false
	for _, change := range move.Changes {
		if coinsChange := change.GetCoinsChanged(); coinsChange != nil {
			hasCoinsChange = true
			actualIncome := coinsChange.NewCoins - coinsChange.PreviousCoins
			if actualIncome != expectedIncome {
				t.Errorf("Income from CoinsChangedChange: got %d, want %d", actualIncome, expectedIncome)
			}
		}
	}

	if !hasCoinsChange {
		t.Error("Expected CoinsChangedChange in result, but not found")
	}
}

// TestProcessEndTurnMinesIncome tests income from mines using custom IncomeConfig
func TestProcessEndTurnMinesIncome(t *testing.T) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add 2 mines for player 1 (ID 20)
	mine1 := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeMines)
	mine1.Player = 1
	world.AddTile(mine1)

	mine2 := NewTile(AxialCoord{Q: 1, R: 0}, lib.TileTypeMines)
	mine2.Player = 1
	world.AddTile(mine2)

	// Custom income config with mines income
	customIncomeConfig := &v1.IncomeConfig{
		MinesIncome: 1000, // 1000 per mine
	}

	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, StartingCoins: 100},
			{PlayerId: 2, StartingCoins: 100},
		},
		IncomeConfigs: customIncomeConfig,
	}

	game := &v1.Game{
		Id:     "test-game",
		Name:   "Test Game",
		Config: gameConfig,
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	err = rtGame.ProcessEndTurn(move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Expected: 2 mines * 1000 = 2000 income
	expectedIncome := int32(2000)
	expectedCoins := int32(100) + expectedIncome

	player1Coins := rtGame.GameState.PlayerStates[1].Coins
	if player1Coins != expectedCoins {
		t.Errorf("Player 1 coins: got %d, want %d (100 + %d mines income)",
			player1Coins, expectedCoins, expectedIncome)
	}
}

// TestProcessEndTurnFallbackToDefaults tests that default income is used when IncomeConfig values are 0
func TestProcessEndTurnFallbackToDefaults(t *testing.T) {
	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	protoWorld := &v1.WorldData{}
	world := NewWorld("test", protoWorld)

	// Add a land base for player 1
	baseTile := NewTile(AxialCoord{Q: 0, R: 0}, lib.TileTypeLandBase)
	baseTile.Player = 1
	world.AddTile(baseTile)

	// IncomeConfig with LandbaseIncome = 0 (should fall back to default 100)
	customIncomeConfig := &v1.IncomeConfig{
		LandbaseIncome: 0, // Zero means use default
	}

	gameConfig := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, StartingCoins: 500},
			{PlayerId: 2, StartingCoins: 500},
		},
		IncomeConfigs: customIncomeConfig,
	}

	game := &v1.Game{
		Id:     "test-game",
		Name:   "Test Game",
		Config: gameConfig,
	}

	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 500, IsActive: true},
			2: {Coins: 500, IsActive: true},
		},
	}

	rtGame := NewGame(game, gameState, world, rulesEngine, 12345)

	move := &v1.GameMove{
		Player: 1,
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	err = rtGame.ProcessEndTurn(move, move.GetEndTurn())
	if err != nil {
		t.Fatalf("ProcessEndTurn failed: %v", err)
	}

	// Should use default land base income of 100
	expectedIncome := int32(100)
	expectedCoins := int32(500) + expectedIncome

	player1Coins := rtGame.GameState.PlayerStates[1].Coins
	if player1Coins != expectedCoins {
		t.Errorf("Player 1 coins: got %d, want %d (should use default income 100)",
			player1Coins, expectedCoins)
	}
}

// TestGetTileIncomeFromConfig tests the GetTileIncomeFromConfig helper function directly
func TestGetTileIncomeFromConfig(t *testing.T) {
	testCases := []struct {
		name         string
		tileType     int32
		incomeConfig *v1.IncomeConfig
		expected     int32
	}{
		{
			name:         "nil config uses default for land base",
			tileType:     lib.TileTypeLandBase,
			incomeConfig: nil,
			expected:     lib.DefaultLandbaseIncome, // 100
		},
		{
			name:         "nil config uses default for naval base",
			tileType:     lib.TileTypeNavalBase,
			incomeConfig: nil,
			expected:     lib.DefaultNavalbaseIncome, // 150
		},
		{
			name:         "nil config uses default for airport",
			tileType:     lib.TileTypeAirport,
			incomeConfig: nil,
			expected:     lib.DefaultAirportbaseIncome, // 200
		},
		{
			name:         "nil config uses default for missile silo",
			tileType:     lib.TileTypeMissileSilo,
			incomeConfig: nil,
			expected:     lib.DefaultMissilesiloIncome, // 300
		},
		{
			name:         "nil config uses default for mines",
			tileType:     lib.TileTypeMines,
			incomeConfig: nil,
			expected:     lib.DefaultMinesIncome, // 500
		},
		{
			name:     "custom land base income",
			tileType: lib.TileTypeLandBase,
			incomeConfig: &v1.IncomeConfig{
				LandbaseIncome: 250,
			},
			expected: 250,
		},
		{
			name:     "custom naval base income",
			tileType: lib.TileTypeNavalBase,
			incomeConfig: &v1.IncomeConfig{
				NavalbaseIncome: 350,
			},
			expected: 350,
		},
		{
			name:     "zero in config falls back to default",
			tileType: lib.TileTypeLandBase,
			incomeConfig: &v1.IncomeConfig{
				LandbaseIncome: 0,
			},
			expected: lib.DefaultLandbaseIncome, // 100
		},
		{
			name:         "unknown tile type returns 0",
			tileType:     999, // Unknown type
			incomeConfig: nil,
			expected:     0,
		},
		{
			name:         "grass tile (non-income) returns 0",
			tileType:     4, // Grass
			incomeConfig: nil,
			expected:     0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lib.GetTileIncomeFromConfig(tc.tileType, tc.incomeConfig)
			if result != tc.expected {
				t.Errorf("GetTileIncomeFromConfig(%d, %v) = %d, want %d",
					tc.tileType, tc.incomeConfig, result, tc.expected)
			}
		})
	}
}
