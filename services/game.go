package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Core Types (from core.go)
// =============================================================================

// Game represents the unified game state and implements GameInterface
type Game struct {
	*v1.Game
	*v1.GameState
	*v1.GameMoveHistory

	// Pure game state - this is a "view" over the GameState so we can do a lot more native ops on this
	World *World `json:"world"` // Contains the pure state (map, units, entities)

	// Game systems and configuration
	Seed int64 `json:"seed"` // Random seed for deterministic gameplay

	// Random number generator
	rng *rand.Rand `json:"-"` // RNG for deterministic gameplay

	// Rules engine for data-driven game mechanics
	rulesEngine *RulesEngine `json:"-"` // Rules engine for movement costs, combat, unit data
}

// NewGame creates a new game instance with the specified parameters
func NewGame(game *v1.Game, state *v1.GameState, world *World, rulesEngine *RulesEngine, seed int64) *Game {
	// Validate parameters
	if rulesEngine == nil {
		panic("rules engine is required")
	}

	// Create the game struct
	out := &Game{
		Game:        game,
		GameState:   state,
		World:       world,
		Seed:        seed,
		rng:         rand.New(rand.NewSource(seed)),
		rulesEngine: rulesEngine,
	}

	// Debug: Check unit movement points after NewGame initialization
	if out.World != nil {
		for playerId := 1; playerId <= int(out.World.PlayerCount()); playerId++ {
			units := out.World.GetPlayerUnits(playerId)
			fmt.Printf("NewGame: After NewGame - Player %d has %d units\n", playerId, len(units))
			for _, unit := range units {
				fmt.Printf("NewGame: Player %d unit at (%d, %d) - DistanceLeft=%f, AvailableHealth=%d\n",
					playerId, unit.Q, unit.R, unit.DistanceLeft, unit.AvailableHealth)
			}
		}
	}
	return out
}

// =============================================================================
// Convenience methods to access World fields
// =============================================================================

// GetGamePlayer returns information about a specific player
func (g *Game) GetGamePlayer(playerID int) (*v1.GamePlayer, error) {
	if playerID < 0 || playerID >= len(g.Game.Config.Players) {
		return nil, fmt.Errorf("invalid player ID: %d", playerID)
	}
	return g.Game.Config.Players[playerID], nil
}

// GetGameTeam returns information about a specific team
func (g *Game) GetGameTeam(teamID int) (*v1.GameTeam, error) {
	if teamID < 0 || teamID >= len(g.Game.Config.Teams) {
		return nil, fmt.Errorf("invalid team ID: %d", teamID)
	}
	return g.Game.Config.Teams[teamID], nil
}

// GetPlayersOnTeam returns all players belonging to a specific team
func (g *Game) GetPlayersOnTeam(teamID int32) []*v1.GamePlayer {
	var teamPlayers []*v1.GamePlayer
	for i := range g.Game.Config.Players {
		if g.Game.Config.Players[i].TeamId == teamID {
			teamPlayers = append(teamPlayers, g.Game.Config.Players[i])
		}
	}
	return teamPlayers
}

// ArePlayersOnSameTeam checks if two players are on the same team
func (g *Game) ArePlayersOnSameTeam(playerID1, playerID2 int) bool {
	if playerID1 < 0 || playerID1 >= len(g.Game.Config.Players) ||
		playerID2 < 0 || playerID2 >= len(g.Game.Config.Players) {
		return false
	}
	return g.Game.Config.Players[playerID1].TeamId == g.Game.Config.Players[playerID2].TeamId
}

// =============================================================================
// Helper Functions
// =============================================================================

// topUpUnitIfNeeded performs lazy top-up of unit stats if the unit hasn't been refreshed this turn
// This checks if unit.LastToppedupTurn < game.TurnCounter and if so:
// - Restores movement points to max
// - Sets available health to max (for new units) or applies healing
// - Updates unit.LastToppedupTurn to game.TurnCounter
func (g *Game) topUpUnitIfNeeded(unit *v1.Unit) error {
	// Check if unit needs top-up (hasn't been refreshed this turn)
	if unit.LastToppedupTurn >= g.TurnCounter {
		return nil // Already topped up this turn
	}

	// Get unit definition from rules engine
	if g.rulesEngine == nil {
		return fmt.Errorf("rules engine not set")
	}

	unitData, err := g.rulesEngine.GetUnitData(unit.UnitType)
	if err != nil {
		return fmt.Errorf("failed to get unit data for type %d: %w", unit.UnitType, err)
	}

	// Top-up movement points
	unit.DistanceLeft = unitData.MovementPoints

	// Top-up health (for new units or apply healing)
	if unit.AvailableHealth == 0 {
		// New unit - set to max health
		unit.AvailableHealth = unitData.Health
	} else {
		// Existing unit - apply healing from terrain (TODO: implement terrain healing)
		// For now, keep current health
	}

	// Mark unit as topped-up for this turn
	unit.LastToppedupTurn = g.TurnCounter

	return nil
}

