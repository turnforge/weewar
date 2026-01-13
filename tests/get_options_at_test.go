package tests

import (
	"context"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetOptionsAtTestCase defines a test case for GetOptionsAt
type GetOptionsAtTestCase struct {
	Name           string
	Q              int32
	R              int32
	ExpectedResult *GetOptionsAtExpectation
}

// GetOptionsAtExpectation defines what we expect from GetOptionsAt
type GetOptionsAtExpectation struct {
	// Expected number of options by type
	MoveOptionCount   int
	AttackOptionCount int
	EndTurnCount      int
	TotalOptionCount  int

	// Specific coordinate checks for movement options
	ExpectedMoveCoords []services.AxialCoord

	// Specific coordinate checks for attack options
	ExpectedAttackCoords []services.AxialCoord

	// Game state expectations
	CurrentPlayer   int32
	GameInitialized bool
}

// GetOptionsAtTestScenario defines a complete test scenario using real world data
type GetOptionsAtTestScenario struct {
	Name             string
	Description      string
	WorldsStorageDir string // Directory containing world data (e.g. ~/dev-app-data/lilbattle/storage/worlds)
	WorldId          string // ID of world to load via GetWorld RPC
	CurrentPlayer    int32  // Override current player
	TurnCounter      int32  // Override turn counter
	TestCases        []GetOptionsAtTestCase
}

// setupGetOptionsAtTest creates a test service loading real world data from storage
func setupGetOptionsAtTest(t *testing.T, scenario GetOptionsAtTestScenario) *SingletonGamesService {
	// Load real world data from storage using FSWorldsService
	rtWorld, gameState, err := LoadTestWorldFromStorage(scenario.WorldsStorageDir, scenario.WorldId)
	if err != nil {
		t.Fatalf("Failed to load world %s from %s: %v",
			scenario.WorldId, scenario.WorldsStorageDir, err)
	}

	// Load rules engine
	rulesEngine, err := LoadRulesEngineFromFile(RULES_DATA_FILE, DAMAGE_DATA_FILE)
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create game and state objects for NewGame
	game := &v1.Game{
		Id:   "test-game-" + scenario.WorldId,
		Name: "Test Game - " + scenario.WorldId,
	}

	// Create runtime game
	rtGame := NewGame(game, gameState, rtWorld, rulesEngine, 12345)

	// Apply scenario overrides
	if scenario.CurrentPlayer > 0 {
		rtGame.CurrentPlayer = scenario.CurrentPlayer
		gameState.CurrentPlayer = scenario.CurrentPlayer
	}
	if scenario.TurnCounter > 0 {
		rtGame.TurnCounter = scenario.TurnCounter
		gameState.TurnCounter = scenario.TurnCounter
	}

	// Create SingletonGamesService
	singletonService := NewSingletonGamesService()

	// Set up the singleton objects
	singletonService.SingletonGame = &v1.Game{
		Id:   "test-game-get-options-" + scenario.WorldId,
		Name: "Get Options Test Game - " + scenario.WorldId,
	}

	singletonService.SingletonGameState = gameState
	singletonService.SingletonGameState.UpdatedAt = timestamppb.Now()

	singletonService.SingletonGameMoveHistory = &v1.GameMoveHistory{
		Groups: []*v1.GameMoveGroup{},
	}

	singletonService.RuntimeGame = rtGame

	t.Logf("Test setup complete - World=%s, CurrentPlayer=%d, TurnCounter=%d, Units=%d",
		scenario.WorldId, rtGame.CurrentPlayer, rtGame.TurnCounter, rtGame.World.NumUnits())

	// Log units for debugging
	for coord, unit := range rtGame.World.UnitsByCoord() {
		t.Logf("  Unit at (%d,%d) player=%d type=%d health=%d movement=%f",
			coord.Q, coord.R, unit.Player, unit.UnitType, unit.AvailableHealth, unit.DistanceLeft)
	}

	return singletonService
}

// runGetOptionsAtTest executes a single GetOptionsAt test case
func runGetOptionsAtTest(t *testing.T, svc *SingletonGamesService, testCase GetOptionsAtTestCase) {
	t.Run(testCase.Name, func(t *testing.T) {
		// Call GetOptionsAt
		req := &v1.GetOptionsAtRequest{
			GameId: svc.SingletonGame.Id,
			Pos:    &v1.Position{Q: testCase.Q, R: testCase.R},
		}

		resp, err := svc.GetOptionsAt(context.Background(), req)
		if err != nil {
			t.Fatalf("GetOptionsAt failed: %v", err)
		}

		exp := testCase.ExpectedResult

		// Check basic response structure
		if resp.CurrentPlayer != exp.CurrentPlayer {
			t.Errorf("Expected CurrentPlayer=%d, got %d", exp.CurrentPlayer, resp.CurrentPlayer)
		}

		if resp.GameInitialized != exp.GameInitialized {
			t.Errorf("Expected GameInitialized=%v, got %v", exp.GameInitialized, resp.GameInitialized)
		}

		// Count option types
		moveCount := 0
		attackCount := 0
		endTurnCount := 0

		moveCoords := []AxialCoord{}
		attackCoords := []AxialCoord{}

		for _, option := range resp.Options {
			switch optionType := option.OptionType.(type) {
			case *v1.GameOption_Move:
				moveCount++
				moveCoords = append(moveCoords, AxialCoord{
					Q: int(optionType.Move.To.Q),
					R: int(optionType.Move.To.R),
				})
			case *v1.GameOption_Attack:
				attackCount++
				attackCoords = append(attackCoords, AxialCoord{
					Q: int(optionType.Attack.Defender.Q),
					R: int(optionType.Attack.Defender.R),
				})
			case *v1.GameOption_EndTurn:
				endTurnCount++
			}
		}

		// Verify option counts
		if moveCount != exp.MoveOptionCount {
			t.Errorf("Expected %d move options, got %d", exp.MoveOptionCount, moveCount)
			for i, coord := range moveCoords {
				t.Logf("  Move option %d: (%d,%d)", i, coord.Q, coord.R)
			}
		}

		if attackCount != exp.AttackOptionCount {
			t.Errorf("Expected %d attack options, got %d", exp.AttackOptionCount, attackCount)
			for i, coord := range attackCoords {
				t.Logf("  Attack option %d: (%d,%d)", i, coord.Q, coord.R)
			}
		}

		if endTurnCount != exp.EndTurnCount {
			t.Errorf("Expected %d end turn options, got %d", exp.EndTurnCount, endTurnCount)
		}

		if len(resp.Options) != exp.TotalOptionCount {
			t.Errorf("Expected %d total options, got %d", exp.TotalOptionCount, len(resp.Options))
		}

		// Verify specific coordinates if provided
		if exp.ExpectedMoveCoords != nil {
			if !coordSlicesEqual(moveCoords, exp.ExpectedMoveCoords) {
				t.Errorf("Move coordinates mismatch. Expected: %v, Got: %v",
					exp.ExpectedMoveCoords, moveCoords)
			}
		}

		if exp.ExpectedAttackCoords != nil {
			if !coordSlicesEqual(attackCoords, exp.ExpectedAttackCoords) {
				t.Errorf("Attack coordinates mismatch. Expected: %v, Got: %v",
					exp.ExpectedAttackCoords, attackCoords)
			}
		}

		t.Logf("✅ GetOptionsAt(%d,%d) - %d total options (%d move, %d attack, %d endturn)", 
			testCase.Q, testCase.R, len(resp.Options), moveCount, attackCount, endTurnCount)
	})
}

// TestGetOptionsAtWithRealWorlds runs GetOptionsAt tests using real world data from storage
func TestGetOptionsAtWithRealWorlds(t *testing.T) {
	worldsStorageDir := DevDataPath("storage/worlds")

	scenarios := []GetOptionsAtTestScenario{
		{
			Name:             "SmallWorldBasicTest",
			Description:      "Test GetOptionsAt using small-world from world editor",
			WorldsStorageDir: worldsStorageDir,
			WorldId:          "small-world",
			CurrentPlayer:    1,
			TurnCounter:      1,
			TestCases: []GetOptionsAtTestCase{
				{
					Name: "CheckFirstUnit",
					Q:    0, // Adjust based on actual unit positions in small-world
					R:    0,
					ExpectedResult: &GetOptionsAtExpectation{
						// Empty tile - no options (end turn is globally available)
						MoveOptionCount:   0,
						AttackOptionCount: 0,
						EndTurnCount:      0,
						TotalOptionCount:  0,
						CurrentPlayer:     1,
						GameInitialized:   true,
					},
				},
				{
					Name: "CheckEmptyTile",
					Q:    5, // Some tile that should be empty
					R:    5,
					ExpectedResult: &GetOptionsAtExpectation{
						MoveOptionCount:   0,
						AttackOptionCount: 0,
						EndTurnCount:      0,
						TotalOptionCount:  0,
						CurrentPlayer:     1,
						GameInitialized:   true,
					},
				},
			},
		},
		// Add more scenarios for other worlds as needed
		{
			Name:             "AnotherWorldTest",
			Description:      "Test with another world ID",
			WorldsStorageDir: worldsStorageDir,
			WorldId:          "32112070", // One of the UUID-named worlds
			CurrentPlayer:    1,
			TurnCounter:      1,
			TestCases: []GetOptionsAtTestCase{
				{
					Name: "CheckOurUnit",
					Q:    1, // Player 1 unit at (1,1) per actual world data
					R:    1,
					ExpectedResult: &GetOptionsAtExpectation{
						// Our unit should have movement options (end turn is global)
						// Soldier with 3 movement points, limited by map terrain
						MoveOptionCount:   14, // Reachable tiles within 3 movement on this map
						AttackOptionCount: 0,  // Enemy units too far away
						EndTurnCount:      0,
						TotalOptionCount:  14, // Just the 14 moves
						CurrentPlayer:     1,
						GameInitialized:   true,
					},
				},
				{
					Name: "CheckEnemyUnit",
					Q:    4, // Player 2 unit at (4,6) per actual world data
					R:    6,
					ExpectedResult: &GetOptionsAtExpectation{
						// Enemy unit - no options (end turn is global)
						MoveOptionCount:   0,
						AttackOptionCount: 0,
						EndTurnCount:      0,
						TotalOptionCount:  0,
						CurrentPlayer:     1,
						GameInitialized:   true,
					},
				},
				{
					Name: "CheckEmptyTile",
					Q:    0,
					R:    0,
					ExpectedResult: &GetOptionsAtExpectation{
						MoveOptionCount:   0,
						AttackOptionCount: 0,
						EndTurnCount:      0,
						TotalOptionCount:  0,
						CurrentPlayer:     1,
						GameInitialized:   true,
					},
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("🎮 Running scenario: %s - %s", scenario.Name, scenario.Description)
			t.Logf("   Loading world %s from %s", scenario.WorldId, scenario.WorldsStorageDir)

			// Setup test service with scenario's world
			svc := setupGetOptionsAtTest(t, scenario)

			// Run all test cases for this scenario
			for _, testCase := range scenario.TestCases {
				runGetOptionsAtTest(t, svc, testCase)
			}
		})
	}
}

// Helper function to compare coordinate slices
func coordSlicesEqual(a, b []AxialCoord) bool {
	if len(a) != len(b) {
		return false
	}

	// Convert to sets for comparison (order doesn't matter)
	setA := make(map[AxialCoord]bool)
	setB := make(map[AxialCoord]bool)

	for _, coord := range a {
		setA[coord] = true
	}
	for _, coord := range b {
		setB[coord] = true
	}

	for coord := range setA {
		if !setB[coord] {
			return false
		}
	}

	return true
}
