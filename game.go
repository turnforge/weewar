package weewar

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// =============================================================================
// Core Types (from core.go)
// =============================================================================

// TerrainData represents terrain type information
type TerrainData struct {
	ID           int
	Name         string
	MoveCost     int
	DefenseBonus int
}

// Basic terrain data
var terrainData = []TerrainData{
	{0, "Unknown", 1, 0},
	{1, "Grass", 1, 0},
	{2, "Desert", 1, 0},
	{3, "Water", 2, 0},
	{4, "Mountain", 2, 10},
	{5, "Rock", 3, 20},
}

// GetTerrainData returns terrain data for the given type
func GetTerrainData(terrainType int) *TerrainData {
	for i := range terrainData {
		if terrainData[i].ID == terrainType {
			return &terrainData[i]
		}
	}
	return &terrainData[0] // Default to unknown
}

// Unit represents a runtime unit instance in the game
type Unit struct {
	UnitType int // Reference to UnitData by ID

	// Runtime state
	DistanceLeft    int // Movement points remaining this turn
	AvailableHealth int // Current health points
	TurnCounter     int // Which turn this unit was created/last acted

	// Position on the map
	Row int
	Col int

	// Player ownership
	PlayerID int
}

// NeighborDirection represents the 6 directions in a hex grid
type NeighborDirection int

const (
	LEFT NeighborDirection = iota
	TOP_LEFT
	TOP_RIGHT
	RIGHT
	BOTTOM_RIGHT
	BOTTOM_LEFT
)

// Tile represents a single hex tile on the map
type Tile struct {
	Row int `json:"row"`
	Col int `json:"col"`

	TileType int `json:"tileType"` // Reference to TerrainData by ID

	// Optional: Unit occupying this tile
	Unit *Unit `json:"unit"`
}

// TileWithCoord represents a tile with its cube coordinate for JSON serialization
type TileWithCoord struct {
	Coord CubeCoord `json:"coord"`
	Tile  *Tile    `json:"tile"`
}

// Map represents the game map with hex grid topology
type Map struct {
	// Display bounds (for UI rendering only)
	NumRows int `json:"numRows"`
	NumCols int `json:"numCols"`

	// Cube coordinate storage - primary data structure
	Tiles map[CubeCoord]*Tile `json:"-"` // Direct cube coordinate lookup (custom JSON handling)
	
	// JSON-friendly representation
	TileList []*TileWithCoord `json:"tiles"`
}

// NewMap creates a new empty map with the specified dimensions
// evenRowsOffset parameter is deprecated and ignored (cube coordinates are universal)
func NewMap(numRows, numCols int, evenRowsOffset bool) *Map {
	_ = evenRowsOffset // Deprecated: cube coordinates eliminate offset confusion
	return &Map{
		NumRows:  numRows,
		NumCols:  numCols,
		Tiles:    make(map[CubeCoord]*Tile),
		TileList: make([]*TileWithCoord, 0),
	}
}

// =============================================================================
// JSON Serialization Methods
// =============================================================================

// MarshalJSON implements custom JSON marshaling for Map
func (m *Map) MarshalJSON() ([]byte, error) {
	// Convert cube map to tile list for JSON
	m.syncTileListFromMap()
	
	// Create a temporary struct with the same fields
	type mapJSON Map
	return json.Marshal((*mapJSON)(m))
}

// UnmarshalJSON implements custom JSON unmarshaling for Map
func (m *Map) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with the same fields
	type mapJSON Map
	if err := json.Unmarshal(data, (*mapJSON)(m)); err != nil {
		return err
	}
	
	// Initialize the cube map if it's nil
	if m.Tiles == nil {
		m.Tiles = make(map[CubeCoord]*Tile)
	}
	
	// Convert tile list back to cube map
	m.syncMapFromTileList()
	return nil
}

// syncTileListFromMap converts the cube map to tile list for JSON serialization
func (m *Map) syncTileListFromMap() {
	m.TileList = make([]*TileWithCoord, 0, len(m.Tiles))
	for coord, tile := range m.Tiles {
		m.TileList = append(m.TileList, &TileWithCoord{
			Coord: coord,
			Tile:  tile,
		})
	}
}

// syncMapFromTileList converts the tile list back to cube map after JSON deserialization
func (m *Map) syncMapFromTileList() {
	for _, tileWithCoord := range m.TileList {
		m.Tiles[tileWithCoord.Coord] = tileWithCoord.Tile
	}
}

// =============================================================================
// Primary Cube-Based Storage Methods
// =============================================================================

// TileAtCube returns the tile at the specified cube coordinate (primary method)
func (m *Map) TileAtCube(coord CubeCoord) *Tile {
	return m.Tiles[coord]
}

// AddTileCube adds a tile at the specified cube coordinate (primary method)
func (m *Map) AddTileCube(coord CubeCoord, tile *Tile) {
	// Update tile's position to match cube coordinate
	tile.Row, tile.Col = m.HexToDisplay(coord)
	m.Tiles[coord] = tile
}

// DeleteTileCube removes the tile at the specified cube coordinate
func (m *Map) DeleteTileCube(coord CubeCoord) {
	delete(m.Tiles, coord)
}

// GetAllTiles returns all tiles as a map from cube coordinates to tiles
func (m *Map) GetAllTiles() map[CubeCoord]*Tile {
	// Return a copy to prevent external modification
	result := make(map[CubeCoord]*Tile)
	for coord, tile := range m.Tiles {
		result[coord] = tile
	}
	return result
}

// =============================================================================
// Display Coordinate Conversion Methods
// =============================================================================

