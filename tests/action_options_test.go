package tests

import (
	"context"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/services/singleton"
)

// TestGetOptionsAt_AttackAfterMove tests that a unit with action_order ["move", "attack"]
// can still attack after moving, even if it hasn't exhausted all movement points.
// This is a common scenario where the player moves a unit and then wants to attack.
func TestGetOptionsAt_AttackAfterMove(t *testing.T) {
	// Create a scenario:
	// - Anti-aircraft (type 6) at position 1,0 with action_order ["move", "attack"]
	// - Enemy unit at position 2,0 (within attack range)
	// - Unit has already moved this turn (progression_step=0, but no more move tiles available)
	// - Unit should be able to attack the enemy

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
			1: {Coins: 1000},
			2: {Coins: 1000},
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0": {Q: 0, R: 0, TileType: 5}, // Grass
				"1,0": {Q: 1, R: 0, TileType: 5}, // Grass - our unit here
				"2,0": {Q: 2, R: 0, TileType: 5}, // Grass - enemy here
			},
			UnitsMap: map[string]*v1.Unit{
				"1,0": {
					Q:                1,
					R:                0,
					Player:           1,
					UnitType:         6, // Anti-aircraft - action_order: ["move", "attack"]
					AvailableHealth:  10,
					DistanceLeft:     0.5, // Some movement left, but not enough to move anywhere useful
					ProgressionStep:  0,   // Still at move step
					LastToppedupTurn: 1,
					Shortcut:         "A1",
				},
				"2,0": {
					Q:                2,
					R:                0,
					Player:           2, // Enemy
					UnitType:         1, // Trooper
					AvailableHealth:  10,
					DistanceLeft:     3,
					LastToppedupTurn: 1,
					Shortcut:         "B1",
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

	// Get options for our unit at 1,0
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: 1,
			R: 0,
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	// Count attack options - there should be at least one (attack enemy at 2,0)
	attackOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetAttack() != nil {
			attackOptionCount++
			t.Logf("Found attack option: attack at (%d,%d)", opt.GetAttack().Defender.Q, opt.GetAttack().Defender.R)
		}
		if opt.GetMove() != nil {
			t.Logf("Found move option: move to (%d,%d)", opt.GetMove().To.Q, opt.GetMove().To.R)
		}
	}

	if attackOptionCount == 0 {
		t.Errorf("Expected attack options for unit that can still attack after move, got 0 attack options")
		t.Logf("Total options: %d", len(resp.Options))
	}
}

// TestGetOptionsAt_AfterFullMove tests that after a unit exhausts all movement,
// it should show attack options if action_order allows attack after move.
func TestGetOptionsAt_AfterFullMove(t *testing.T) {
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
			1: {Coins: 1000},
			2: {Coins: 1000},
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0": {Q: 0, R: 0, TileType: 5},
				"1,0": {Q: 1, R: 0, TileType: 5},
				"2,0": {Q: 2, R: 0, TileType: 5},
			},
			UnitsMap: map[string]*v1.Unit{
				"1,0": {
					Q:                1,
					R:                0,
					Player:           1,
					UnitType:         6, // Anti-aircraft - action_order: ["move", "attack"]
					AvailableHealth:  10,
					DistanceLeft:     0, // No movement left
					ProgressionStep:  1, // Advanced to attack step
					LastToppedupTurn: 1,
					Shortcut:         "A1",
				},
				"2,0": {
					Q:                2,
					R:                0,
					Player:           2,
					UnitType:         1,
					AvailableHealth:  10,
					DistanceLeft:     3,
					LastToppedupTurn: 1,
					Shortcut:         "B1",
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

	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: 1,
			R: 0,
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	attackOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetAttack() != nil {
			attackOptionCount++
		}
	}

	if attackOptionCount == 0 {
		t.Errorf("Expected attack options after full move (progression_step=1), got 0")
	}

	t.Logf("Got %d total options, %d attack options", len(resp.Options), attackOptionCount)
}

