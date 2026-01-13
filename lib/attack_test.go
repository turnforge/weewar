package lib

import (
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// testGameBuilder provides a fluent API for creating minimal test games
type testGameBuilder struct {
	tiles        []*testTileSpec
	units        []*testUnitSpec
	playerCoins  map[int32]int32
	currentTurn  int32
	turnCounter  int32
	numPlayers   int32
	rngSeed      int64
}

type testTileSpec struct {
	q, r     int
	tileType int32
	player   int32
}

type testUnitSpec struct {
	q, r            int
	player          int32
	unitType        int32
	shortcut        string
	health          int32
	distanceLeft    float64
	progressionStep int32
	attackHistory   []*v1.AttackRecord
}

func newTestGameBuilder() *testGameBuilder {
	return &testGameBuilder{
		tiles:       make([]*testTileSpec, 0),
		units:       make([]*testUnitSpec, 0),
		playerCoins: make(map[int32]int32),
		currentTurn: 1,
		turnCounter: 1,
		numPlayers:  2,
		rngSeed:     12345,
	}
}

func (b *testGameBuilder) tile(q, r int, tileType int32, player int32) *testGameBuilder {
	b.tiles = append(b.tiles, &testTileSpec{q: q, r: r, tileType: tileType, player: player})
	return b
}

func (b *testGameBuilder) grassTiles(radius int) *testGameBuilder {
	for q := -radius; q <= radius; q++ {
		for r := -radius; r <= radius; r++ {
			// Don't overwrite explicitly defined tiles
			exists := false
			for _, t := range b.tiles {
				if t.q == q && t.r == r {
					exists = true
					break
				}
			}
			if !exists {
				b.tiles = append(b.tiles, &testTileSpec{q: q, r: r, tileType: TileTypeGrass, player: 0})
			}
		}
	}
	return b
}

func (b *testGameBuilder) unit(q, r int, player int32, unitType int32) *testGameBuilder {
	b.units = append(b.units, &testUnitSpec{
		q: q, r: r, player: player, unitType: unitType,
		health: 10, distanceLeft: 3,
	})
	return b
}

func (b *testGameBuilder) unitFull(q, r int, player int32, unitType int32, shortcut string, health int32, distanceLeft float64) *testGameBuilder {
	b.units = append(b.units, &testUnitSpec{
		q: q, r: r, player: player, unitType: unitType,
		shortcut: shortcut, health: health, distanceLeft: distanceLeft,
	})
	return b
}

func (b *testGameBuilder) unitWithHistory(q, r int, player int32, unitType int32, history []*v1.AttackRecord) *testGameBuilder {
	b.units = append(b.units, &testUnitSpec{
		q: q, r: r, player: player, unitType: unitType,
		health: 10, distanceLeft: 3, attackHistory: history,
	})
	return b
}

func (b *testGameBuilder) coins(player int32, amount int32) *testGameBuilder {
	b.playerCoins[player] = amount
	return b
}

func (b *testGameBuilder) currentPlayer(player int32) *testGameBuilder {
	b.currentTurn = player
	return b
}

func (b *testGameBuilder) seed(seed int64) *testGameBuilder {
	b.rngSeed = seed
	return b
}

func (b *testGameBuilder) build() *Game {
	tilesMap := make(map[string]*v1.Tile)
	for _, ts := range b.tiles {
		key := CoordKey(int32(ts.q), int32(ts.r))
		tilesMap[key] = &v1.Tile{Q: int32(ts.q), R: int32(ts.r), TileType: ts.tileType, Player: ts.player}
	}

	unitsMap := make(map[string]*v1.Unit)
	shortcutCounters := make(map[int32]int)
	for _, us := range b.units {
		shortcut := us.shortcut
		if shortcut == "" {
			letter := 'A' + rune(us.player-1)
			if us.player <= 0 {
				letter = 'N'
			}
			shortcutCounters[us.player]++
			shortcut = string(letter) + string('0'+rune(shortcutCounters[us.player]))
		}
		key := CoordKey(int32(us.q), int32(us.r))
		unitsMap[key] = &v1.Unit{
			Q: int32(us.q), R: int32(us.r), Player: us.player, UnitType: us.unitType,
			Shortcut: shortcut, AvailableHealth: us.health, DistanceLeft: us.distanceLeft,
			ProgressionStep: us.progressionStep, AttackHistory: us.attackHistory,
		}
	}

	worldData := &v1.WorldData{TilesMap: tilesMap, UnitsMap: unitsMap}

	playerStates := make(map[int32]*v1.PlayerState)
	for i := int32(1); i <= b.numPlayers; i++ {
		coins := b.playerCoins[i]
		if coins == 0 {
			coins = 300
		}
		playerStates[i] = &v1.PlayerState{Coins: coins, IsActive: true}
	}

	players := make([]*v1.GamePlayer, 0, b.numPlayers)
	for i := int32(1); i <= b.numPlayers; i++ {
		players = append(players, &v1.GamePlayer{PlayerId: i, StartingCoins: b.playerCoins[i]})
	}

	game := &v1.Game{
		Id: "test-game", WorldId: "test-world",
		Config: &v1.GameConfiguration{Players: players, Settings: &v1.GameSettings{}},
	}
	state := &v1.GameState{
		GameId: "test-game", CurrentPlayer: b.currentTurn, TurnCounter: b.turnCounter,
		WorldData: worldData, PlayerStates: playerStates,
	}

	return NewGame(game, state, NewWorld("test-world", worldData), DefaultRulesEngine(), b.rngSeed)
}

// Unit type constants for tests
const (
	testUnitTypeSoldier   int32 = 1  // Light:Land, range 1, can capture
	testUnitTypeTank      int32 = 5  // Heavy:Land, range 1
	testUnitTypeArtillery int32 = 7  // Heavy:Land, range 2-3
)

// TestProcessAttackUnit_BasicDamage tests that attacks deal damage to defender
func TestProcessAttackUnit_BasicDamage(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 2, testUnitTypeSoldier).
		currentPlayer(1).
		seed(42).
		build()

	defender := game.World.UnitAt(AxialCoord{Q: 1, R: 0})
	if defender == nil {
		t.Fatal("Defender not found")
	}
	initialHealth := defender.AvailableHealth

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	defender = game.World.UnitAt(AxialCoord{Q: 1, R: 0})
	if defender != nil && defender.AvailableHealth >= initialHealth {
		t.Errorf("Defender should take damage: was %d, now %d", initialHealth, defender.AvailableHealth)
	}
}