// HexToDisplay converts cube coordinates to display coordinates (row, col)
// Uses a standard hex-to-array conversion (odd-row offset style)
func (m *Map) HexToDisplay(coord CubeCoord) (row, col int) {
	row = coord.R
	col = coord.Q + (coord.R + (coord.R & 1)) / 2
	return row, col
}

// DisplayToHex converts display coordinates (row, col) to cube coordinates
// Uses a standard array-to-hex conversion (odd-row offset style)
func (m *Map) DisplayToHex(row, col int) CubeCoord {
	q := col - (row + (row & 1)) / 2
	return NewCubeCoord(q, row)
}

// =============================================================================
// Legacy Display-Based Methods (for backward compatibility)
// =============================================================================

// EvenRowsOffset returns the offset configuration (deprecated - cube coordinates are universal)
func (m *Map) EvenRowsOffset() bool {
	// Deprecated: cube coordinates eliminate the need for offset configuration
	// This method exists only for backward compatibility and always returns false
	return false
}

// TileAt returns the tile at the specified display position (legacy method)
func (m *Map) TileAt(row, col int) *Tile {
	coord := m.DisplayToHex(row, col)
	return m.TileAtCube(coord)
}

// AddTile adds a tile to the map at its specified row/column position (legacy method)
func (m *Map) AddTile(t *Tile) {
	coord := m.DisplayToHex(t.Row, t.Col)
	m.AddTileCube(coord, t)
}

// GetHexNeighborCoords returns the coordinates of the 6 hex neighbors
func (m *Map) GetHexNeighborCoords(row, col int) [6][2]int {
	var neighbors [6][2]int

	// Hex grid neighbor calculation depends on whether we're in even or odd row
	isEvenRow := (row % 2) == 0

	if m.EvenRowsOffset() {
		// Even rows are offset to the right
		if isEvenRow {
			// Even row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col}     // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col + 1} // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col + 1} // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col}     // BOTTOM_LEFT
		} else {
			// Odd row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col - 1} // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col}     // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col}     // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col - 1} // BOTTOM_LEFT
		}
	} else {
		// Odd rows are offset to the right
		if isEvenRow {
			// Even row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col - 1} // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col}     // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col}     // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col - 1} // BOTTOM_LEFT
		} else {
			// Odd row neighbors
			neighbors[0] = [2]int{row, col - 1}     // LEFT
			neighbors[1] = [2]int{row - 1, col}     // TOP_LEFT
			neighbors[2] = [2]int{row - 1, col + 1} // TOP_RIGHT
			neighbors[3] = [2]int{row, col + 1}     // RIGHT
			neighbors[4] = [2]int{row + 1, col + 1} // BOTTOM_RIGHT
			neighbors[5] = [2]int{row + 1, col}     // BOTTOM_LEFT
		}
	}

	return neighbors
}

// GetNeighbor returns the neighboring tile in the specified direction - maintains backward compatibility
func (m *Map) GetNeighbor(row, col int, direction NeighborDirection) *Tile {
	coords := m.GetHexNeighborCoords(row, col)
	neighborRow, neighborCol := coords[direction][0], coords[direction][1]
	return m.TileAt(neighborRow, neighborCol)
}

// XYForTile converts tile row/col coordinates to pixel x,y coordinates for rendering
func (m *Map) XYForTile(row, col int, tileWidth, tileHeight, yIncrement float64) (x, y float64) {
	// Calculate base x position with margin to ensure tiles are fully within bounds
	x = float64(col)*tileWidth + tileWidth/2

	// Apply offset for alternating rows (hex grid staggering)
	isEvenRow := (row % 2) == 0
	if m.EvenRowsOffset() {
		// Even rows are offset to the right
		if isEvenRow {
			x += tileWidth / 2
		}
	} else {
		// Odd rows are offset to the right
		if !isEvenRow {
			x += tileWidth / 2
		}
	}

	// Calculate y position with margin to ensure tiles are fully within bounds
	y = float64(row)*yIncrement + tileHeight/2

	return x, y
}

// getMapBounds calculates the pixel bounds of the entire map
func (m *Map) getMapBounds(tileWidth, tileHeight, yIncrement float64) (minX, minY, maxX, maxY float64) {
	minX = math.Inf(1)
	minY = math.Inf(1)
	maxX = math.Inf(-1)
	maxY = math.Inf(-1)

	for _, tile := range m.Tiles {
		row, col := tile.Row, tile.Col
		x, y := m.XYForTile(row, col, tileWidth, tileHeight, yIncrement)

		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x+tileWidth > maxX {
			maxX = x + tileWidth
		}
		if y+tileHeight > maxY {
			maxY = y + tileHeight
		}
	}

	return minX, minY, maxX, maxY
}


// DeleteTile removes the tile at the specified position (legacy method)
func (m *Map) DeleteTile(row, col int) {
	coord := m.DisplayToHex(row, col)
	m.DeleteTileCube(coord)
}

// NewTile creates a new tile at the specified position
func NewTile(row, col, tileType int) *Tile {
	return &Tile{
		Row:      row,
		Col:      col,
		TileType: tileType,
	}
}

// NewUnit creates a new unit instance
func NewUnit(unitType, playerID int) *Unit {
	return &Unit{
		UnitType:        unitType,
		PlayerID:        playerID,
		DistanceLeft:    0, // Will be set based on UnitData
		AvailableHealth: 0, // Will be set based on UnitData
		TurnCounter:     0,
	}
}

// =============================================================================
// Unified Game Implementation
// =============================================================================
// This file implements the unified Game struct that combines the best parts
// of the existing core.go with the new interface architecture. It provides
// a single, coherent implementation of all core game interfaces.