// TestGetOptionsAt_PartialMoveShowsBothMoveAndAttack tests that after a partial move,
// the unit should show BOTH remaining move options AND attack options.
// This is the helicopter scenario: move TR,TR,TR then should be able to attack.
func TestGetOptionsAt_PartialMoveShowsBothMoveAndAttack(t *testing.T) {
	// Scenario: Helicopter (type 17) with action_order ["move", "attack", "retreat"]
	// After partial move, still has DistanceLeft=2, progression_step=0
	// Should show BOTH move options (can still move) AND attack options (enemy adjacent)

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
			1: {Coins: 1000},
			2: {Coins: 1000},
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0":  {Q: 0, R: 0, TileType: 5},  // Grass
				"1,0":  {Q: 1, R: 0, TileType: 5},  // Grass - our helicopter here after partial move
				"2,0":  {Q: 2, R: 0, TileType: 5},  // Grass - enemy helicopter here
				"-1,0": {Q: -1, R: 0, TileType: 5}, // Grass - can still move here
				"0,-1": {Q: 0, R: -1, TileType: 5}, // Grass - can still move here
			},
			UnitsMap: map[string]*v1.Unit{
				"1,0": {
					Q:                1,
					R:                0,
					Player:           1,
					UnitType:         17, // Helicopter - action_order: ["move", "attack", "retreat"]
					AvailableHealth:  10,
					DistanceLeft:     2, // Still has movement left after partial move
					ProgressionStep:  0, // Still at move step
					LastToppedupTurn: 1,
					Shortcut:         "A5",
				},
				"2,0": {
					Q:                2,
					R:                0,
					Player:           2, // Enemy helicopter
					UnitType:         17,
					AvailableHealth:  10,
					DistanceLeft:     5,
					LastToppedupTurn: 1,
					Shortcut:         "B3",
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

	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: 1,
			R: 0,
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	moveOptionCount := 0
	attackOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetMove() != nil {
			moveOptionCount++
		}
		if opt.GetAttack() != nil {
			attackOptionCount++
			t.Logf("Found attack option: attack at (%d,%d)", opt.GetAttack().Defender.Q, opt.GetAttack().Defender.R)
		}
	}

	t.Logf("Move options: %d, Attack options: %d, Total: %d", moveOptionCount, attackOptionCount, len(resp.Options))

	// Should have BOTH move options (can still move with remaining MP)
	// AND attack options (enemy adjacent and attack is next step)
	if moveOptionCount == 0 {
		t.Errorf("Expected move options (unit still has DistanceLeft), got 0")
	}

	if attackOptionCount == 0 {
		t.Errorf("Expected attack options after partial move, got 0")
	}
}

// TestGetOptionsAt_NoMoveOptionsAutoAdvance tests that when a unit has no valid move options
// (even with some DistanceLeft), it should auto-advance to the next action in action_order.
func TestGetOptionsAt_NoMoveOptionsAutoAdvance(t *testing.T) {
	// Scenario: Unit is surrounded by occupied tiles or impassable terrain
	// Even though DistanceLeft > 0 and ProgressionStep = 0, there are no move options
	// The system should auto-advance to allow attack

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
			1: {Coins: 1000},
			2: {Coins: 1000},
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0":  {Q: 0, R: 0, TileType: 5},  // Grass - friendly unit
				"1,0":  {Q: 1, R: 0, TileType: 5},  // Grass - our test unit
				"-1,0": {Q: -1, R: 0, TileType: 5}, // Grass - friendly unit blocking
				"0,-1": {Q: 0, R: -1, TileType: 5}, // Grass - friendly unit blocking
				"1,-1": {Q: 1, R: -1, TileType: 5}, // Grass - friendly unit blocking
				"0,1":  {Q: 0, R: 1, TileType: 5},  // Grass - enemy unit here (can attack)
				"-1,1": {Q: -1, R: 1, TileType: 5}, // Grass - friendly unit blocking
			},
			UnitsMap: map[string]*v1.Unit{
				// Our unit in the center
				"1,0": {
					Q:                1,
					R:                0,
					Player:           1,
					UnitType:         6, // Anti-aircraft
					AvailableHealth:  10,
					DistanceLeft:     3, // Has full movement points
					ProgressionStep:  0, // At move step
					LastToppedupTurn: 1,
					Shortcut:         "A1",
				},
				// Surrounding friendly units (blocking movement)
				"0,0": {
					Q: 0, R: 0, Player: 1, UnitType: 1, AvailableHealth: 10, DistanceLeft: 3, LastToppedupTurn: 1,
				},
				"-1,0": {
					Q: -1, R: 0, Player: 1, UnitType: 1, AvailableHealth: 10, DistanceLeft: 3, LastToppedupTurn: 1,
				},
				"0,-1": {
					Q: 0, R: -1, Player: 1, UnitType: 1, AvailableHealth: 10, DistanceLeft: 3, LastToppedupTurn: 1,
				},
				"1,-1": {
					Q: 1, R: -1, Player: 1, UnitType: 1, AvailableHealth: 10, DistanceLeft: 3, LastToppedupTurn: 1,
				},
				"-1,1": {
					Q: -1, R: 1, Player: 1, UnitType: 1, AvailableHealth: 10, DistanceLeft: 3, LastToppedupTurn: 1,
				},
				// Enemy unit within attack range
				"0,1": {
					Q:                0,
					R:                1,
					Player:           2,
					UnitType:         1,
					AvailableHealth:  10,
					DistanceLeft:     3,
					LastToppedupTurn: 1,
					Shortcut:         "B1",
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

	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: 1,
			R: 0,
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	moveOptionCount := 0
	attackOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetMove() != nil {
			moveOptionCount++
		}
		if opt.GetAttack() != nil {
			attackOptionCount++
		}
	}

	// Should have no move options (surrounded)
	// But SHOULD have attack options (enemy in range)
	t.Logf("Move options: %d, Attack options: %d, Total: %d", moveOptionCount, attackOptionCount, len(resp.Options))

	if attackOptionCount == 0 {
		t.Errorf("Expected attack options when surrounded (no move options available), got 0")
	}
}

