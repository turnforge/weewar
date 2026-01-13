package tests

import (
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// GameBuilder provides a fluent API for creating minimal test games
// without needing to specify full map files.
//
// Example usage:
//
//	game := NewGameBuilder().
//	    Tile(0, 0, TileTypeLandBase, 1).  // Player 1's base
//	    Tile(1, 0, TileTypeGrass, 0).     // Neutral grass
//	    Unit(0, 0, 1, UnitTypeSoldier).   // Player 1's soldier at base
//	    Coins(1, 500).                    // Player 1 has 500 coins
//	    Build()
type GameBuilder struct {
	tiles        []*tileSpec
	units        []*unitSpec
	playerCoins  map[int32]int32
	currentTurn  int32
	turnCounter  int32
	numPlayers   int32
	rngSeed      int64
	gameSettings *v1.GameSettings
}

type tileSpec struct {
	q, r     int
	tileType int32
	player   int32
}

type unitSpec struct {
	q, r            int
	player          int32
	unitType        int32
	shortcut        string
	health          int32
	distanceLeft    float64
	progressionStep int32
}

// NewGameBuilder creates a new game builder with sensible defaults
func NewGameBuilder() *GameBuilder {
	return &GameBuilder{
		tiles:       make([]*tileSpec, 0),
		units:       make([]*unitSpec, 0),
		playerCoins: make(map[int32]int32),
		currentTurn: 1,
		turnCounter: 1,
		numPlayers:  2,
		rngSeed:     12345,
	}
}

// Tile adds a tile at the specified position with given type and player ownership
// player=0 means neutral/unowned
func (b *GameBuilder) Tile(q, r int, tileType int32, player int32) *GameBuilder {
	b.tiles = append(b.tiles, &tileSpec{
		q:        q,
		r:        r,
		tileType: tileType,
		player:   player,
	})
	return b
}

// GrassTiles adds a grid of grass tiles centered at origin (for basic movement tests)
func (b *GameBuilder) GrassTiles(radius int) *GameBuilder {
	for q := -radius; q <= radius; q++ {
		for r := -radius; r <= radius; r++ {
			b.tiles = append(b.tiles, &tileSpec{
				q:        q,
				r:        r,
				tileType: lib.TileTypeGrass,
				player:   0,
			})
		}
	}
	return b
}

// Unit adds a unit at the specified position
func (b *GameBuilder) Unit(q, r int, player int32, unitType int32) *GameBuilder {
	b.units = append(b.units, &unitSpec{
		q:            q,
		r:            r,
		player:       player,
		unitType:     unitType,
		health:       10, // Default health
		distanceLeft: 3,  // Default movement
	})
	return b
}

// UnitWithShortcut adds a unit with a specific shortcut identifier
func (b *GameBuilder) UnitWithShortcut(q, r int, player int32, unitType int32, shortcut string) *GameBuilder {
	b.units = append(b.units, &unitSpec{
		q:            q,
		r:            r,
		player:       player,
		unitType:     unitType,
		shortcut:     shortcut,
		health:       10,
		distanceLeft: 3,
	})
	return b
}

// UnitFull adds a unit with full control over all attributes
func (b *GameBuilder) UnitFull(q, r int, player int32, unitType int32, shortcut string, health int32, distanceLeft float64, progressionStep int32) *GameBuilder {
	b.units = append(b.units, &unitSpec{
		q:               q,
		r:               r,
		player:          player,
		unitType:        unitType,
		shortcut:        shortcut,
		health:          health,
		distanceLeft:    distanceLeft,
		progressionStep: progressionStep,
	})
	return b
}

// Coins sets the coin balance for a player
func (b *GameBuilder) Coins(player int32, amount int32) *GameBuilder {
	b.playerCoins[player] = amount
	return b
}

// CurrentPlayer sets which player's turn it is
func (b *GameBuilder) CurrentPlayer(player int32) *GameBuilder {
	b.currentTurn = player
	return b
}

// Turn sets the turn counter
func (b *GameBuilder) Turn(turn int32) *GameBuilder {
	b.turnCounter = turn
	return b
}

// Players sets the number of players (default 2)
func (b *GameBuilder) Players(n int32) *GameBuilder {
	b.numPlayers = n
	return b
}

// Seed sets the RNG seed for deterministic combat
func (b *GameBuilder) Seed(seed int64) *GameBuilder {
	b.rngSeed = seed
	return b
}

// Settings sets game settings (for allowed units, etc.)
func (b *GameBuilder) Settings(settings *v1.GameSettings) *GameBuilder {
	b.gameSettings = settings
	return b
}

// Build constructs the Game from the builder configuration
func (b *GameBuilder) Build() *lib.Game {
	// Build tiles map
	tilesMap := make(map[string]*v1.Tile)
	for _, ts := range b.tiles {
		key := lib.CoordKey(int32(ts.q), int32(ts.r))
		tilesMap[key] = &v1.Tile{
			Q:        int32(ts.q),
			R:        int32(ts.r),
			TileType: ts.tileType,
			Player:   ts.player,
		}
	}

	// Build units map with auto-generated shortcuts
	unitsMap := make(map[string]*v1.Unit)
	shortcutCounters := make(map[int32]int) // per-player counters
	for _, us := range b.units {
		shortcut := us.shortcut
		if shortcut == "" {
			// Auto-generate shortcut like "A1", "A2", "B1", etc.
			letter := 'A' + rune(us.player-1)
			if us.player <= 0 {
				letter = 'N' // Neutral
			}
			shortcutCounters[us.player]++
			shortcut = string(letter) + string('0'+rune(shortcutCounters[us.player]))
		}

		key := lib.CoordKey(int32(us.q), int32(us.r))
		unitsMap[key] = &v1.Unit{
			Q:               int32(us.q),
			R:               int32(us.r),
			Player:          us.player,
			UnitType:        us.unitType,
			Shortcut:        shortcut,
			AvailableHealth: us.health,
			DistanceLeft:    us.distanceLeft,
			ProgressionStep: us.progressionStep,
		}
	}

	worldData := &v1.WorldData{
		TilesMap: tilesMap,
		UnitsMap: unitsMap,
	}

	// Build player states
	playerStates := make(map[int32]*v1.PlayerState)
	for i := int32(1); i <= b.numPlayers; i++ {
		coins := b.playerCoins[i]
		if coins == 0 {
			coins = 300 // Default starting coins
		}
		playerStates[i] = &v1.PlayerState{
			Coins:    coins,
			IsActive: true,
		}
	}

	// Build game configuration
	players := make([]*v1.GamePlayer, 0, b.numPlayers)
	for i := int32(1); i <= b.numPlayers; i++ {
		players = append(players, &v1.GamePlayer{
			PlayerId:      i,
			StartingCoins: b.playerCoins[i],
		})
	}

	settings := b.gameSettings
	if settings == nil {
		settings = &v1.GameSettings{}
	}

	game := &v1.Game{
		Id:      "test-game",
		WorldId: "test-world",
		Config: &v1.GameConfiguration{
			Players:  players,
			Settings: settings,
		},
	}

	state := &v1.GameState{
		GameId:        "test-game",
		CurrentPlayer: b.currentTurn,
		TurnCounter:   b.turnCounter,
		WorldData:     worldData,
		PlayerStates:  playerStates,
	}

	rulesEngine := lib.DefaultRulesEngine()
	return lib.NewGame(game, state, lib.NewWorld("test-world", worldData), rulesEngine, b.rngSeed)
}

// Common unit type constants for convenience
const (
	// Land units
	UnitTypeSoldierBasic    int32 = 1  // Can capture, Light:Land
	UnitTypeSoldierAdvanced int32 = 2  // Can capture, Light:Land
	UnitTypeStriker         int32 = 3  // Light:Land, no capture
	UnitTypeTank            int32 = 5  // Heavy:Land
	UnitTypeArtillery       int32 = 7  // Heavy:Land, ranged
	UnitTypeArtilleryMega   int32 = 8  // Heavy:Land, ranged, splash
	UnitTypeAntiAir         int32 = 9  // Heavy:Land
	UnitTypeRocketLauncher  int32 = 10 // Heavy:Land, ranged

	// Naval units
	UnitTypeSpeedboat   int32 = 11 // Light:Water
	UnitTypeDestroyer   int32 = 12 // Heavy:Water
	UnitTypeBattleship  int32 = 13 // Heavy:Water, ranged
	UnitTypeSubmarine   int32 = 14 // Stealth:Water
	UnitTypeHovercraft  int32 = 15 // Light:Water, can capture

	// Air units
	UnitTypeHelicopter     int32 = 16 // Light:Air
	UnitTypeFighter        int32 = 17 // Light:Air
	UnitTypeBomber         int32 = 18 // Heavy:Air
	UnitTypeZeppelin       int32 = 19 // Light:Air
	UnitTypeJetFighter     int32 = 21 // Light:Air
	UnitTypeHeavyBomber    int32 = 22 // Heavy:Air, splash
	UnitTypeStealthBomber  int32 = 23 // Stealth:Air
	UnitTypeStealthFighter int32 = 24 // Stealth:Air

	// Support units (with fix ability)
	UnitTypeMedic          int32 = 27 // Can fix
	UnitTypeStratotanker   int32 = 28 // Can fix air units
	UnitTypeEngineer       int32 = 29 // Can fix and capture
	UnitTypeTugboat        int32 = 31 // Can fix naval
	UnitTypeAircraftCarrier int32 = 39 // Can fix air units
)

// Additional tile type constants
const (
	TileTypePlains       int32 = 5
	TileTypeWaterShallow int32 = 14
	TileTypeWaterRegular int32 = 10
	TileTypeWaterDeep    int32 = 15
	TileTypeRoad         int32 = 22
)