// Game represents the unified game state and implements GameInterface
type Game struct {
	// Core game state (from core.go)
	Map           *Map       `json:"map"`
	Units         [][]*Unit  `json:"units"`         // Units[playerID][unitIndex]
	CurrentPlayer int        `json:"currentPlayer"` // 0-based player index
	TurnCounter   int        `json:"turnCounter"`   // 1-based turn number
	Status        GameStatus `json:"status"`        // Game status

	// Game configuration
	PlayerCount int   `json:"playerCount"` // Number of players
	Seed        int64 `json:"seed"`        // Random seed for deterministic gameplay

	// Random number generator
	rng *rand.Rand `json:"-"` // RNG for deterministic gameplay

	// Event system
	eventManager *EventManager `json:"-"` // Event manager for observer pattern

	// Asset management
	assetManager *AssetManager `json:"-"` // Asset manager for tiles and units

	// Game metadata
	CreatedAt    time.Time `json:"createdAt"`    // When game was created
	LastActionAt time.Time `json:"lastActionAt"` // When last action was taken

	// Internal state
	winner    int  `json:"winner"`    // Winner player ID (-1 if no winner)
	hasWinner bool `json:"hasWinner"` // Whether game has ended with winner
}

// =============================================================================
// Game Creation and Initialization
// =============================================================================

