package weewar

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// =============================================================================
// Core Types (from core.go)
// =============================================================================

// Game represents the unified game state and implements GameInterface
type Game struct {
	// Pure game state
	World *World `json:"world"` // Contains the pure state (map, units, entities)

	// Game flow control
	CurrentPlayer int        `json:"currentPlayer"` // 0-based player index
	TurnCounter   int        `json:"turnCounter"`   // 1-based turn number
	Status        GameStatus `json:"status"`        // Game status

	// Game systems and configuration
	Seed int64 `json:"seed"` // Random seed for deterministic gameplay

	// Random number generator
	rng *rand.Rand `json:"-"` // RNG for deterministic gameplay

	// Event system
	eventManager *EventManager `json:"-"` // Event manager for observer pattern

	// Asset management
	assetProvider AssetProvider `json:"-"` // Asset provider for tiles and units (interface for platform flexibility)

	// Game metadata
	CreatedAt    time.Time `json:"createdAt"`    // When game was created
	LastActionAt time.Time `json:"lastActionAt"` // When last action was taken

	// Internal state
	winner    int  `json:"winner"`    // Winner player ID (-1 if no winner)
	hasWinner bool `json:"hasWinner"` // Whether game has ended with winner
}

// =============================================================================
// Convenience methods to access World fields
// =============================================================================

// Map returns the game map
func (g *Game) Map() *Map {
	if g.World == nil {
		return nil
	}
	return g.World.Map
}

// Units returns all units in the world
func (g *Game) Units() []*Unit {
	if g.World == nil {
		return nil
	}
	return g.World.Units
}

// PlayerCount returns the number of players
func (g *Game) PlayerCount() int {
	if g.World == nil {
		return 0
	}
	return g.World.PlayerCount
}

// =============================================================================
// Helper Functions
// =============================================================================

// initializeStartingUnits adds initial units to the game
func (g *Game) initializeStartingUnits() error {
	// Add some basic starting units for each player using cube coordinates
	startingPositions := [][]CubeCoord{
		{{Q: 1, R: 1}, {Q: 2, R: 1}},  // Player 0
		{{Q: 9, R: 6}, {Q: 10, R: 6}}, // Player 1
	}

	for playerID := 0; playerID < g.World.PlayerCount && playerID < len(startingPositions); playerID++ {
		positions := startingPositions[playerID]

		for _, coord := range positions {
			// Create a basic infantry unit
			unit := NewUnit(1, playerID) // Unit type 1 = Infantry
			unit.Coord = coord
			unit.AvailableHealth = 100
			unit.DistanceLeft = 3

			// Add unit to game
			g.AddUnit(unit, playerID)
		}
	}

	return nil
}

// resetPlayerUnits resets movement and actions for a player's units
func (g *Game) resetPlayerUnits(playerID int) error {
	if playerID < 0 || playerID >= len(g.World.Units) {
		return fmt.Errorf("invalid player ID: %d", playerID)
	}

	for _, unit := range g.World.Units[playerID] {
		// Reset movement points (simplified)
		unit.DistanceLeft = 3 // TODO: Get from unit data
		unit.TurnCounter = g.TurnCounter
	}

	return nil
}

