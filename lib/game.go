package lib

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
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
	RulesEngine *RulesEngine `json:"-"` // Rules engine for movement costs, combat, unit data
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
		RulesEngine: rulesEngine,
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

// NumPlayers returns the number of players configured for this game.
// Uses the game configuration as the source of truth.
func (g *Game) NumPlayers() int32 {
	if g.Config == nil || g.Config.Players == nil {
		return 0
	}
	return int32(len(g.Config.Players))
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

// TopUpTileIfNeeded performs lazy top-up of tile stats if the tile hasn't been refreshed this turn
// This checks if tile.LastToppedupTurn < game.TurnCounter and if so:
// - Restores movement points to max
// - Sets available health to max (for new tiles) or applies healing
// - Clears attack history and attacks received counter (wound bonus resets each turn)
// - Updates tile.LastToppedupTurn to game.TurnCounter
func (g *Game) TopUpTileIfNeeded(tile *v1.Tile) error {
	// Check if tile needs top-up (hasn't been refreshed this turn)
	if tile.LastToppedupTurn >= g.TurnCounter {
		return nil // Already topped up this turn
	}

	// Get tile definition from rules engine
	if g.RulesEngine == nil {
		return fmt.Errorf("rules engine not set")
	}

	/* - TODO - use when needed
	tileData, err := g.RulesEngine.GetTerrainData(tile.TileType)
	if err != nil {
		return fmt.Errorf("failed to get tile data for type %d: %w", tile.TileType, err)
	}
	*/

	// Mark tile as topped-up for this turn
	tile.LastToppedupTurn = g.TurnCounter

	return nil
}

// TopUpUnitIfNeeded performs lazy top-up of unit stats if the unit hasn't been refreshed this turn
// This checks if unit.LastToppedupTurn < game.TurnCounter and if so:
// - Restores movement points to max
// - Sets available health to max (for new units) or applies healing
// - Clears attack history and attacks received counter (wound bonus resets each turn)
// - Updates unit.LastToppedupTurn to game.TurnCounter
func (g *Game) TopUpUnitIfNeeded(unit *v1.Unit) error {
	// Check if unit needs top-up (hasn't been refreshed this turn)
	if unit.LastToppedupTurn >= g.TurnCounter {
		return nil // Already topped up this turn
	}

	// Get unit definition from rules engine
	if g.RulesEngine == nil {
		return fmt.Errorf("rules engine not set")
	}

	unitData, err := g.RulesEngine.GetUnitData(unit.UnitType)
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
		// Existing unit - apply healing from terrain if eligible
		// Healing rules:
		// 1. Unit must not have acted last turn (unused units heal)
		// 2. Unit cannot heal on enemy-owned bases
		// 3. Air units can only heal on Airport Bases
		healAmount := g.calculateHealAmount(unit, unitData)
		if healAmount > 0 {
			newHealth := unit.AvailableHealth + healAmount
			if newHealth > unitData.Health {
				newHealth = unitData.Health // Cap at max health
			}
			unit.AvailableHealth = newHealth
		}
	}

	// Clear attack history and attacks received counter for new turn
	// Wound bonus only accumulates within a single turn
	unit.AttackHistory = nil
	unit.AttacksReceivedThisTurn = 0

	// Reset action progression for new turn
	unit.ProgressionStep = 0
	unit.ChosenAlternative = ""

	// Check for pending capture completion
	// If unit started capturing in a previous turn and survived, complete the capture
	if unit.CaptureStartedTurn > 0 && unit.CaptureStartedTurn < g.TurnCounter {
		coord := AxialCoord{Q: int(unit.Q), R: int(unit.R)}
		tile := g.World.TileAt(coord)
		if tile != nil && tile.Player != unit.Player {
			// Complete the capture - transfer ownership
			tile.Player = unit.Player
			fmt.Printf("Capture completed: tile at (%d,%d) now belongs to player %d\n",
				tile.Q, tile.R, unit.Player)
		}
		// Clear capture state
		unit.CaptureStartedTurn = 0
	}

	// Mark unit as topped-up for this turn
	unit.LastToppedupTurn = g.TurnCounter

	return nil
}

