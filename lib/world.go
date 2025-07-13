package weewar

// =============================================================================
// World - Pure Game State Container
// =============================================================================

// World represents the pure game state without any rendering or UI concerns.
// This is the single source of truth for all game data.
type World struct {
	Map   *Map    `json:"map"`   // The game map with terrain and tiles
	Units []*Unit `json:"units"` // All units in the game world
	
	// Game metadata
	PlayerCount int `json:"playerCount"` // Number of players in the game
	Seed        int `json:"seed"`        // Random seed for reproducible games
	
	// Turn management
	CurrentPlayer int `json:"currentPlayer"` // Current player's turn (0-based)
	TurnNumber    int `json:"turnNumber"`    // Current turn number
}

// ViewState represents UI-specific state that doesn't affect game logic.
// This includes visual concerns like selections, highlights, and camera position.
type ViewState struct {
	// Selection and highlighting
	SelectedUnit      *Unit      `json:"selectedUnit"`      // Currently selected unit
	HoveredTile       *Tile      `json:"hoveredTile"`       // Tile under cursor
	MovableTiles      []Position `json:"movableTiles"`      // Highlighted movement tiles
	AttackableTiles   []Position `json:"attackableTiles"`   // Highlighted attack tiles
	
	// Visual settings
	ShowGrid        bool `json:"showGrid"`        // Whether to show hex grid lines
	ShowCoordinates bool `json:"showCoordinates"` // Whether to show coordinate labels
	ShowPaths       bool `json:"showPaths"`       // Whether to show movement paths
	
	// Camera and viewport
	CameraX   float64 `json:"cameraX"`   // Camera X position
	CameraY   float64 `json:"cameraY"`   // Camera Y position
	ZoomLevel float64 `json:"zoomLevel"` // Zoom level (1.0 = normal)
	
	// Editor-specific state
	BrushTerrain int `json:"brushTerrain"` // Current terrain type for painting
	BrushSize    int `json:"brushSize"`    // Brush radius (0 = single hex)
}

// Note: Position type is already defined in game_interface.go

// =============================================================================
// World Creation and Management
// =============================================================================

// NewWorld creates a new game world with the specified parameters
func NewWorld(playerCount int, gameMap *Map, seed int) *World {
	return &World{
		Map:           gameMap,
		Units:         make([]*Unit, 0),
		PlayerCount:   playerCount,
		Seed:          seed,
		CurrentPlayer: 0,
		TurnNumber:    1,
	}
}

// NewViewState creates a new view state with default settings
func NewViewState() *ViewState {
	return &ViewState{
		SelectedUnit:      nil,
		HoveredTile:       nil,
		MovableTiles:      make([]Position, 0),
		AttackableTiles:   make([]Position, 0),
		ShowGrid:          true,
		ShowCoordinates:   false,
		ShowPaths:         true,
		CameraX:           0.0,
		CameraY:           0.0,
		ZoomLevel:         1.0,
		BrushTerrain:      1, // Default to grass
		BrushSize:         0, // Single hex brush
	}
}

// =============================================================================
// World State Access Methods
// =============================================================================

// GetMapSize returns the dimensions of the world map
func (w *World) GetMapSize() (rows, cols int) {
	if w.Map == nil {
		return 0, 0
	}
	return w.Map.NumRows, w.Map.NumCols
}

// GetTileAt returns the tile at the specified display coordinates
func (w *World) GetTileAt(row, col int) *Tile {
	if w.Map == nil {
		return nil
	}
	return w.Map.TileAt(row, col)
}

// GetTileAtCube returns the tile at the specified cube coordinates
func (w *World) GetTileAtCube(coord CubeCoord) *Tile {
	if w.Map == nil {
		return nil
	}
	return w.Map.TileAtCube(coord)
}

// GetUnitsAt returns all units at the specified display coordinates
func (w *World) GetUnitsAt(row, col int) []*Unit {
	units := make([]*Unit, 0)
	for _, unit := range w.Units {
		if unit.Row == row && unit.Col == col {
			units = append(units, unit)
		}
	}
	return units
}

// GetUnitAt returns the first unit found at the specified coordinates (for single-unit-per-tile games)
func (w *World) GetUnitAt(row, col int) *Unit {
	units := w.GetUnitsAt(row, col)
	if len(units) > 0 {
		return units[0]
	}
	return nil
}

// GetPlayerUnits returns all units belonging to the specified player
func (w *World) GetPlayerUnits(playerID int) []*Unit {
	units := make([]*Unit, 0)
	for _, unit := range w.Units {
		if unit.PlayerID == playerID {
			units = append(units, unit)
		}
	}
	return units
}

// =============================================================================
// World State Mutation Methods
// =============================================================================