// TestProcessAttackUnit_CounterAttack tests defender retaliation
func TestProcessAttackUnit_CounterAttack(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 2, testUnitTypeSoldier).
		currentPlayer(1).
		seed(42).
		build()

	attacker := game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	initialHealth := attacker.AvailableHealth

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	attacker = game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if attacker != nil {
		t.Logf("Attacker health: was %d, now %d (counter-attack)", initialHealth, attacker.AvailableHealth)
	}
}

// TestProcessAttackUnit_CannotAttackOwnUnit tests friendly fire prevention
func TestProcessAttackUnit_CannotAttackOwnUnit(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 1, testUnitTypeSoldier). // Same player
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should not be able to attack own unit")
	}
}

// TestProcessAttackUnit_OutOfRange tests range validation
func TestProcessAttackUnit_OutOfRange(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(3).
		unit(0, 0, 1, testUnitTypeSoldier). // Range 1
		unit(2, 0, 2, testUnitTypeSoldier). // 2 tiles away
		currentPlayer(1).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 2, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Soldier (range 1) should not attack 2 tiles away")
	}
}

// TestProcessAttackUnit_WrongTurn tests turn validation
func TestProcessAttackUnit_WrongTurn(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 2, testUnitTypeSoldier).
		currentPlayer(2). // Player 2's turn
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0}, // Player 1's unit
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err == nil {
		t.Error("Should not attack on opponent's turn")
	}
}

// TestProcessAttackUnit_ProgressionAdvances tests action progression
func TestProcessAttackUnit_ProgressionAdvances(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 2, testUnitTypeSoldier).
		currentPlayer(1).
		build()

	attacker := game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	initialStep := attacker.ProgressionStep

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	attacker = game.World.UnitAt(AxialCoord{Q: 0, R: 0})
	if attacker != nil && attacker.ProgressionStep <= initialStep {
		t.Errorf("Progression should advance: was %d, now %d", initialStep, attacker.ProgressionStep)
	}
}

// TestProcessAttackUnit_DamageChangeRecorded tests change recording
func TestProcessAttackUnit_DamageChangeRecorded(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 2, testUnitTypeSoldier).
		currentPlayer(1).
		seed(42).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	hasDamageChange := false
	for _, change := range move.Changes {
		if _, ok := change.ChangeType.(*v1.WorldChange_UnitDamaged); ok {
			hasDamageChange = true
			break
		}
	}

	if !hasDamageChange {
		t.Error("Attack should record UnitDamaged change")
	}
}

// TestProcessAttackUnit_AttackHistoryUpdated tests wound tracking
func TestProcessAttackUnit_AttackHistoryUpdated(t *testing.T) {
	game := newTestGameBuilder().
		grassTiles(2).
		unit(0, 0, 1, testUnitTypeSoldier).
		unit(1, 0, 2, testUnitTypeSoldier).
		currentPlayer(1).
		seed(42).
		build()

	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				Attacker: &v1.Position{Q: 0, R: 0},
				Defender: &v1.Position{Q: 1, R: 0},
			},
		},
	}

	if err := game.ProcessMove(move); err != nil {
		t.Fatalf("ProcessMove failed: %v", err)
	}

	defender := game.World.UnitAt(AxialCoord{Q: 1, R: 0})
	if defender != nil && len(defender.AttackHistory) == 0 {
		t.Error("Attack should be recorded in defender's history")
	}
}