// calculateHealAmount determines how much healing a unit receives based on terrain and restrictions.
// Returns 0 if the unit is not eligible for healing.
func (g *Game) calculateHealAmount(unit *v1.Unit, unitData *v1.UnitDefinition) int32 {
	// Rule 1: Unit must not have acted last turn (only unused units heal)
	// LastActedTurn tracks when the unit last moved/attacked
	// If it acted in the previous turn (TurnCounter-1), it doesn't heal
	previousTurn := g.TurnCounter - 1
	if previousTurn < 1 {
		previousTurn = 1 // First turn edge case
	}
	if unit.LastActedTurn >= previousTurn {
		return 0 // Unit was used last turn, no healing
	}

	// Get the tile the unit is standing on
	coord := AxialCoord{Q: int(unit.Q), R: int(unit.R)}
	tile := g.World.TileAt(coord)
	if tile == nil {
		return 0 // No tile, no healing
	}

	// Rule 2: Cannot heal on enemy-owned bases
	// If tile is owned by another player (not neutral and not the unit's player)
	if tile.Player != 0 && tile.Player != unit.Player {
		return 0 // Enemy base, no healing
	}

	// Rule 3: Air units can only heal on Airport Bases
	// Check unit terrain type (Air, Land, Water)
	if unitData.UnitTerrain == "Air" {
		// Get terrain definition to check if it's an airport
		terrainData, err := g.RulesEngine.GetTerrainData(tile.TileType)
		if err != nil || terrainData == nil {
			return 0
		}
		// Check if terrain name contains "Airport" (Airport Base, etc.)
		if terrainData.Name != "Airport Base" {
			return 0 // Air units can only heal on Airport Base
		}
	}

	// Look up healing bonus from TerrainUnitProperties
	terrainProps := g.RulesEngine.GetTerrainUnitPropertiesForUnit(tile.TileType, unit.UnitType)
	if terrainProps == nil || terrainProps.HealingBonus <= 0 {
		return 0 // No healing available on this terrain for this unit
	}

	return terrainProps.HealingBonus
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
	return g.RulesEngine
}

// SetRulesEngine sets the RulesEngine instance for data-driven game mechanics
func (g *Game) SetRulesEngine(rulesEngine *RulesEngine) {
	g.RulesEngine = rulesEngine
}