// TestGetOptionsAt_AfterAttackOnlyRetreat tests that after a unit attacks,
// the next action should be retreat (not attack again).
// Helicopter action_order: ["move", "attack", "retreat"]
// After attacking, progression_step=2, so only retreat (shown as move) should be available.
func TestGetOptionsAt_AfterAttackOnlyRetreat(t *testing.T) {
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
			1: {Coins: 1000},
			2: {Coins: 1000},
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				// Unit at (1,0) - neighbors are: (0,0), (1,-1), (2,-1), (2,0), (1,1), (0,1)
				"1,0": {Q: 1, R: 0, TileType: 5}, // Our helicopter here after attacking
				"2,0": {Q: 2, R: 0, TileType: 5}, // Enemy still here (damaged)
				"0,0": {Q: 0, R: 0, TileType: 5}, // Can retreat here (LEFT neighbor)
				"0,1": {Q: 0, R: 1, TileType: 5}, // Can retreat here (BOTTOM_LEFT neighbor)
			},
			UnitsMap: map[string]*v1.Unit{
				"1,0": {
					Q:                1,
					R:                0,
					Player:           1,
					UnitType:         17, // Helicopter - action_order: ["move", "attack", "retreat"]
					AvailableHealth:  10,
					DistanceLeft:     2, // Has retreat points left
					ProgressionStep:  2, // At retreat step (after move=0, attack=1)
					LastToppedupTurn: 1,
					Shortcut:         "A5",
				},
				"2,0": {
					Q:               2,
					R:               0,
					Player:          2,
					UnitType:        17,
					AvailableHealth: 5, // Damaged from attack
					DistanceLeft:    5,
					Shortcut:        "B3",
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

	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: 1,
			R: 0,
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	moveOptionCount := 0
	attackOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetMove() != nil {
			moveOptionCount++
			t.Logf("Found move/retreat option: move to (%d,%d)", opt.GetMove().To.Q, opt.GetMove().To.R)
		}
		if opt.GetAttack() != nil {
			attackOptionCount++
			t.Logf("Found attack option: attack at (%d,%d)", opt.GetAttack().Defender.Q, opt.GetAttack().Defender.R)
		}
	}

	t.Logf("Move/Retreat options: %d, Attack options: %d, Total: %d",
		moveOptionCount, attackOptionCount, len(resp.Options))

	// After attacking (progression_step=2), should have NO attack options
	if attackOptionCount > 0 {
		t.Errorf("Expected 0 attack options after attacking (progression_step=2), got %d", attackOptionCount)
	}

	// Should have move options (retreat is shown as move options)
	if moveOptionCount == 0 {
		t.Errorf("Expected move/retreat options after attacking, got 0")
	}
}

