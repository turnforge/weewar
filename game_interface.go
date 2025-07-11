package weewar

// =============================================================================
// WeeWar Core Game Interface Definitions
// =============================================================================
// This file defines the core game interface contracts for the WeeWar game system.
// These interfaces focus purely on game mechanics: game state, map operations,
// unit management, and core events.

// =============================================================================
// Core Data Types
// =============================================================================

// GameStatus represents the current state of the game
type GameStatus int

const (
	GameStatusPlaying GameStatus = iota
	GameStatusPaused
	GameStatusEnded
)

func (gs GameStatus) String() string {
	switch gs {
	case GameStatusPlaying:
		return "playing"
	case GameStatusPaused:
		return "paused"
	case GameStatusEnded:
		return "ended"
	default:
		return "unknown"
	}
}

// Position represents a coordinate position (row, col)
type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// CombatResult represents the outcome of a combat action
type CombatResult struct {
	AttackerDamage int  `json:"attackerDamage"` // Damage dealt to attacker
	DefenderDamage int  `json:"defenderDamage"` // Damage dealt to defender
	AttackerKilled bool `json:"attackerKilled"` // Whether attacker was destroyed
	DefenderKilled bool `json:"defenderKilled"` // Whether defender was destroyed
	AttackerHealth int  `json:"attackerHealth"` // Attacker's health after combat
	DefenderHealth int  `json:"defenderHealth"` // Defender's health after combat
}

// =============================================================================
// Core Game Management Interface
// =============================================================================

// GameController manages game lifecycle and core state
type GameController interface {
	// Game Lifecycle
	// LoadGame restores game from saved state
	// Called by: CLI load command, Browser load button, Test fixtures
	// Returns: Restored game instance
	LoadGame(saveData []byte) (*Game, error)
	
	// SaveGame serializes current game state
	// Called by: CLI save command, Browser save button, Auto-save system
	// Returns: Serialized game data for storage
	SaveGame() ([]byte, error)
	
	// Game State Queries
	// GetCurrentPlayer returns active player ID
	// Called by: UI updates, Turn indicators, AI systems, CLI prompts
	// Returns: Current player index (0-based)
	GetCurrentPlayer() int
	
	// GetTurnNumber returns current turn count
	// Called by: UI displays, Game statistics, Victory conditions
	// Returns: Current turn number (1-based)
	GetTurnNumber() int
	
	// GetGameStatus returns current game state
	// Called by: UI updates, Game loop control, Victory checks
	// Returns: GameStatus enum (playing, paused, ended)
	GetGameStatus() GameStatus
	
	// GetWinner returns winning player if game ended
	// Called by: Victory screens, Statistics, Game result logging
	// Returns: Winner player ID and whether game has ended
	GetWinner() (int, bool)
	
	// Turn Management
	// NextTurn advances to next player's turn
	// Called by: End turn button, AI turn completion, CLI turn command
	// Returns: Error if turn cannot advance
	NextTurn() error
	
	// EndTurn completes current player's turn
	// Called by: UI end turn button, AI decision completion
	// Returns: Error if turn cannot end
	EndTurn() error
	
	// CanEndTurn checks if current player can end their turn
	// Called by: UI button enabling, Turn validation, AI decision logic
	// Returns: True if turn can be ended
	CanEndTurn() bool
}

// =============================================================================
// Map and Position Interface
// =============================================================================