// LoadGame restores a game from saved JSON data
func LoadGame(saveData []byte) (*Game, error) {
	var game Game
	if err := json.Unmarshal(saveData, &game); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	// Restore transient state
	game.rng = rand.New(rand.NewSource(game.Seed))
	game.RulesEngine = nil // Will be set by caller

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
// Position Parsing Methods
// =============================================================================

// FromPos converts a Position proto to an AxialCoord.
// If pos.Label is set, it parses the label (supports "A1", "3,4", "r4,5", etc.)
// Otherwise, it uses pos.Q and pos.R directly.
func (g *Game) FromPos(pos *v1.Position) (AxialCoord, error) {
	return g.FromPosWithBase(pos, nil)
}

// FromPosWithBase converts a Position proto to an AxialCoord with an optional base coordinate.
// The base coordinate enables relative directions like "L", "TR", "TL,TL,R".
// When parsing from a label, the Q/R fields on the Position are populated with the resolved coordinates.
func (g *Game) FromPosWithBase(pos *v1.Position, base *AxialCoord) (AxialCoord, error) {
	if pos == nil {
		return AxialCoord{}, fmt.Errorf("position is nil")
	}

	// If label is empty, use Q/R directly (this is the expected case for pre-resolved positions)
	if pos.Label == "" {
		return AxialCoord{Q: int(pos.Q), R: int(pos.R)}, nil
	}

	// If Q/R are already set to non-origin values, use them directly
	// This handles positions that were already resolved by ParseTarget.Position()
	if pos.Q != 0 || pos.R != 0 {
		return AxialCoord{Q: int(pos.Q), R: int(pos.R)}, nil
	}

	// Label is set and Q/R are both 0 - need to parse the label
	// This handles both origin (0,0) positions and positions that need parsing
	target, err := ParsePositionOrUnitWithContext(g, pos.Label, base)
	if err != nil {
		return AxialCoord{}, fmt.Errorf("failed to parse position label %q: %w", pos.Label, err)
	}
	// Populate Q/R on the Position proto for later use
	pos.Q = int32(target.Coordinate.Q)
	pos.R = int32(target.Coordinate.R)
	return target.Coordinate, nil
}

// Pos parses a position string and returns a ParseTarget.
// Supports: "A1" (unit), "3,4" (Q,R), "r4,5" (row,col), "L"/"TR" (direction), "t:A1" (tile)
// Optional second argument provides base coordinate for relative directions.
func (g *Game) Pos(input string, from ...string) (*ParseTarget, error) {
	var baseCoord *AxialCoord
	if len(from) > 0 {
		baseTarget, err := ParsePositionOrUnit(g, from[0])
		if err != nil {
			return nil, fmt.Errorf("invalid base position %q: %w", from[0], err)
		}
		coord := baseTarget.Coordinate
		baseCoord = &coord
	}
	return ParsePositionOrUnitWithContext(g, input, baseCoord)
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

// IsUnitExhausted returns true if a unit should be shown as exhausted.
// A unit is exhausted only if:
// 1. It has been topped up this turn (LastToppedupTurn >= TurnCounter)
// 2. AND it has no movement left (DistanceLeft <= 0)
// If LastToppedupTurn < TurnCounter, the unit will be topped up when accessed (lazy pattern).
func (g *Game) IsUnitExhausted(unit *v1.Unit) bool {
	return unit.LastToppedupTurn >= g.TurnCounter && unit.DistanceLeft <= 0
}

// GetExhaustedUnits returns all units for the current player that are exhausted.
// Uses the lazy top-up pattern: units not yet topped up this turn are NOT considered exhausted.
func (g *Game) GetExhaustedUnits() []*v1.Unit {
	var exhausted []*v1.Unit
	for _, unit := range g.World.UnitsByCoord() {
		if unit.Player == g.CurrentPlayer && g.IsUnitExhausted(unit) {
			exhausted = append(exhausted, unit)
		}
	}
	return exhausted
}

// =============================================================================
// Controller Methods - High-level game actions
// =============================================================================

// Move moves a unit to target position.
// unit: position string for the unit ("A1", "3,4", etc.)
// target: target string (can be relative like "R", "TL,TR", or absolute)
// Returns world changes from the move.
func (g *Game) Move(unit, target string) ([]*v1.WorldChange, error) {
	// Parse unit position
	src, err := g.Pos(unit)
	if err != nil {
		return nil, fmt.Errorf("invalid unit position %q: %w", unit, err)
	}
	if src.Unit == nil {
		return nil, fmt.Errorf("no unit at position %q", unit)
	}

	// Parse target position relative to unit
	dest, err := g.Pos(target, unit)
	if err != nil {
		return nil, fmt.Errorf("invalid target position %q: %w", target, err)
	}

	// Create move action
	action := &v1.MoveUnitAction{
		From: src.Position(),
		To:   dest.Position(),
	}

	move := &v1.GameMove{
		Player:   g.CurrentPlayer,
		MoveType: &v1.GameMove_MoveUnit{MoveUnit: action},
	}

	// Process move
	if err := g.ProcessMoveUnit(move, action, false); err != nil {
		return nil, err
	}

	return move.Changes, nil
}

// Attack attacks target from attacker position.
// attacker: position string for the attacking unit
// defender: position string for the target (can be relative like "R", "TL")
// Returns world changes from the attack.
func (g *Game) Attack(attacker, defender string) ([]*v1.WorldChange, error) {
	// Parse attacker position
	src, err := g.Pos(attacker)
	if err != nil {
		return nil, fmt.Errorf("invalid attacker position %q: %w", attacker, err)
	}
	if src.Unit == nil {
		return nil, fmt.Errorf("no unit at attacker position %q", attacker)
	}

	// Parse defender position relative to attacker
	dest, err := g.Pos(defender, attacker)
	if err != nil {
		return nil, fmt.Errorf("invalid defender position %q: %w", defender, err)
	}

	// Create attack action
	action := &v1.AttackUnitAction{
		Attacker: src.Position(),
		Defender: dest.Position(),
	}

	move := &v1.GameMove{
		Player:   g.CurrentPlayer,
		MoveType: &v1.GameMove_AttackUnit{AttackUnit: action},
	}

	// Process attack
	if err := g.ProcessAttackUnit(move, action); err != nil {
		return nil, err
	}

	return move.Changes, nil
}

// Build creates a unit at tile position.
// tile: position string for the building tile ("t:A1", "3,4", etc.)
// unitType: the type of unit to build
// Returns world changes from the build.
func (g *Game) Build(tile string, unitType int32) ([]*v1.WorldChange, error) {
	// Parse tile position
	target, err := g.Pos(tile)
	if err != nil {
		return nil, fmt.Errorf("invalid tile position %q: %w", tile, err)
	}
	if target.Tile == nil {
		return nil, fmt.Errorf("no tile at position %q", tile)
	}

	// Create build action
	action := &v1.BuildUnitAction{
		Pos:      target.Position(),
		UnitType: unitType,
	}

	move := &v1.GameMove{
		Player:   g.CurrentPlayer,
		MoveType: &v1.GameMove_BuildUnit{BuildUnit: action},
	}

	// Process build
	if err := g.ProcessBuildUnit(move, action); err != nil {
		return nil, err
	}

	return move.Changes, nil
}

// Capture starts capturing building with unit at position.
// unit: position string for the capturing unit
// Returns world changes from the capture action.
func (g *Game) Capture(unit string) ([]*v1.WorldChange, error) {
	// Parse unit position
	target, err := g.Pos(unit)
	if err != nil {
		return nil, fmt.Errorf("invalid unit position %q: %w", unit, err)
	}
	if target.Unit == nil {
		return nil, fmt.Errorf("no unit at position %q", unit)
	}

	// Create capture action
	action := &v1.CaptureBuildingAction{
		Pos: target.Position(),
	}

	move := &v1.GameMove{
		Player:   g.CurrentPlayer,
		MoveType: &v1.GameMove_CaptureBuilding{CaptureBuilding: action},
	}

	// Process capture
	if err := g.ProcessCaptureBuilding(move, action); err != nil {
		return nil, err
	}

	return move.Changes, nil
}

// EndTurn advances to next player.
// Returns world changes from ending the turn.
func (g *Game) EndTurn() ([]*v1.WorldChange, error) {
	action := &v1.EndTurnAction{}

	move := &v1.GameMove{
		Player:   g.CurrentPlayer,
		MoveType: &v1.GameMove_EndTurn{EndTurn: action},
	}

	// Process end turn
	if err := g.ProcessEndTurn(move, action); err != nil {
		return nil, err
	}

	return move.Changes, nil
}

// =============================================================================
// Options Methods - Query available actions
// =============================================================================

// GetOptionsAt returns available options at a position.
// position: position string ("A1", "3,4", "t:A1", etc.)
// Returns the options response with available actions.
func (g *Game) GetOptionsAt(position string) (*v1.GetOptionsAtResponse, error) {
	// Parse the position
	target, err := g.Pos(position)
	if err != nil {
		return &v1.GetOptionsAtResponse{
			Options:         []*v1.GameOption{},
			CurrentPlayer:   g.CurrentPlayer,
			GameInitialized: true,
		}, fmt.Errorf("invalid position: %w", err)
	}

	coord := target.Coordinate
	unit := g.World.UnitAt(coord)
	tile := g.World.TileAt(coord)

	// Lazy top-up if there's a unit
	if unit != nil {
		if err := g.TopUpUnitIfNeeded(unit); err != nil {
			return nil, fmt.Errorf("failed to top-up unit: %w", err)
		}
	}

	var options []*v1.GameOption
	var allPaths *v1.AllPaths

	if unit == nil {
		options, err = g.GetTileOptions(tile)
	} else {
		options, allPaths, err = g.GetUnitOptions(unit)
	}
	if err != nil {
		return nil, err
	}

	return &v1.GetOptionsAtResponse{
		Options:         options,
		CurrentPlayer:   g.CurrentPlayer,
		GameInitialized: g.World != nil,
		AllPaths:        allPaths,
	}, nil
}

// GetUnitOptions returns available options for a unit (move, attack, capture).
func (g *Game) GetUnitOptions(unit *v1.Unit) (options []*v1.GameOption, allPaths *v1.AllPaths, err error) {
	// Get unit definition for progression rules
	unitDef, err := g.RulesEngine.GetUnitData(unit.UnitType)
	if err != nil {
		unitDef = &v1.UnitDefinition{
			ActionOrder: []string{"move", "attack|capture"},
		}
	}

	// Get allowed actions based on progression state
	allowedActions := g.RulesEngine.GetAllowedActionsForUnit(unit, unitDef)

	moveAllowed := ContainsAction(allowedActions, "move")
	retreatAllowed := ContainsAction(allowedActions, "retreat")

	// Get movement options
	if unit.AvailableHealth > 0 && unit.DistanceLeft > 0 && (moveAllowed || retreatAllowed) {
		pathsResult, err := g.GetMovementOptions(unit.Q, unit.R, false)
		if err == nil {
			allPaths = pathsResult

			for _, edge := range allPaths.Edges {
				if edge.IsOccupied {
					continue
				}

				path, err := ReconstructPath(allPaths, edge.ToQ, edge.ToR)
				if err != nil {
					continue
				}

				moveAction := &v1.MoveUnitAction{
					From:              &v1.Position{Label: unit.Shortcut, Q: unit.Q, R: unit.R},
					To:                &v1.Position{Q: edge.ToQ, R: edge.ToR},
					MovementCost:      edge.TotalCost,
					ReconstructedPath: path,
				}

				options = append(options, &v1.GameOption{
					OptionType: &v1.GameOption_Move{Move: moveAction},
				})
			}
		}
	}

	// Check if attack is allowed (including look-ahead for point-based steps)
	attackAllowed := ContainsAction(allowedActions, "attack")
	isPointBasedStep := moveAllowed || retreatAllowed
	if isPointBasedStep && !attackAllowed {
		nextStepUnit := &v1.Unit{
			ProgressionStep:   unit.ProgressionStep + 1,
			ChosenAlternative: "",
			DistanceLeft:      0,
		}
		nextAllowedActions := g.RulesEngine.GetAllowedActionsForUnit(nextStepUnit, unitDef)
		attackAllowed = ContainsAction(nextAllowedActions, "attack")
	}

	// Get attack options
	if unit.AvailableHealth > 0 && attackAllowed {
		attackCoords, err := g.GetAttackOptions(unit.Q, unit.R)
		if err == nil {
			for _, coord := range attackCoords {
				targetUnit := g.World.UnitAt(coord)
				if targetUnit != nil {
					damageEstimate := int32(50) // TODO: Use proper damage calculation

					attackAction := &v1.AttackUnitAction{
						Attacker:         &v1.Position{Label: unit.Shortcut, Q: unit.Q, R: unit.R},
						Defender:         &v1.Position{Q: int32(coord.Q), R: int32(coord.R)},
						TargetUnitType:   targetUnit.UnitType,
						TargetUnitHealth: targetUnit.AvailableHealth,
						CanAttack:        true,
						DamageEstimate:   damageEstimate,
					}

					options = append(options, &v1.GameOption{
						OptionType: &v1.GameOption_Attack{Attack: attackAction},
					})
				}
			}
		}
	}

	// Check if capture is allowed (including look-ahead)
	captureAllowed := ContainsAction(allowedActions, "capture")
	if isPointBasedStep && !captureAllowed {
		nextStepUnit := &v1.Unit{
			ProgressionStep:   unit.ProgressionStep + 1,
			ChosenAlternative: "",
			DistanceLeft:      0,
		}
		nextAllowedActions := g.RulesEngine.GetAllowedActionsForUnit(nextStepUnit, unitDef)
		captureAllowed = ContainsAction(nextAllowedActions, "capture")
	}

	// Get capture option
	if unit.AvailableHealth > 0 && captureAllowed && unit.CaptureStartedTurn == 0 {
		coord := CoordFromInt32(unit.Q, unit.R)
		tile := g.World.TileAt(coord)
		if tile != nil && tile.Player != unit.Player {
			terrainProps := g.RulesEngine.GetTerrainUnitPropertiesForUnit(tile.TileType, unit.UnitType)
			if terrainProps != nil && terrainProps.CanCapture {
				captureAction := &v1.CaptureBuildingAction{
					Pos:      &v1.Position{Label: unit.Shortcut, Q: unit.Q, R: unit.R},
					TileType: tile.TileType,
				}
				options = append(options, &v1.GameOption{
					OptionType: &v1.GameOption_Capture{Capture: captureAction},
				})
			}
		}
	}

	// Get heal option if unit is below max health and can heal on current terrain
	// Heal is available if unit hasn't acted this turn yet
	if unit.AvailableHealth > 0 && unit.AvailableHealth < unitDef.Health && unit.LastActedTurn < g.TurnCounter {
		coord := CoordFromInt32(unit.Q, unit.R)
		tile := g.World.TileAt(coord)
		if tile != nil {
			// Can't heal on enemy-owned tiles
			if tile.Player == 0 || tile.Player == unit.Player {
				canHeal := true
				// Air units can only heal on Airport Base
				if unitDef.UnitTerrain == "Air" {
					terrainData, _ := g.RulesEngine.GetTerrainData(tile.TileType)
					if terrainData == nil || terrainData.Name != "Airport Base" {
						canHeal = false
					}
				}
				if canHeal {
					terrainProps := g.RulesEngine.GetTerrainUnitPropertiesForUnit(tile.TileType, unit.UnitType)
					if terrainProps != nil && terrainProps.HealingBonus > 0 {
						healAction := &v1.HealUnitAction{
							Pos:        &v1.Position{Label: unit.Shortcut, Q: unit.Q, R: unit.R},
							HealAmount: terrainProps.HealingBonus,
						}
						options = append(options, &v1.GameOption{
							OptionType: &v1.GameOption_Heal{Heal: healAction},
						})
					}
				}
			}
		}
	}

	return
}

// GetTileOptions returns available options for a tile (build units).
func (g *Game) GetTileOptions(tile *v1.Tile) (options []*v1.GameOption, err error) {
	if tile == nil {
		return nil, nil
	}

	// Lazy top-up
	if err := g.TopUpTileIfNeeded(tile); err != nil {
		return nil, fmt.Errorf("failed to top-up tile: %w", err)
	}

	// Only check tile actions if tile belongs to current player
	if tile.Player != g.CurrentPlayer {
		return nil, nil
	}

	terrainDef, err := g.RulesEngine.GetTerrainData(tile.TileType)
	if err != nil {
		return nil, nil
	}

	// Get current player's coins
	playerCoins := int32(0)
	if playerState := g.GameState.PlayerStates[g.CurrentPlayer]; playerState != nil {
		playerCoins = playerState.Coins
	}

	// Get allowed actions for this tile
	tileActions := g.RulesEngine.GetAllowedActionsForTile(tile, terrainDef, playerCoins)

	for _, action := range tileActions {
		if action == "build" {
			// Filter buildable units by game's allowed units setting
			buildableUnits := FilterBuildOptionsByAllowedUnits(
				terrainDef.BuildableUnitIds,
				g.Config.Settings.GetAllowedUnits(),
			)

			for _, unitTypeID := range buildableUnits {
				unitDef, err := g.RulesEngine.GetUnitData(unitTypeID)
				if err != nil {
					continue
				}

				if unitDef.Coins <= playerCoins {
					options = append(options, &v1.GameOption{
						OptionType: &v1.GameOption_Build{
							Build: &v1.BuildUnitAction{
								Pos:      &v1.Position{Label: tile.Shortcut, Q: tile.Q, R: tile.R},
								UnitType: unitTypeID,
								Cost:     unitDef.Coins,
							},
						},
					})
				}
			}
		}
	}

	return
}

// FilterBuildOptionsByAllowedUnits filters buildable units by the game's allowed units setting.
func FilterBuildOptionsByAllowedUnits(buildableUnits, allowedUnits []int32) []int32 {
	if allowedUnits == nil {
		return buildableUnits
	}

	allowedSet := make(map[int32]bool)
	for _, unitID := range allowedUnits {
		allowedSet[unitID] = true
	}

	var filtered []int32
	for _, unitID := range buildableUnits {
		if allowedSet[unitID] {
			filtered = append(filtered, unitID)
		}
	}
	return filtered
}