// TestApplyUnitDamaged_ProgressionStepPersisted tests that when applyUnitDamaged
// is called (after an attack), the ProgressionStep is correctly persisted.
// This is the actual bug: applyUnitDamaged wasn't copying ProgressionStep.
func TestApplyUnitDamaged_ProgressionStepPersisted(t *testing.T) {
	game := &v1.Game{
		Id:   "test-game",
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Settings: &v1.GameSettings{},
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: TestUserID},
				{PlayerId: 2, UserId: "player-2"},
			},
		},
	}

	// Start with helicopter at ProgressionStep=1 (at attack step)
	gameState := &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 1000},
			2: {Coins: 1000},
		},
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				// Unit at (1,0) - neighbors are: (0,0), (1,-1), (2,-1), (2,0), (1,1), (0,1)
				"1,0": {Q: 1, R: 0, TileType: 5}, // Our helicopter
				"2,0": {Q: 2, R: 0, TileType: 5}, // Enemy here
				"0,0": {Q: 0, R: 0, TileType: 5}, // Can retreat here (LEFT neighbor)
				"0,1": {Q: 0, R: 1, TileType: 5}, // Can retreat here (BOTTOM_LEFT neighbor)
			},
			UnitsMap: map[string]*v1.Unit{
				"1,0": {
					Q:                1,
					R:                0,
					Player:           1,
					UnitType:         17, // Helicopter
					AvailableHealth:  10,
					DistanceLeft:     2,
					ProgressionStep:  1, // At attack step BEFORE attack
					LastToppedupTurn: 1,
					Shortcut:         "A5",
				},
				"2,0": {
					Q:               2,
					R:               0,
					Player:          2,
					UnitType:        17,
					AvailableHealth: 10,
					DistanceLeft:    5,
					Shortcut:        "B3",
				},
			},
		},
	}

	gamesService := singleton.NewSingletonGamesService()
	gamesService.SingletonGame = game
	gamesService.SingletonGameState = gameState
	gamesService.SingletonGameMoveHistory = &v1.GameMoveHistory{}
	gamesService.Self = gamesService

	ctx := AuthenticatedContext()

	// Verify we start at progression step 1
	unit := gameState.WorldData.UnitsMap["1,0"]
	if unit.ProgressionStep != 1 {
		t.Fatalf("Expected starting ProgressionStep=1, got %d", unit.ProgressionStep)
	}

	// Simulate processing an attack by calling ProcessMoves with an attack move
	// This will generate WorldChanges including UnitDamaged with updated ProgressionStep
	moveResp, err := gamesService.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: "test-game",
		Moves: []*v1.GameMove{
			{
				MoveType: &v1.GameMove_AttackUnit{
					AttackUnit: &v1.AttackUnitAction{
						Attacker: &v1.Position{
							Q: 1,
							R: 0,
						},
						Defender: &v1.Position{
							Q: 2,
							R: 0,
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("ProcessMoves failed: %v", err)
	}

	// The response should contain moves that were processed
	if len(moveResp.Moves) != 1 {
		t.Fatalf("Expected 1 move result, got %d", len(moveResp.Moves))
	}

	// Now check the unit's progression step - it should have advanced to 2
	resp, err := gamesService.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: "test-game",
		Pos: &v1.Position{
			Q: 1,
			R: 0,
		},
	})

	if err != nil {
		t.Fatalf("GetOptionsAt failed: %v", err)
	}

	attackOptionCount := 0
	moveOptionCount := 0
	for _, opt := range resp.Options {
		if opt.GetAttack() != nil {
			attackOptionCount++
		}
		if opt.GetMove() != nil {
			moveOptionCount++
		}
	}

	t.Logf("Attack options: %d, Move/Retreat options: %d", attackOptionCount, moveOptionCount)

	// After attacking, should have NO attack options (already used attack)
	if attackOptionCount > 0 {
		t.Errorf("BUG: After attacking, still showing %d attack options. ProgressionStep not persisted!", attackOptionCount)
	}

	// Should have move options (retreat is shown as move)
	if moveOptionCount == 0 {
		t.Errorf("Expected move/retreat options after attacking, got 0")
	}
}