// checkVictoryConditions checks if any player has won
func (g *Game) checkVictoryConditions() (winner int, hasWinner bool) {
	// Simple victory condition: last player with units wins
	playersWithUnits := 0
	lastPlayerWithUnits := -1

	for playerID := 0; playerID < g.World.PlayerCount; playerID++ {
		if len(g.World.Units[playerID]) > 0 {
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
	if g.World.Map == nil {
		return fmt.Errorf("game has no map")
	}

	if g.World.PlayerCount < 2 || g.World.PlayerCount > 6 {
		return fmt.Errorf("invalid player count: %d", g.World.PlayerCount)
	}

	if g.CurrentPlayer < 0 || g.CurrentPlayer >= g.World.PlayerCount {
		return fmt.Errorf("invalid current player: %d", g.CurrentPlayer)
	}

	if g.TurnCounter < 1 {
		return fmt.Errorf("invalid turn counter: %d", g.TurnCounter)
	}

	if len(g.World.Units) != g.World.PlayerCount {
		return fmt.Errorf("units array length (%d) doesn't match player count (%d)", len(g.World.Units), g.World.PlayerCount)
	}

	return nil
}

// GetUnitID generates a unique identifier for a unit in the format PN
// where P is the player letter (A-Z) and N is the unit number for that player
func (g *Game) GetUnitID(unit *Unit) string {
	if unit == nil {
		return ""
	}

	// Convert player ID to letter (0=A, 1=B, etc.)
	playerLetter := string(rune('A' + unit.PlayerID))

	// Count units for this player to determine unit number
	unitNumber := 0
	for _, playerUnits := range g.World.Units {
		for _, playerUnit := range playerUnits {
			if playerUnit.PlayerID == unit.PlayerID {
				unitNumber++
				if playerUnit == unit {
					// Found our unit, return the ID
					return fmt.Sprintf("%s%d", playerLetter, unitNumber)
				}
			}
		}
	}

	// Fallback - shouldn't happen but handle gracefully
	return fmt.Sprintf("%s?", playerLetter)
}

// GetAssetManager returns the current AssetManager instance (legacy compatibility)
func (g *Game) GetAssetManager() *AssetManager {
	// Try to cast the AssetProvider to *AssetManager for backward compatibility
	if am, ok := g.assetProvider.(*AssetManager); ok {
		return am
	}
	return nil
}

// SetAssetManager sets the AssetManager instance for tile and unit rendering (legacy compatibility)
func (g *Game) SetAssetManager(assetManager *AssetManager) {
	g.assetProvider = assetManager
}

// GetAssetProvider returns the current AssetProvider instance
func (g *Game) GetAssetProvider() AssetProvider {
	return g.assetProvider
}

// SetAssetProvider sets the AssetProvider instance for tile and unit rendering
func (g *Game) SetAssetProvider(provider AssetProvider) {
	g.assetProvider = provider
}

// NewGame creates a new game instance with the specified parameters
func NewGame(world *World, seed int64) (*Game, error) {
	// Validate parameters

	// Create the World (pure state)
	// Create the game struct
	game := &Game{
		World:         world,
		Seed:          seed,
		CurrentPlayer: 0,
		TurnCounter:   1,
		Status:        GameStatusPlaying,
		winner:        -1,
		hasWinner:     false,
		CreatedAt:     time.Now(),
		LastActionAt:  time.Now(),
		rng:           rand.New(rand.NewSource(seed)),
		eventManager:  NewEventManager(),
		assetProvider: NewAssetManager("data"),
	}

	// Initialize units storage for compatibility (will be migrated)

	// Map is already assigned in the struct initialization above

	// Initialize starting units (simplified for now)
	// TODO: Replace with actual unit placement from map data
	if err := game.initializeStartingUnits(); err != nil {
		return nil, fmt.Errorf("failed to initialize starting units: %w", err)
	}

	// Emit game created event
	game.eventManager.EmitGameStateChanged(GameStateChangeGameStarted, game)

	return game, nil
}

// LoadGame restores a game from saved JSON data
func LoadGame(saveData []byte) (*Game, error) {
	var game Game
	if err := json.Unmarshal(saveData, &game); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	// Restore transient state
	game.rng = rand.New(rand.NewSource(game.Seed))
	game.eventManager = NewEventManager()
	game.assetProvider = NewAssetManager("data")

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
	g.LastActionAt = time.Now()

	// Serialize to JSON
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize game state: %w", err)
	}

	return data, nil
}

// GetCurrentPlayer returns active player ID
func (g *Game) GetCurrentPlayer() int {
	return g.CurrentPlayer
}

// GetTurnNumber returns current turn count
func (g *Game) GetTurnNumber() int {
	return g.TurnCounter
}

// GetGameStatus returns current game state
func (g *Game) GetGameStatus() GameStatus {
	return g.Status
}

// GetWinner returns winning player if game ended
func (g *Game) GetWinner() (int, bool) {
	return g.winner, g.hasWinner
}

// =============================================================================
// MapInterface Interface Implementation
// =============================================================================

// GetMapSize returns map dimensions
func (g *Game) GetMapSize() (rows, cols int) {
	if g.World.Map == nil {
		return 0, 0
	}
	return g.World.Map.NumRows(), g.World.Map.NumCols()
}

// GetMapName returns loaded map name
func (g *Game) GetMapName() string {
	return "DefaultMap" // For now, since we're using map instances directly
}

// GetMapBounds returns pixel boundaries for rendering
func (g *Game) GetMapBounds() (minX, minY, maxX, maxY float64) {
	if g.World.Map == nil {
		return 0, 0, 0, 0
	}

	// Use standard tile dimensions for bounds calculation
	tileWidth := DefaultTileWidth
	tileHeight := DefaultTileHeight
	yIncrement := DefaultYIncrement

	return g.World.Map.getMapBounds(tileWidth, tileHeight, yIncrement)
}

// GetTileType returns terrain type at position using cube coordinates
func (g *Game) GetTileType(coord CubeCoord) int {
	tile := g.World.Map.TileAt(coord)
	if tile == nil {
		return 0 // Default/unknown terrain
	}
	return tile.TileType
}

// RowColToPixel converts grid coordinates to screen coordinates
func (g *Game) RowColToPixel(row, col int) (x, y float64) {
	if g.World.Map == nil {
		return 0, 0
	}

	// Use standard tile dimensions
	tileWidth := DefaultTileHeight
	tileHeight := DefaultTileWidth
	yIncrement := DefaultYIncrement

	return g.World.Map.XYForTile(row, col, tileWidth, tileHeight, yIncrement)
}

// PixelToRowCol converts screen coordinates to grid coordinates
func (g *Game) PixelToRowCol(x, y float64) (row, col int, valid bool) {
	if g.World.Map == nil {
		return 0, 0, false
	}

	// Use standard tile dimensions
	tileWidth := DefaultTileWidth
	yIncrement := DefaultYIncrement

	// Calculate approximate row and column
	row = int(y / yIncrement)

	// Calculate column accounting for hex offset
	isEvenRow := (row % 2) == 0
	baseX := x
	if g.World.Map.EvenRowsOffset() {
		if isEvenRow {
			baseX -= tileWidth / 2
		}
	} else {
		if !isEvenRow {
			baseX -= tileWidth / 2
		}
	}
	col = int(baseX / tileWidth)

	// Validate that the calculated position exists on the map
	coord := g.World.Map.DisplayToHex(row, col)
	if tile := g.World.Map.TileAt(coord); tile != nil {
		return row, col, true
	}

	return 0, 0, false
}

// =============================================================================
// UnitInterface Interface Implementation
// =============================================================================

// GetUnitAt returns unit at specific position using cube coordinates
func (g *Game) GetUnitAt(coord CubeCoord) *Unit {
	tile := g.World.Map.TileAt(coord)
	if tile == nil {
		return nil
	}
	return tile.Unit
}

// GetUnitsForPlayer returns all units owned by player
func (g *Game) GetUnitsForPlayer(playerID int) []*Unit {
	if playerID < 0 || playerID >= len(g.World.Units) {
		return nil
	}

	// Return a copy to prevent external modification
	units := make([]*Unit, len(g.World.Units[playerID]))
	copy(units, g.World.Units[playerID])
	return units
}

// GetAllUnits returns every unit on the map
func (g *Game) GetAllUnits() []*Unit {
	var allUnits []*Unit

	for _, playerUnits := range g.World.Units {
		allUnits = append(allUnits, playerUnits...)
	}

	return allUnits
}

// GetUnitType returns unit type identifier
func (g *Game) GetUnitType(unit *Unit) int {
	if unit == nil {
		return 0
	}
	return unit.UnitType
}

// GetUnitTypeName returns the display name for a unit type
func (g *Game) GetUnitTypeName(unitType int) string {
	if g.assetProvider != nil {
		// Try to get unit data from JSON if asset provider is loaded
		if am, ok := g.assetProvider.(*AssetManager); ok {
			if err := am.LoadGameData(); err == nil {
				if unitData, err := am.GetUnitData(unitType); err == nil {
					return unitData.Name
				}
			}
		}
	}

	// Fallback to generic name
	return fmt.Sprintf("Unit Type %d", unitType)
}

// GetUnitHealth returns current health points
func (g *Game) GetUnitHealth(unit *Unit) int {
	if unit == nil {
		return 0
	}
	return unit.AvailableHealth
}

// CreateUnit spawns new unit using cube coordinates
func (g *Game) CreateUnit(unitType, playerID int, coord CubeCoord) (*Unit, error) {
	// Validate parameters
	if playerID < 0 || playerID >= g.World.PlayerCount {
		return nil, fmt.Errorf("invalid player ID: %d", playerID)
	}

	// Check if position is valid and empty
	tile := g.World.Map.TileAt(coord)
	if tile == nil {
		return nil, fmt.Errorf("invalid position: %v", coord)
	}

	if tile.Unit != nil {
		return nil, fmt.Errorf("position %v is occupied", coord)
	}

	// Create the unit
	unit := NewUnit(unitType, playerID)
	unit.Coord = coord
	unit.AvailableHealth = 100 // TODO: Get from unit data
	unit.DistanceLeft = 3      // TODO: Get from unit data

	// Add to game
	g.AddUnit(unit, playerID)

	// Emit events
	g.eventManager.EmitUnitCreated(unit)
	g.eventManager.EmitGameStateChanged(GameStateChangeUnitCreated, unit)

	return unit, nil
}

// RemoveUnit removes unit from game
func (g *Game) RemoveUnit(unit *Unit) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	// Remove from tile using cube coordinates
	tile := g.World.Map.TileAt(unit.Coord)
	if tile != nil && tile.Unit == unit {
		tile.Unit = nil
	}

	// Remove from player's unit list
	if unit.PlayerID >= 0 && unit.PlayerID < len(g.World.Units) {
		playerUnits := g.World.Units[unit.PlayerID]
		for i, u := range playerUnits {
			if u == unit {
				// Remove from slice
				g.World.Units[unit.PlayerID] = append(playerUnits[:i], playerUnits[i+1:]...)
				break
			}
		}
	}

	// Emit events
	g.eventManager.EmitUnitDestroyed(unit)
	g.eventManager.EmitGameStateChanged(GameStateChangeUnitDestroyed, unit)

	return nil
}

// AddUnit adds a unit to the game for the specified player
func (g *Game) AddUnit(unit *Unit, playerID int) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	if playerID < 0 || playerID >= len(g.World.Units) {
		return fmt.Errorf("invalid player ID: %d", playerID)
	}

	// Set unit's player ID
	unit.PlayerID = playerID

	// Add to player's unit list
	g.World.Units[playerID] = append(g.World.Units[playerID], unit)

	// Place unit on the map if it has a valid position
	if tile := g.World.Map.TileAt(unit.Coord); tile != nil {
		tile.Unit = unit
	}

	return nil
}

// Helper math functions
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