// checkVictoryConditions checks if any player has won
func (g *Game) checkVictoryConditions() (winner int32, hasWinner bool) {
	// Simple victory condition: last player with units wins
	playersWithUnits := 0
	lastPlayerWithUnits := int32(-1)

	for playerID := int32(1); playerID <= g.World.PlayerCount(); playerID++ {
		units := g.World.GetPlayerUnits(int(playerID))
		if len(units) > 0 {
			playersWithUnits++
			lastPlayerWithUnits = playerID
		}
	}

	if playersWithUnits == 1 {
		return lastPlayerWithUnits, true
	}

	return -1, false
}

// validateGameState validates the current game state
func (g *Game) validateGameState() error {
	if g.World == nil {
		return fmt.Errorf("game has no world")
	}

	if g.GameState.CurrentPlayer < 0 || g.CurrentPlayer > g.World.PlayerCount() {
		return fmt.Errorf("invalid current player: %d", g.CurrentPlayer)
	}

	if g.TurnCounter < 1 {
		return fmt.Errorf("invalid turn counter: %d", g.TurnCounter)
	}

	if int32(len(g.World.unitsByPlayer)) != g.World.PlayerCount() {
		return fmt.Errorf("units array length (%d) doesn't match player count (%d)", len(g.World.unitsByPlayer), g.World.PlayerCount())
	}

	return nil
}

// GetUnitID generates a unique identifier for a unit in the format PN
// where P is the player letter (A-Z) and N is the unit number for that player
func (g *Game) GetUnitID(unit *v1.Unit) string {
	if unit == nil {
		return ""
	}

	// This method was only used for cli - we can come back to this when needed
	panic("to be deprecated")
	// return unit.unitID
}

// GetRulesEngine returns the current RulesEngine instance
func (g *Game) GetRulesEngine() *RulesEngine {
	return g.rulesEngine
}

// SetRulesEngine sets the RulesEngine instance for data-driven game mechanics
func (g *Game) SetRulesEngine(rulesEngine *RulesEngine) {
	g.rulesEngine = rulesEngine
}

// LoadGame restores a game from saved JSON data
func LoadGame(saveData []byte) (*Game, error) {
	var game Game
	if err := json.Unmarshal(saveData, &game); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	// Restore transient state
	game.rng = rand.New(rand.NewSource(game.Seed))
	game.rulesEngine = nil // Will be set by caller

	// Note: Neighbor connections are no longer stored, calculated on-demand

	// Validate loaded game state
	if err := game.validateGameState(); err != nil {
		return nil, fmt.Errorf("invalid saved game state: %w", err)
	}

	return &game, nil
}

// =============================================================================
// GameController Interface Implementation
// =============================================================================

// LoadGame restores game from saved state (interface method)
func (g *Game) LoadGame(saveData []byte) (*Game, error) {
	return LoadGame(saveData)
}

// SaveGame serializes current game state
func (g *Game) SaveGame() ([]byte, error) {
	// Update last action time
	g.Game.UpdatedAt = tspb.New(time.Now())

	// Serialize to JSON
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize game state: %w", err)
	}

	return data, nil
}

// GetCurrentPlayer returns active player ID
func (g *Game) GetCurrentPlayer() int32 {
	return g.CurrentPlayer
}

// GetTurnNumber returns current turn count
func (g *Game) GetTurnNumber() int32 {
	return g.TurnCounter
}

// =============================================================================
// UnitInterface Interface Implementation
// =============================================================================

// GetUnitsForPlayer returns all units owned by player
func (g *Game) GetUnitsForPlayer(playerID int) []*v1.Unit {
	if playerID < 0 || playerID >= len(g.World.unitsByPlayer) {
		return nil
	}

	// Return a copy to prevent external modification
	units := make([]*v1.Unit, len(g.World.unitsByPlayer[playerID]))
	copy(units, g.World.unitsByPlayer[playerID])
	return units
}
