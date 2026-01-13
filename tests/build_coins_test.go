package tests

import (
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services/fsbe"
)

// TestProcessBuildUnit_DeductsCoins tests that building a unit deducts coins from the player
func TestProcessBuildUnit_DeductsCoins(t *testing.T) {
	// Terrain IDs from lilbattle-rules.json:
	// - 1 = Land Base (builds land units)
	// - 2 = Naval Base (builds naval units)
	// - 3 = Airport Base (builds air units)
	//
	// Unit costs from lilbattle-rules.json:
	// - Unit 1 (Soldier Basic) = 75 coins
	// - Unit 5 (Striker) = 200 coins
	// - Unit 3 (Tank Basic) = 300 coins
	// - Unit 10 (Speedboat) = 200 coins
	// - Unit 17 (Helicopter) = 600 coins
	tests := []struct {
		name          string
		initialCoins  int32
		unitType      int32
		expectedCoins int32
		tileType      int32
	}{
		{
			name:          "building soldier at landbase deducts 75 coins",
			initialCoins:  1000,
			unitType:      1, // Soldier Basic = 75 coins
			expectedCoins: 925,
			tileType:      lib.TileTypeLandBase,
		},
		{
			name:          "building striker at landbase deducts 200 coins",
			initialCoins:  1000,
			unitType:      5, // Striker = 200 coins
			expectedCoins: 800,
			tileType:      lib.TileTypeLandBase,
		},
		{
			name:          "building tank at landbase deducts 300 coins",
			initialCoins:  1000,
			unitType:      3, // Tank Basic = 300 coins
			expectedCoins: 700,
			tileType:      lib.TileTypeLandBase,
		},
		{
			name:          "building speedboat at naval base deducts 200 coins",
			initialCoins:  1000,
			unitType:      10, // Speedboat = 200 coins
			expectedCoins: 800,
			tileType:      lib.TileTypeNavalBase,
		},
		{
			name:          "building helicopter at airport deducts 600 coins",
			initialCoins:  1000,
			unitType:      17, // Helicopter = 600 coins
			expectedCoins: 400,
			tileType:      lib.TileTypeAirport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game with specified initial coins and tile type
			game := createTestGameForBuildCoins(tt.initialCoins, tt.tileType)

			// Verify initial coins (from GameState.PlayerStates)
			playerState := game.GameState.PlayerStates[1]
			if playerState == nil || playerState.Coins != tt.initialCoins {
				t.Fatalf("initial coins should be %d, got %v", tt.initialCoins, playerState)
			}

			// Build the unit
			move := &v1.GameMove{}
			action := &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: tt.unitType,
			}

			err := game.ProcessBuildUnit(move, action)
			if err != nil {
				t.Fatalf("ProcessBuildUnit failed: %v", err)
			}

			// Verify coins were deducted (from GameState.PlayerStates)
			actualCoins := game.GameState.PlayerStates[1].Coins
			if actualCoins != tt.expectedCoins {
				t.Errorf("coins after build: got %d, want %d (deducted %d)",
					actualCoins, tt.expectedCoins, tt.initialCoins-actualCoins)
			}

			// Verify the CoinsChangedChange was recorded
			foundCoinsChange := false
			for _, change := range move.Changes {
				if cc, ok := change.ChangeType.(*v1.WorldChange_CoinsChanged); ok {
					foundCoinsChange = true
					if cc.CoinsChanged.PreviousCoins != tt.initialCoins {
						t.Errorf("CoinsChangedChange.PreviousCoins: got %d, want %d",
							cc.CoinsChanged.PreviousCoins, tt.initialCoins)
					}
					if cc.CoinsChanged.NewCoins != tt.expectedCoins {
						t.Errorf("CoinsChangedChange.NewCoins: got %d, want %d",
							cc.CoinsChanged.NewCoins, tt.expectedCoins)
					}
				}
			}
			if !foundCoinsChange {
				t.Error("expected CoinsChangedChange in move changes, not found")
			}
		})
	}
}

// TestProcessBuildUnit_InsufficientCoins tests that building fails when player doesn't have enough coins
func TestProcessBuildUnit_InsufficientCoins(t *testing.T) {
	tests := []struct {
		name         string
		initialCoins int32
		unitType     int32
		unitCost     int32
	}{
		{
			name:         "cannot build soldier with only 50 coins",
			initialCoins: 50,
			unitType:     1, // Soldier Basic costs 75
			unitCost:     75,
		},
		{
			name:         "cannot build tank with only 200 coins",
			initialCoins: 200,
			unitType:     3, // Tank Basic costs 300
			unitCost:     300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := createTestGameForBuildCoins(tt.initialCoins, lib.TileTypeLandBase)

			move := &v1.GameMove{}
			action := &v1.BuildUnitAction{
				Pos:      &v1.Position{Q: 0, R: 0},
				UnitType: tt.unitType,
			}

			err := game.ProcessBuildUnit(move, action)
			if err == nil {
				t.Errorf("expected error for insufficient coins (have %d, need %d), got nil",
					tt.initialCoins, tt.unitCost)
			}

			// Verify coins were NOT deducted (from GameState.PlayerStates)
			actualCoins := game.GameState.PlayerStates[1].Coins
			if actualCoins != tt.initialCoins {
				t.Errorf("coins should not change on failed build: got %d, want %d",
					actualCoins, tt.initialCoins)
			}
		})
	}
}