// MapInterface provides map queries and coordinate operations
type MapInterface interface {
	// Map Properties
	// GetMapSize returns map dimensions
	// Called by: Rendering systems, Bounds checking, UI layout
	// Returns: Row and column count
	GetMapSize() (rows, cols int)
	
	// GetMapName returns loaded map name
	// Called by: UI display, Save file naming, Statistics
	// Returns: Human-readable map name
	GetMapName() string
	
	// GetMapBounds returns pixel boundaries for rendering
	// Called by: Canvas sizing, Camera positioning, Zoom calculations
	// Returns: Minimum and maximum x,y coordinates
	GetMapBounds() (minX, minY, maxX, maxY float64)
	
	// Tile Queries
	// GetTileAt returns tile at specific position
	// Called by: Click handling, Pathfinding, Unit placement validation
	// Returns: Tile instance or nil if invalid position
	GetTileAt(row, col int) *Tile
	
	// GetTileType returns terrain type at position
	// Called by: Movement cost calculation, Rendering, Combat bonuses
	// Returns: Terrain type ID
	GetTileType(row, col int) int
	
	// GetTileNeighbors returns adjacent tiles (hex grid)
	// Called by: Pathfinding algorithms, Unit placement, Combat range
	// Returns: Array of 6 neighboring tiles (some may be nil)
	GetTileNeighbors(row, col int) []*Tile
	
	// Coordinate Conversion
	// RowColToPixel converts grid coordinates to screen coordinates
	// Called by: Rendering systems, UI positioning, Animation systems
	// Returns: Pixel coordinates for rendering
	RowColToPixel(row, col int) (x, y float64)
	
	// PixelToRowCol converts screen coordinates to grid coordinates
	// Called by: Click handling, Mouse hover, Touch input
	// Returns: Grid coordinates and validity flag
	PixelToRowCol(x, y float64) (row, col int, valid bool)
	
	// Path Finding
	// FindPath calculates movement path between positions
	// Called by: Unit movement, AI pathfinding, Movement preview
	// Returns: Sequence of tiles for movement path
	FindPath(fromRow, fromCol, toRow, toCol int) ([]Tile, error)
	
	// IsValidMove checks if movement is legal
	// Called by: Input validation, AI move filtering, UI feedback
	// Returns: True if move is valid
	IsValidMove(fromRow, fromCol, toRow, toCol int) bool
	
	// GetMovementCost calculates movement points required
	// Called by: Movement validation, AI cost analysis, UI display
	// Returns: Movement points required
	GetMovementCost(fromRow, fromCol, toRow, toCol int) int
}

// =============================================================================
// Unit Management Interface
// =============================================================================

// UnitInterface provides unit queries and actions
type UnitInterface interface {
	// Unit Queries
	// GetUnitAt returns unit at specific position
	// Called by: Click handling, Combat target selection, Rendering
	// Returns: Unit instance or nil if no unit present
	GetUnitAt(row, col int) *Unit
	
	// GetUnitsForPlayer returns all units owned by player
	// Called by: Turn processing, Victory conditions, AI analysis
	// Returns: Array of units owned by specified player
	GetUnitsForPlayer(playerID int) []*Unit
	
	// GetAllUnits returns every unit on the map
	// Called by: Rendering systems, Game statistics, Victory checks
	// Returns: Array of all units regardless of ownership
	GetAllUnits() []*Unit
	
	// Unit Properties
	// GetUnitType returns unit type identifier
	// Called by: Combat calculations, Rendering, Ability checks
	// Returns: Unit type ID referencing UnitData
	GetUnitType(unit *Unit) int
	
	// GetUnitHealth returns current health points
	// Called by: Combat calculations, UI display, Victory conditions
	// Returns: Current health value
	GetUnitHealth(unit *Unit) int
	
	// GetUnitMovementLeft returns remaining movement points
	// Called by: Movement validation, UI display, AI planning
	// Returns: Movement points remaining this turn
	GetUnitMovementLeft(unit *Unit) int
	
	// GetUnitAttackRange returns attack range in tiles
	// Called by: Combat target validation, AI targeting, UI highlighting
	// Returns: Attack range in hex tiles
	GetUnitAttackRange(unit *Unit) int
	
	// Unit Actions
	// MoveUnit executes unit movement
	// Called by: UI move confirmation, AI move execution, CLI move command
	// Returns: Error if movement fails
	MoveUnit(unit *Unit, toRow, toCol int) error
	
	// AttackUnit executes combat between units
	// Called by: UI attack confirmation, AI attack execution, CLI attack command
	// Returns: Combat result with damage/casualties
	AttackUnit(attacker, defender *Unit) (*CombatResult, error)
	
	// CanMoveUnit validates potential movement
	// Called by: UI button enabling, AI move filtering, Input validation
	// Returns: True if unit can move to specified position
	CanMoveUnit(unit *Unit, toRow, toCol int) bool
	
	// CanAttackUnit validates potential attack
	// Called by: UI button enabling, AI target filtering, Input validation
	// Returns: True if attacker can attack defender
	CanAttackUnit(attacker, defender *Unit) bool
	
	// Unit Creation/Removal
	// CreateUnit spawns new unit (base production, reinforcements)
	// Called by: Base production, Scenario setup, Cheat commands
	// Returns: Created unit instance
	CreateUnit(unitType, playerID, row, col int) (*Unit, error)
	
	// RemoveUnit removes unit from game (destruction, capture)
	// Called by: Combat resolution, Victory conditions, Scenario events
	// Returns: Error if removal fails
	RemoveUnit(unit *Unit) error
}



// =============================================================================
// Core Game Interface
// =============================================================================

// GameInterface combines core game interfaces into a single contract
// This is the main interface that the unified Game struct will implement
type GameInterface interface {
	GameController
	MapInterface
	UnitInterface
}