// SetTileType changes the terrain type of a tile at the specified coordinates
func (w *World) SetTileType(row, col, terrainType int) bool {
	if w.Map == nil {
		return false
	}
	
	// Get or create tile at position
	tile := w.Map.TileAt(row, col)
	if tile == nil {
		// Create new tile
		tile = NewTile(row, col, terrainType)
		w.Map.AddTile(tile)
	} else {
		// Update existing tile
		tile.TileType = terrainType
	}
	
	return true
}

// SetTileTypeCube changes the terrain type at the specified cube coordinates
func (w *World) SetTileTypeCube(coord CubeCoord, terrainType int) bool {
	if w.Map == nil {
		return false
	}
	
	// Convert to display coordinates for tile creation
	row, col := w.Map.HexToDisplay(coord)
	
	// Get or create tile at position
	tile := w.Map.TileAtCube(coord)
	if tile == nil {
		// Create new tile
		tile = NewTile(row, col, terrainType)
		w.Map.AddTileCube(coord, tile)
	} else {
		// Update existing tile
		tile.TileType = terrainType
	}
	
	return true
}

// AddUnit adds a new unit to the world at the specified position
func (w *World) AddUnit(unit *Unit) {
	w.Units = append(w.Units, unit)
}

// RemoveUnit removes a unit from the world
func (w *World) RemoveUnit(unit *Unit) bool {
	for i, u := range w.Units {
		if u == unit {
			// Remove unit from slice
			w.Units = append(w.Units[:i], w.Units[i+1:]...)
			return true
		}
	}
	return false
}

// MoveUnit moves a unit to a new position
func (w *World) MoveUnit(unit *Unit, newRow, newCol int) {
	unit.Row = newRow
	unit.Col = newCol
}

// =============================================================================
// World Validation and Utilities
// =============================================================================

// IsValidPosition checks if the given coordinates are within the world bounds
func (w *World) IsValidPosition(row, col int) bool {
	if w.Map == nil {
		return false
	}
	return row >= 0 && row < w.Map.NumRows && col >= 0 && col < w.Map.NumCols
}

// IsValidCubePosition checks if the given cube coordinates are within the world bounds
func (w *World) IsValidCubePosition(coord CubeCoord) bool {
	if w.Map == nil {
		return false
	}
	row, col := w.Map.HexToDisplay(coord)
	return w.IsValidPosition(row, col)
}

// GetWorldBounds returns the bounding box of the world in display coordinates
func (w *World) GetWorldBounds() (minRow, minCol, maxRow, maxCol int) {
	if w.Map == nil {
		return 0, 0, 0, 0
	}
	return 0, 0, w.Map.NumRows - 1, w.Map.NumCols - 1
}

// Clone creates a deep copy of the world state (useful for undo/redo systems)
func (w *World) Clone() *World {
	if w == nil {
		return nil
	}
	
	// Clone map
	var clonedMap *Map
	if w.Map != nil {
		clonedMap = NewMap(w.Map.NumRows, w.Map.NumCols, false)
		for coord, tile := range w.Map.Tiles {
			if tile != nil {
				newTile := &Tile{
					Row:      tile.Row,
					Col:      tile.Col,
					TileType: tile.TileType,
					Unit:     nil, // Units are cloned separately
				}
				clonedMap.AddTileCube(coord, newTile)
			}
		}
	}
	
	// Clone units
	clonedUnits := make([]*Unit, len(w.Units))
	for i, unit := range w.Units {
		if unit != nil {
			clonedUnits[i] = &Unit{
				UnitType:        unit.UnitType,
				DistanceLeft:    unit.DistanceLeft,
				AvailableHealth: unit.AvailableHealth,
				TurnCounter:     unit.TurnCounter,
				Row:             unit.Row,
				Col:             unit.Col,
				PlayerID:        unit.PlayerID,
			}
		}
	}
	
	return &World{
		Map:           clonedMap,
		Units:         clonedUnits,
		PlayerCount:   w.PlayerCount,
		Seed:          w.Seed,
		CurrentPlayer: w.CurrentPlayer,
		TurnNumber:    w.TurnNumber,
	}
}

// =============================================================================
// ViewState Management
// =============================================================================

// ClearSelection clears the current unit selection and highlights
func (vs *ViewState) ClearSelection() {
	vs.SelectedUnit = nil
	vs.MovableTiles = make([]Position, 0)
	vs.AttackableTiles = make([]Position, 0)
}

// SetSelection sets the selected unit and updates related highlights
func (vs *ViewState) SetSelection(unit *Unit, movableTiles, attackableTiles []Position) {
	vs.SelectedUnit = unit
	vs.MovableTiles = movableTiles
	vs.AttackableTiles = attackableTiles
}

// SetCamera updates the camera position and zoom
func (vs *ViewState) SetCamera(x, y, zoom float64) {
	vs.CameraX = x
	vs.CameraY = y
	vs.ZoomLevel = zoom
}

// SetBrush updates the brush settings for terrain editing
func (vs *ViewState) SetBrush(terrainType, brushSize int) {
	vs.BrushTerrain = terrainType
	vs.BrushSize = brushSize
}