// createTestGameForBuildCoins creates a game for testing coin deduction during build
func createTestGameForBuildCoins(initialCoins int32, tileType int32) *lib.Game {
	worldData := &v1.WorldData{
		TilesMap: map[string]*v1.Tile{
			"0,0": {Q: 0, R: 0, TileType: tileType, Player: 1}, // Player 1's base
		},
		UnitsMap: map[string]*v1.Unit{},
	}

	game := &v1.Game{
		Id:      "test-game",
		WorldId: "test-world",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, StartingCoins: initialCoins},
			},
			Settings: &v1.GameSettings{},
		},
	}

	state := &v1.GameState{
		GameId:        "test-game",
		CurrentPlayer: 1,
		TurnCounter:   1,
		WorldData:     worldData,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: initialCoins, IsActive: true},
		},
	}

	rulesEngine := lib.DefaultRulesEngine()
	return lib.NewGame(game, state, lib.NewWorld("test-world", worldData), rulesEngine, 0)
}

// TestBuildUnit_CoinsPersistence tests that coin deduction is persisted after ProcessMoves
// This is an integration test that uses the full GamesService flow
func TestBuildUnit_CoinsPersistence(t *testing.T) {
	ctx := AuthenticatedContext()

	// Create a temp directory for the test
	tempDir, err := os.MkdirTemp("", "lilbattle-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gamesDir := filepath.Join(tempDir, "games")
	if err := os.MkdirAll(gamesDir, 0755); err != nil {
		t.Fatalf("failed to create games dir: %v", err)
	}

	// Copy testgame-template to the temp directory
	gameId := "testgame"
	gameDir := filepath.Join(gamesDir, gameId)
	if err := os.MkdirAll(gameDir, 0755); err != nil {
		t.Fatalf("failed to create game dir: %v", err)
	}

	// Copy all files from testgame-template
	templateDir := "testgame"
	for _, filename := range []string{"metadata.json", "state.json", "history.json"} {
		srcPath := filepath.Join(templateDir, filename)
		dstPath := filepath.Join(gameDir, filename)

		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", srcPath, err)
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", dstPath, err)
		}
	}

	// Create GamesService pointing to temp directory
	gamesService := fsbe.NewFSGamesService(gamesDir, nil)

	// Load the game to get initial coins (from State.PlayerStates)
	getGameResp, err := gamesService.GetGame(ctx, &v1.GetGameRequest{Id: gameId})
	if err != nil {
		t.Fatalf("failed to get game: %v", err)
	}
	playerState := getGameResp.State.PlayerStates[1]
	if playerState == nil {
		t.Fatal("player state not found for player 1")
	}
	initialCoins := playerState.Coins
	t.Logf("Initial coins for player 1: %d", initialCoins)

	// Find a landbase tile owned by player 1 that has no unit on it
	var buildQ, buildR int32
	foundBuildableTile := false
	for _, tile := range getGameResp.State.WorldData.TilesMap {
		if tile.Player == 1 && tile.TileType == lib.TileTypeLandBase {
			// Check if there's no unit at this position
			key := lib.CoordKey(tile.Q, tile.R)
			if _, hasUnit := getGameResp.State.WorldData.UnitsMap[key]; !hasUnit {
				buildQ, buildR = tile.Q, tile.R
				foundBuildableTile = true
				t.Logf("Found buildable tile at (%d, %d)", buildQ, buildR)
				break
			}
		}
	}
	if !foundBuildableTile {
		t.Fatal("no buildable landbase tile found for player 1")
	}

	// Build a soldier (costs 75 coins)
	_, err = gamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gameId,
		Moves: []*v1.GameMove{
			{
				MoveType: &v1.GameMove_BuildUnit{
					BuildUnit: &v1.BuildUnitAction{
						Pos:      &v1.Position{Q: buildQ, R: buildR},
						UnitType: 1, // Soldier Basic = 75 coins
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ProcessMoves failed: %v", err)
	}

	// Reload the game from disk to verify persistence
	getGameResp2, err := gamesService.GetGame(ctx, &v1.GetGameRequest{Id: gameId})
	if err != nil {
		t.Fatalf("failed to reload game: %v", err)
	}

	// Check coins after reload - should be 75 less than initial (from State.PlayerStates)
	expectedCoins := initialCoins - 75
	playerState2 := getGameResp2.State.PlayerStates[1]
	if playerState2 == nil {
		t.Fatal("player state not found for player 1 after reload")
	}
	actualCoins := playerState2.Coins
	if actualCoins != expectedCoins {
		t.Errorf("PERSISTENCE BUG: coins after reload: got %d, want %d (75 deducted from %d)",
			actualCoins, expectedCoins, initialCoins)
	}

	// Also verify the unit was created
	key := lib.CoordKey(buildQ, buildR)
	if _, hasUnit := getGameResp2.State.WorldData.UnitsMap[key]; !hasUnit {
		t.Errorf("expected unit at (%d, %d) after build, not found", buildQ, buildR)
	}
}