// NewGame creates a new game instance with the specified parameters
func NewGame(playerCount int, gameMap *Map, seed int64) (*Game, error) {
	// Validate parameters
	if playerCount < 2 || playerCount > 6 {
		return nil, fmt.Errorf("invalid player count: %d (must be 2-6)", playerCount)
	}

	if gameMap == nil {
		return nil, fmt.Errorf("map cannot be nil")
	}

	// Create the game struct
	game := &Game{
		Map:           gameMap,
		PlayerCount:   playerCount,
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
		assetManager:  NewAssetManager("data"),
	}

	// Initialize units slice
	game.Units = make([][]*Unit, playerCount)
	for i := range game.Units {
		game.Units[i] = make([]*Unit, 0)
	}

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
	game.assetManager = NewAssetManager("data")

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

// NextTurn advances to next player's turn
func (g *Game) NextTurn() error {
	if g.Status != GameStatusPlaying {
		return fmt.Errorf("cannot advance turn: game is not in playing state")
	}

	// Reset unit movement for current player
	if err := g.resetPlayerUnits(g.CurrentPlayer); err != nil {
		return fmt.Errorf("failed to reset player units: %w", err)
	}

	// Advance to next player
	g.CurrentPlayer = (g.CurrentPlayer + 1) % g.PlayerCount

	// If we've cycled back to player 0, increment turn counter
	if g.CurrentPlayer == 0 {
		g.TurnCounter++
	}

	// Check for victory conditions
	if winner, hasWinner := g.checkVictoryConditions(); hasWinner {
		g.winner = winner
		g.hasWinner = true
		g.Status = GameStatusEnded
		g.eventManager.EmitGameEnded(winner)
		g.eventManager.EmitGameStateChanged(GameStateChangeGameEnded, winner)
	}

	// Update timestamp
	g.LastActionAt = time.Now()

	// Emit turn changed event
	g.eventManager.EmitTurnChanged(g.CurrentPlayer, g.TurnCounter)
	g.eventManager.EmitGameStateChanged(GameStateChangeTurnChanged, map[string]interface{}{
		"newPlayer":  g.CurrentPlayer,
		"turnNumber": g.TurnCounter,
	})

	return nil
}

// EndTurn completes current player's turn
func (g *Game) EndTurn() error {
	if g.Status != GameStatusPlaying {
		return fmt.Errorf("cannot end turn: game is not in playing state")
	}

	// For now, EndTurn is the same as NextTurn
	// In a full implementation, this might involve different logic
	// (e.g., checking if player has mandatory actions to complete)
	return g.NextTurn()
}

// CanEndTurn checks if current player can end their turn
func (g *Game) CanEndTurn() bool {
	if g.Status != GameStatusPlaying {
		return false
	}

	// For now, player can always end turn
	// In a full implementation, this might check:
	// - Whether player has units that must move
	// - Whether player has mandatory actions to complete
	// - Whether player has captured a base this turn
	return true
}

// =============================================================================
// MapInterface Interface Implementation
// =============================================================================

// GetMapSize returns map dimensions
func (g *Game) GetMapSize() (rows, cols int) {
	if g.Map == nil {
		return 0, 0
	}
	return g.Map.NumRows, g.Map.NumCols
}

// GetMapName returns loaded map name
func (g *Game) GetMapName() string {
	return "DefaultMap" // For now, since we're using map instances directly
}

// GetMapBounds returns pixel boundaries for rendering
func (g *Game) GetMapBounds() (minX, minY, maxX, maxY float64) {
	if g.Map == nil {
		return 0, 0, 0, 0
	}

	// Use standard tile dimensions for bounds calculation
	tileWidth := 60.0
	tileHeight := 52.0
	yIncrement := 39.0

	return g.Map.getMapBounds(tileWidth, tileHeight, yIncrement)
}

// GetTileAt returns tile at specific position
func (g *Game) GetTileAt(row, col int) *Tile {
	if g.Map == nil {
		return nil
	}
	return g.Map.TileAt(row, col)
}

// GetTileType returns terrain type at position
func (g *Game) GetTileType(row, col int) int {
	tile := g.GetTileAt(row, col)
	if tile == nil {
		return 0 // Default/unknown terrain
	}
	return tile.TileType
}

// GetTileNeighbors returns adjacent tiles (hex grid) - maintains backward compatibility
func (g *Game) GetTileNeighbors(row, col int) []*Tile {
	if g.Map == nil {
		return make([]*Tile, 6)
	}

	// Use original array-based calculation for backward compatibility
	neighborCoords := g.Map.GetHexNeighborCoords(row, col)
	neighbors := make([]*Tile, 6)

	for i, coord := range neighborCoords {
		neighbors[i] = g.Map.TileAt(coord[0], coord[1])
	}

	return neighbors
}

// RowColToPixel converts grid coordinates to screen coordinates
func (g *Game) RowColToPixel(row, col int) (x, y float64) {
	if g.Map == nil {
		return 0, 0
	}

	// Use standard tile dimensions
	tileWidth := 60.0
	tileHeight := 52.0
	yIncrement := 39.0

	return g.Map.XYForTile(row, col, tileWidth, tileHeight, yIncrement)
}

// PixelToRowCol converts screen coordinates to grid coordinates
func (g *Game) PixelToRowCol(x, y float64) (row, col int, valid bool) {
	if g.Map == nil {
		return 0, 0, false
	}

	// Use standard tile dimensions
	tileWidth := 60.0
	yIncrement := 39.0

	// Calculate approximate row and column
	row = int(y / yIncrement)

	// Calculate column accounting for hex offset
	isEvenRow := (row % 2) == 0
	baseX := x
	if g.Map.EvenRowsOffset() {
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
	if tile := g.GetTileAt(row, col); tile != nil {
		return row, col, true
	}

	return 0, 0, false
}

// FindPath calculates movement path between positions
func (g *Game) FindPath(fromRow, fromCol, toRow, toCol int) ([]Tile, error) {
	if g.Map == nil {
		return nil, fmt.Errorf("no map loaded")
	}

	// Check if start and end positions are valid
	startTile := g.GetTileAt(fromRow, fromCol)
	endTile := g.GetTileAt(toRow, toCol)

	if startTile == nil {
		return nil, fmt.Errorf("invalid start position: (%d, %d)", fromRow, fromCol)
	}
	if endTile == nil {
		return nil, fmt.Errorf("invalid end position: (%d, %d)", toRow, toCol)
	}

	// For now, return a simple direct path
	// TODO: Implement proper A* pathfinding
	path := []Tile{*startTile, *endTile}
	return path, nil
}

// IsValidMove checks if movement is legal
func (g *Game) IsValidMove(fromRow, fromCol, toRow, toCol int) bool {
	// Check if both positions are valid
	startTile := g.GetTileAt(fromRow, fromCol)
	endTile := g.GetTileAt(toRow, toCol)

	if startTile == nil || endTile == nil {
		return false
	}

	// Check if there's a unit at the start position
	if startTile.Unit == nil {
		return false
	}

	// Check if the unit belongs to the current player
	if startTile.Unit.PlayerID != g.CurrentPlayer {
		return false
	}

	// Check if destination is empty
	if endTile.Unit != nil {
		return false
	}

	// Check if unit has movement left
	if startTile.Unit.DistanceLeft <= 0 {
		return false
	}

	// For now, allow movement to any adjacent tile
	// TODO: Implement proper movement range and pathfinding validation
	return true
}

// GetMovementCost calculates movement points required
func (g *Game) GetMovementCost(fromRow, fromCol, toRow, toCol int) int {
	// For now, return a simple cost based on distance
	// TODO: Implement proper terrain-based movement costs
	if fromRow == toRow && fromCol == toCol {
		return 0
	}

	// Calculate hex distance (simplified)
	dRow := abs(toRow - fromRow)
	dCol := abs(toCol - fromCol)

	if dRow <= 1 && dCol <= 1 {
		return 1 // Adjacent tiles cost 1 movement point
	}

	return dRow + dCol // Simplified distance calculation
}

// =============================================================================
// UnitInterface Interface Implementation
// =============================================================================

// GetUnitAt returns unit at specific position
func (g *Game) GetUnitAt(row, col int) *Unit {
	tile := g.GetTileAt(row, col)
	if tile == nil {
		return nil
	}
	return tile.Unit
}

// GetUnitsForPlayer returns all units owned by player
func (g *Game) GetUnitsForPlayer(playerID int) []*Unit {
	if playerID < 0 || playerID >= len(g.Units) {
		return nil
	}

	// Return a copy to prevent external modification
	units := make([]*Unit, len(g.Units[playerID]))
	copy(units, g.Units[playerID])
	return units
}

// GetAllUnits returns every unit on the map
func (g *Game) GetAllUnits() []*Unit {
	var allUnits []*Unit

	for _, playerUnits := range g.Units {
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
	if g.assetManager != nil {
		// Try to get unit data from JSON if asset manager is loaded
		if err := g.assetManager.LoadGameData(); err == nil {
			if unitData, err := g.assetManager.GetUnitData(unitType); err == nil {
				return unitData.Name
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

// GetUnitMovementLeft returns remaining movement points
func (g *Game) GetUnitMovementLeft(unit *Unit) int {
	if unit == nil {
		return 0
	}
	return unit.DistanceLeft
}

// GetUnitAttackRange returns attack range in tiles
func (g *Game) GetUnitAttackRange(unit *Unit) int {
	if unit == nil {
		return 0
	}

	// For now, return a simple range based on unit type
	// TODO: Get from unit data
	switch unit.UnitType {
	case 1: // Infantry
		return 1
	case 2: // Artillery
		return 3
	case 3: // Tank
		return 1
	default:
		return 1
	}
}

// MoveUnit executes unit movement
func (g *Game) MoveUnit(unit *Unit, toRow, toCol int) error {
	if unit == nil {
		return fmt.Errorf("unit is nil")
	}

	// Check if it's the correct player's turn
	if unit.PlayerID != g.CurrentPlayer {
		return fmt.Errorf("not player %d's turn", unit.PlayerID)
	}

	// Check if move is valid
	if !g.IsValidMove(unit.Row, unit.Col, toRow, toCol) {
		return fmt.Errorf("invalid move from (%d,%d) to (%d,%d)", unit.Row, unit.Col, toRow, toCol)
	}

	// Get movement cost
	cost := g.GetMovementCost(unit.Row, unit.Col, toRow, toCol)
	if cost > unit.DistanceLeft {
		return fmt.Errorf("insufficient movement points: need %d, have %d", cost, unit.DistanceLeft)
	}

	// Store original position for event
	fromPos := Position{Row: unit.Row, Col: unit.Col}
	toPos := Position{Row: toRow, Col: toCol}

	// Remove unit from current tile
	currentTile := g.GetTileAt(unit.Row, unit.Col)
	if currentTile != nil {
		currentTile.Unit = nil
	}

	// Move unit to new position
	unit.Row = toRow
	unit.Col = toCol
	unit.DistanceLeft -= cost

	// Place unit on new tile
	newTile := g.GetTileAt(toRow, toCol)
	if newTile != nil {
		newTile.Unit = unit
	}

	// Update timestamp
	g.LastActionAt = time.Now()

	// Emit events
	g.eventManager.EmitUnitMoved(unit, fromPos, toPos)
	g.eventManager.EmitGameStateChanged(GameStateChangeUnitMoved, map[string]interface{}{
		"unit": unit,
		"from": fromPos,
		"to":   toPos,
	})

	return nil
}

// AttackUnit executes combat between units
func (g *Game) AttackUnit(attacker, defender *Unit) (*CombatResult, error) {
	if attacker == nil || defender == nil {
		return nil, fmt.Errorf("attacker or defender is nil")
	}

	// Check if it's the correct player's turn
	if attacker.PlayerID != g.CurrentPlayer {
		return nil, fmt.Errorf("not player %d's turn", attacker.PlayerID)
	}

	// Check if units can attack each other
	if !g.CanAttackUnit(attacker, defender) {
		return nil, fmt.Errorf("attacker cannot attack defender")
	}

	// Calculate damage (simplified combat)
	attackerDamage := 0
	defenderDamage := g.calculateDamage(attacker, defender)

	// Apply damage
	defender.AvailableHealth -= defenderDamage
	if defender.AvailableHealth < 0 {
		defender.AvailableHealth = 0
	}

	// Check if defender was killed
	defenderKilled := defender.AvailableHealth <= 0

	// Remove defender if killed
	if defenderKilled {
		g.RemoveUnit(defender)
	}

	// Create combat result
	result := &CombatResult{
		AttackerDamage: attackerDamage,
		DefenderDamage: defenderDamage,
		AttackerKilled: false,
		DefenderKilled: defenderKilled,
		AttackerHealth: attacker.AvailableHealth,
		DefenderHealth: defender.AvailableHealth,
	}

	// Update timestamp
	g.LastActionAt = time.Now()

	// Emit events
	g.eventManager.EmitUnitAttacked(attacker, defender, result)
	g.eventManager.EmitGameStateChanged(GameStateChangeUnitAttacked, map[string]interface{}{
		"attacker": attacker,
		"defender": defender,
		"result":   result,
	})

	return result, nil
}

// CanMoveUnit validates potential movement
func (g *Game) CanMoveUnit(unit *Unit, toRow, toCol int) bool {
	if unit == nil {
		return false
	}

	// Check if it's the correct player's turn
	if unit.PlayerID != g.CurrentPlayer {
		return false
	}

	// Check if move is valid
	return g.IsValidMove(unit.Row, unit.Col, toRow, toCol)
}

// CanAttackUnit validates potential attack
func (g *Game) CanAttackUnit(attacker, defender *Unit) bool {
	if attacker == nil || defender == nil {
		return false
	}

	// Check if it's the correct player's turn
	if attacker.PlayerID != g.CurrentPlayer {
		return false
	}

	// Check if units are enemies
	if attacker.PlayerID == defender.PlayerID {
		return false
	}

	// Check if attacker is in range
	distance := g.calculateDistance(attacker.Row, attacker.Col, defender.Row, defender.Col)
	attackRange := g.GetUnitAttackRange(attacker)

	return distance <= attackRange
}

// CanAttack validates potential attack using position coordinates
func (g *Game) CanAttack(fromRow, fromCol, toRow, toCol int) (bool, error) {
	attacker := g.GetUnitAt(fromRow, fromCol)
	if attacker == nil {
		return false, fmt.Errorf("no unit at attacker position (%d, %d)", fromRow, fromCol)
	}

	defender := g.GetUnitAt(toRow, toCol)
	if defender == nil {
		return false, fmt.Errorf("no unit at target position (%d, %d)", toRow, toCol)
	}

	return g.CanAttackUnit(attacker, defender), nil
}

// CanMove validates potential movement using position coordinates
func (g *Game) CanMove(fromRow, fromCol, toRow, toCol int) (bool, error) {
	unit := g.GetUnitAt(fromRow, fromCol)
	if unit == nil {
		return false, fmt.Errorf("no unit at position (%d, %d)", fromRow, fromCol)
	}

	return g.CanMoveUnit(unit, toRow, toCol), nil
}

// CreateUnit spawns new unit
func (g *Game) CreateUnit(unitType, playerID, row, col int) (*Unit, error) {
	// Validate parameters
	if playerID < 0 || playerID >= g.PlayerCount {
		return nil, fmt.Errorf("invalid player ID: %d", playerID)
	}

	// Check if position is valid and empty
	tile := g.GetTileAt(row, col)
	if tile == nil {
		return nil, fmt.Errorf("invalid position: (%d, %d)", row, col)
	}

	if tile.Unit != nil {
		return nil, fmt.Errorf("position (%d, %d) is occupied", row, col)
	}

	// Create the unit
	unit := NewUnit(unitType, playerID)
	unit.Row = row
	unit.Col = col
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

	// Remove from tile
	tile := g.GetTileAt(unit.Row, unit.Col)
	if tile != nil && tile.Unit == unit {
		tile.Unit = nil
	}

	// Remove from player's unit list
	if unit.PlayerID >= 0 && unit.PlayerID < len(g.Units) {
		playerUnits := g.Units[unit.PlayerID]
		for i, u := range playerUnits {
			if u == unit {
				// Remove from slice
				g.Units[unit.PlayerID] = append(playerUnits[:i], playerUnits[i+1:]...)
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

	if playerID < 0 || playerID >= len(g.Units) {
		return fmt.Errorf("invalid player ID: %d", playerID)
	}

	// Set unit's player ID
	unit.PlayerID = playerID

	// Add to player's unit list
	g.Units[playerID] = append(g.Units[playerID], unit)

	// Place unit on the map if it has a valid position
	if tile := g.Map.TileAt(unit.Row, unit.Col); tile != nil {
		tile.Unit = unit
	}

	return nil
}

// calculateDamage calculates damage dealt in combat (simplified)
func (g *Game) calculateDamage(attacker, defender *Unit) int {
	// Simplified damage calculation
	// TODO: Implement proper damage calculation based on unit types, terrain, etc.

	baseDamage := 30

	// Add some randomness
	variation := g.rng.Intn(20) - 10 // -10 to +10
	damage := baseDamage + variation

	if damage < 10 {
		damage = 10 // Minimum damage
	}

	return damage
}

// calculateDistance calculates distance between two positions
func (g *Game) calculateDistance(row1, col1, row2, col2 int) int {
	// Simplified hex distance calculation
	// TODO: Implement proper hex distance calculation
	dRow := abs(row2 - row1)
	dCol := abs(col2 - col1)

	return max(dRow, dCol)
}

// =============================================================================
// Rendering Integration
// =============================================================================

// RenderToBuffer renders the complete game state to a buffer
func (g *Game) RenderToBuffer(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) error {
	if buffer == nil {
		return fmt.Errorf("buffer is nil")
	}

	if g.Map == nil {
		return fmt.Errorf("no map loaded")
	}

	// Clear buffer first
	buffer.Clear()

	// Render terrain layer
	g.RenderTerrain(buffer, tileWidth, tileHeight, yIncrement)

	// Render units layer
	g.RenderUnits(buffer, tileWidth, tileHeight, yIncrement)

	// Render UI layer
	g.RenderUI(buffer, tileWidth, tileHeight, yIncrement)

	return nil
}

// RenderTerrain renders the terrain tiles to a buffer
func (g *Game) RenderTerrain(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	if g.Map == nil {
		return
	}

	// Render terrain tiles
	for _, tile := range g.Map.Tiles {
		if tile != nil {
			// Calculate tile position
			x, y := g.Map.XYForTile(tile.Row, tile.Col, tileWidth, tileHeight, yIncrement)

				// Try to load real tile asset first
				if g.assetManager != nil && g.assetManager.HasTileAsset(tile.TileType) {
					if tileImg, err := g.assetManager.GetTileImage(tile.TileType); err == nil {
						// Render real tile image (XYForTile already returns centered coordinates)
						buffer.DrawImage(x-tileWidth/2, y-tileHeight/2, tileWidth, tileHeight, tileImg)
						continue
					}
				}

				// Fallback to colored hexagon if asset not available
				hexPath := g.createHexagonPath(x, y, tileWidth, tileHeight, yIncrement)
				tileColor := g.getTerrainColor(tile.TileType)
				buffer.FillPath(hexPath, tileColor)

				// Add border
				borderColor := Color{R: 100, G: 100, B: 100, A: 100}
				strokeProps := StrokeProperties{Width: 1.0, LineCap: "round", LineJoin: "round"}
				buffer.StrokePath(hexPath, borderColor, strokeProps)
			}
		}
	}

// RenderUnits renders the units to a buffer
func (g *Game) RenderUnits(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	if g.Map == nil {
		return
	}

	// Define colors for different players
	playerColors := []Color{
		{R: 255, G: 0, B: 0, A: 255},   // Player 0 - red
		{R: 0, G: 0, B: 255, A: 255},   // Player 1 - blue
		{R: 0, G: 255, B: 0, A: 255},   // Player 2 - green
		{R: 255, G: 255, B: 0, A: 255}, // Player 3 - yellow
		{R: 255, G: 0, B: 255, A: 255}, // Player 4 - magenta
		{R: 0, G: 255, B: 255, A: 255}, // Player 5 - cyan
	}

	// Render units for each player
	for playerID, units := range g.Units {
		for _, unit := range units {
			if unit != nil {
				// Calculate unit position (same as tile position)
				x, y := g.Map.XYForTile(unit.Row, unit.Col, tileWidth, tileHeight, yIncrement)

				// Try to load real unit sprite first
				if g.assetManager != nil && g.assetManager.HasUnitAsset(unit.UnitType, playerID) {
					if unitImg, err := g.assetManager.GetUnitImage(unit.UnitType, playerID); err == nil {
						// Render real unit sprite (XYForTile already returns centered coordinates)
						buffer.DrawImage(x-tileWidth/2, y-tileHeight/2, tileWidth, tileHeight, unitImg)

						// Add health indicator if unit is damaged
						if unit.AvailableHealth < 100 {
							g.renderHealthBar(buffer, x, y, tileWidth, tileHeight, unit.AvailableHealth, 100)
						}

						// Add unit ID and health text overlay
						g.renderUnitText(buffer, unit, x, y, tileWidth, tileHeight)
						continue
					}
				}

				// Fallback to colored circle if asset not available
				var unitColor Color
				if playerID < len(playerColors) {
					unitColor = playerColors[playerID]
				} else {
					unitColor = Color{R: 128, G: 128, B: 128, A: 255} // Default gray
				}

				// Create unit representation (circle)
				unitPath := g.createUnitCircle(x, y, tileWidth, tileHeight)

				// Fill unit with player color
				buffer.FillPath(unitPath, unitColor)

				// Add unit border
				borderColor := Color{R: 0, G: 0, B: 0, A: 255}
				strokeProps := StrokeProperties{Width: 2.0, LineCap: "round", LineJoin: "round"}
				buffer.StrokePath(unitPath, borderColor, strokeProps)

				// Add health indicator if unit is damaged
				if unit.AvailableHealth < 100 {
					g.renderHealthBar(buffer, x, y, tileWidth, tileHeight, unit.AvailableHealth, 100)
				}

				// Add unit ID and health text overlay
				g.renderUnitText(buffer, unit, x, y, tileWidth, tileHeight)
			}
		}
	}
}

// RenderUI renders UI elements to a buffer
func (g *Game) RenderUI(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	// Create a simple indicator for current player
	indicatorSize := 20.0

	// Get current player color
	playerColors := []Color{
		{R: 255, G: 0, B: 0, A: 200},   // Player 0 - red
		{R: 0, G: 0, B: 255, A: 200},   // Player 1 - blue
		{R: 0, G: 255, B: 0, A: 200},   // Player 2 - green
		{R: 255, G: 255, B: 0, A: 200}, // Player 3 - yellow
		{R: 255, G: 0, B: 255, A: 200}, // Player 4 - magenta
		{R: 0, G: 255, B: 255, A: 200}, // Player 5 - cyan
	}

	var currentPlayerColor Color
	if g.CurrentPlayer < len(playerColors) {
		currentPlayerColor = playerColors[g.CurrentPlayer]
	} else {
		currentPlayerColor = Color{R: 255, G: 255, B: 255, A: 200} // Default white
	}

	// Create indicator rectangle in top-left corner
	indicatorPath := []Point{
		{X: 5, Y: 5},
		{X: 5 + indicatorSize, Y: 5},
		{X: 5 + indicatorSize, Y: 5 + indicatorSize},
		{X: 5, Y: 5 + indicatorSize},
	}

	// Fill indicator with current player color
	buffer.FillPath(indicatorPath, currentPlayerColor)

	// Add border
	borderColor := Color{R: 0, G: 0, B: 0, A: 255}
	strokeProps := StrokeProperties{Width: 2.0, LineCap: "round", LineJoin: "round"}
	buffer.StrokePath(indicatorPath, borderColor, strokeProps)
}

// createHexagonPath creates a hexagon path for a tile
func (g *Game) createHexagonPath(x, y, tileWidth, tileHeight, yIncrement float64) []Point {
	return []Point{
		{X: x + tileWidth/2, Y: y},                         // Top
		{X: x + tileWidth, Y: y + tileHeight - yIncrement}, // Top-right
		{X: x + tileWidth, Y: y + yIncrement},              // Bottom-right
		{X: x + tileWidth/2, Y: y + tileHeight},            // Bottom
		{X: x, Y: y + yIncrement},                          // Bottom-left
		{X: x, Y: y + tileHeight - yIncrement},             // Top-left
	}
}

// createUnitCircle creates a circular path for a unit
func (g *Game) createUnitCircle(x, y, tileWidth, tileHeight float64) []Point {
	// Create a circular approximation using polygon
	centerX := x + tileWidth/2
	centerY := y + tileHeight/2
	radius := minFloat(tileWidth, tileHeight) * 0.3 // Unit size relative to tile

	segments := 12
	points := make([]Point, segments)

	for i := 0; i < segments; i++ {
		angle := 2 * 3.14159 * float64(i) / float64(segments)
		unitX := centerX + radius*approximateCos(angle)
		unitY := centerY + radius*approximateSin(angle)
		points[i] = Point{X: unitX, Y: unitY}
	}

	return points
}

// renderHealthBar renders a health bar for a unit
func (g *Game) renderHealthBar(buffer *Buffer, x, y, tileWidth, tileHeight float64, currentHealth, maxHealth int) {
	if currentHealth >= maxHealth {
		return // Don't render health bar for full health
	}

	// Calculate health bar dimensions
	barWidth := tileWidth * 0.8
	barHeight := 6.0
	barX := x + (tileWidth-barWidth)/2
	barY := y + tileHeight - barHeight - 2

	// Background bar (red)
	backgroundBar := []Point{
		{X: barX, Y: barY},
		{X: barX + barWidth, Y: barY},
		{X: barX + barWidth, Y: barY + barHeight},
		{X: barX, Y: barY + barHeight},
	}
	buffer.FillPath(backgroundBar, Color{R: 255, G: 0, B: 0, A: 200})

	// Health bar (green)
	healthPercent := float64(currentHealth) / float64(maxHealth)
	healthBarWidth := barWidth * healthPercent

	if healthBarWidth > 0 {
		healthBar := []Point{
			{X: barX, Y: barY},
			{X: barX + healthBarWidth, Y: barY},
			{X: barX + healthBarWidth, Y: barY + barHeight},
			{X: barX, Y: barY + barHeight},
		}
		buffer.FillPath(healthBar, Color{R: 0, G: 255, B: 0, A: 200})
	}
}

// renderUnitText renders unit ID and health text overlay on PNG output
func (g *Game) renderUnitText(buffer *Buffer, unit *Unit, x, y, tileWidth, tileHeight float64) {
	// Get unit ID
	unitID := g.GetUnitID(unit)

	// Render unit ID below the unit with bold font and dark background
	idTextColor := Color{R: 255, G: 255, B: 255, A: 255}    // White text for visibility
	idBackgroundColor := Color{R: 0, G: 0, B: 0, A: 180}    // Semi-transparent black background
	idFontSize := 28.0                                       // Large font size for readability
	idX := x - 15                                            // Slightly left of center
	idY := y + (tileHeight * 0.4)                            // Below the unit
	buffer.DrawTextWithStyle(idX, idY, unitID, idFontSize, idTextColor, true, idBackgroundColor)

	// Render health with bold font and dark background (upper right)
	healthText := fmt.Sprintf("%d", unit.AvailableHealth)
	healthTextColor := Color{R: 255, G: 255, B: 0, A: 255}     // Yellow text for better visibility
	healthBackgroundColor := Color{R: 0, G: 0, B: 0, A: 180}   // Semi-transparent black background
	healthFontSize := 22.0                                      // Large font for health
	healthX := x + 15                                           // Upper right area
	healthY := y - (tileHeight * 0.3)                           // Above center
	buffer.DrawTextWithStyle(healthX, healthY, healthText, healthFontSize, healthTextColor, true, healthBackgroundColor)
}

// getTerrainColor returns color for terrain type
func (g *Game) getTerrainColor(terrainType int) Color {
	switch terrainType {
	case 1: // Grass
		return Color{R: 50, G: 150, B: 50, A: 255}
	case 2: // Desert
		return Color{R: 200, G: 180, B: 100, A: 255}
	case 3: // Water
		return Color{R: 50, G: 50, B: 200, A: 255}
	case 4: // Mountain
		return Color{R: 150, G: 100, B: 50, A: 255}
	case 5: // Rock
		return Color{R: 150, G: 150, B: 150, A: 255}
	default: // Unknown
		return Color{R: 200, G: 200, B: 200, A: 255}
	}
}

// Helper math functions
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func approximateCos(angle float64) float64 {
	// Simple approximation - in a real implementation, use math.Cos
	return 1.0 - angle*angle/2.0 + angle*angle*angle*angle/24.0
}

func approximateSin(angle float64) float64 {
	// Simple approximation - in a real implementation, use math.Sin
	return angle - angle*angle*angle/6.0 + angle*angle*angle*angle*angle/120.0
}

// =============================================================================
// Helper Functions
// =============================================================================

// createTestMap creates a simple test map for development
func createTestMap(mapName string) (*Map, error) {
	// Create a small test map
	gameMap := NewMap(8, 12, false) // 8 rows, 12 columns, odd rows offset

	// Add some test tiles
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			// Create varied terrain
			tileType := 1 // Default to grass
			if (row+col)%4 == 0 {
				tileType = 2 // Some desert
			} else if (row+col)%7 == 0 {
				tileType = 3 // Some water
			}

			tile := NewTile(row, col, tileType)
			gameMap.AddTile(tile)
		}
	}

	// Note: Neighbor connections calculated on-demand via GetNeighbor()

	return gameMap, nil
}

// initializeStartingUnits adds initial units to the game
func (g *Game) initializeStartingUnits() error {
	// Add some basic starting units for each player
	startingPositions := [][]Position{
		{{Row: 1, Col: 1}, {Row: 1, Col: 2}},  // Player 0
		{{Row: 6, Col: 9}, {Row: 6, Col: 10}}, // Player 1
	}

	for playerID := 0; playerID < g.PlayerCount && playerID < len(startingPositions); playerID++ {
		positions := startingPositions[playerID]

		for _, pos := range positions {
			// Create a basic infantry unit
			unit := NewUnit(1, playerID) // Unit type 1 = Infantry
			unit.Row = pos.Row
			unit.Col = pos.Col
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
	if playerID < 0 || playerID >= len(g.Units) {
		return fmt.Errorf("invalid player ID: %d", playerID)
	}

	for _, unit := range g.Units[playerID] {
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

	for playerID := 0; playerID < g.PlayerCount; playerID++ {
		if len(g.Units[playerID]) > 0 {
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
	if g.Map == nil {
		return fmt.Errorf("game has no map")
	}

	if g.PlayerCount < 2 || g.PlayerCount > 6 {
		return fmt.Errorf("invalid player count: %d", g.PlayerCount)
	}

	if g.CurrentPlayer < 0 || g.CurrentPlayer >= g.PlayerCount {
		return fmt.Errorf("invalid current player: %d", g.CurrentPlayer)
	}

	if g.TurnCounter < 1 {
		return fmt.Errorf("invalid turn counter: %d", g.TurnCounter)
	}

	if len(g.Units) != g.PlayerCount {
		return fmt.Errorf("units array length (%d) doesn't match player count (%d)", len(g.Units), g.PlayerCount)
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
	for _, playerUnits := range g.Units {
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
