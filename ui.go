package weewar

// =============================================================================
// WeeWar UI Interface Definitions
// =============================================================================
// This file defines UI-specific interfaces and types for the WeeWar game system.
// These interfaces focus on browser interaction, rendering, input handling,
// and visual feedback for web-based gameplay.

// =============================================================================
// UI Data Types
// =============================================================================

// ClickResult represents the result of a click interaction
type ClickResult struct {
	Type        string   `json:"type"`        // "tile", "unit", "ui", "empty"
	Position    Position `json:"position"`    // Grid position clicked
	Unit        *Unit    `json:"unit"`        // Unit at position (if any)
	Tile        *Tile    `json:"tile"`        // Tile at position (if any)
	Action      string   `json:"action"`      // Suggested action: "select", "move", "attack", "info"
	Valid       bool     `json:"valid"`       // Whether click resulted in valid action
	Message     string   `json:"message"`     // Human-readable result message
}

// HoverResult represents the result of a hover interaction
type HoverResult struct {
	Type        string   `json:"type"`        // "tile", "unit", "ui", "empty"
	Position    Position `json:"position"`    // Grid position hovered
	Unit        *Unit    `json:"unit"`        // Unit at position (if any)
	Tile        *Tile    `json:"tile"`        // Tile at position (if any)
	Info        string   `json:"info"`        // Information to display
	Tooltip     string   `json:"tooltip"`     // Tooltip text
	Cursor      string   `json:"cursor"`      // Suggested cursor style
}

// UIState represents the current UI state
type UIState struct {
	SelectedUnit  *Unit      `json:"selectedUnit"`  // Currently selected unit
	HoveredTile   *Tile      `json:"hoveredTile"`   // Currently hovered tile
	ValidMoves    []Position `json:"validMoves"`    // Valid moves for selected unit
	AttackRanges  []Position `json:"attackRanges"`  // Attack ranges for selected unit
	ShowGrid      bool       `json:"showGrid"`      // Whether to show hex grid
	ShowPaths     bool       `json:"showPaths"`     // Whether to show movement paths
	CameraX       float64    `json:"cameraX"`       // Camera X position
	CameraY       float64    `json:"cameraY"`       // Camera Y position
	ZoomLevel     float64    `json:"zoomLevel"`     // Zoom level (1.0 = normal)
}

// RenderOptions represents rendering configuration
type RenderOptions struct {
	Width         int     `json:"width"`         // Canvas width in pixels
	Height        int     `json:"height"`        // Canvas height in pixels
	TileSize      float64 `json:"tileSize"`      // Size of each hex tile
	ShowGrid      bool    `json:"showGrid"`      // Whether to show hex grid lines
	ShowCoords    bool    `json:"showCoords"`    // Whether to show coordinate labels
	ShowPaths     bool    `json:"showPaths"`     // Whether to show movement paths
	HighlightMode string  `json:"highlightMode"` // "selection", "movement", "attack", "none"
}

// AnimationState represents current animation state
type AnimationState struct {
	IsAnimating     bool     `json:"isAnimating"`     // Whether animation is in progress
	AnimationType   string   `json:"animationType"`   // "movement", "attack", "explosion"
	AnimationUnit   *Unit    `json:"animationUnit"`   // Unit being animated
	StartPosition   Position `json:"startPosition"`   // Animation start position
	EndPosition     Position `json:"endPosition"`     // Animation end position
	Progress        float64  `json:"progress"`        // Animation progress (0-1)
	Duration        float64  `json:"duration"`        // Animation duration in seconds
	StartTime       float64  `json:"startTime"`       // Animation start time
}

// =============================================================================
// UI Interface
// =============================================================================

// UIInterface provides browser-specific functionality
type UIInterface interface {
	// Rendering
	// RenderGame generates PNG image of current game state
	// Called by: Canvas update, Screenshot capture, Turn replay
	// Returns: PNG image data as byte array
	RenderGame(options RenderOptions) []byte
	
	// RenderGameToCanvas renders directly to HTML5 canvas
	// Called by: Browser animation loops, Real-time updates
	// Returns: Error if rendering fails
	RenderGameToCanvas(canvasID string, options RenderOptions) error
	
	// Input Handling
	// HandleClick processes mouse/touch click events
	// Called by: Browser click event handlers, Touch event handlers
	// Returns: Click result describing what was clicked
	HandleClick(x, y float64) (*ClickResult, error)
	
	// HandleHover processes mouse hover events
	// Called by: Browser mousemove event handlers
	// Returns: Hover result for UI feedback
	HandleHover(x, y float64) (*HoverResult, error)
	
	// HandleKeyboard processes keyboard input
	// Called by: Browser keyboard event handlers
	// Returns: Whether key was handled
	HandleKeyboard(key string, pressed bool) bool
	
	// Visual Feedback
	// ShowValidMoves returns positions where unit can move
	// Called by: Unit selection, Movement preview, UI highlighting
	// Returns: Array of valid movement positions
	ShowValidMoves(unit *Unit) []Position
	
	// ShowAttackRange returns positions unit can attack
	// Called by: Unit selection, Attack preview, UI highlighting
	// Returns: Array of positions within attack range
	ShowAttackRange(unit *Unit) []Position
	
	// ShowPath returns movement path between positions
	// Called by: Movement preview, Path visualization, UI animation
	// Returns: Array of positions forming movement path
	ShowPath(from, to Position) []Position
	
	// UI State Management
	// GetSelectedUnit returns currently selected unit
	// Called by: UI updates, Action button enabling, Info panels
	// Returns: Selected unit or nil
	GetSelectedUnit() *Unit
	
	// SetSelectedUnit changes current selection
	// Called by: Click handlers, Keyboard shortcuts, AI demonstration
	SetSelectedUnit(unit *Unit)
	
	// GetHoveredTile returns tile under cursor
	// Called by: Tooltip display, Preview systems, UI feedback
	// Returns: Hovered tile or nil
	GetHoveredTile() *Tile
	
	// SetHoveredTile changes hover state
	// Called by: Mouse move handlers, Touch event handlers
	SetHoveredTile(tile *Tile)
	
	// GetUIState returns current UI state
	// Called by: State serialization, UI synchronization
	// Returns: Complete UI state
	GetUIState() UIState
	
	// SetUIState updates UI state
	// Called by: State restoration, UI synchronization
	SetUIState(state UIState)
	
	// Camera and Viewport
	// SetCamera sets camera position and zoom
	// Called by: Camera controls, Viewport management
	SetCamera(x, y float64, zoom float64)
	
	// GetCamera returns current camera settings
	// Called by: Viewport calculations, State persistence
	GetCamera() (x, y float64, zoom float64)
	
	// ScreenToWorld converts screen coordinates to world coordinates
	// Called by: Input handling, Coordinate conversion
	ScreenToWorld(screenX, screenY float64) (worldX, worldY float64)
	
	// WorldToScreen converts world coordinates to screen coordinates
	// Called by: Rendering, UI positioning
	WorldToScreen(worldX, worldY float64) (screenX, screenY float64)
}

// =============================================================================
// Animation Interface
// =============================================================================

// AnimationInterface provides animation capabilities
type AnimationInterface interface {
	// Animation Control
	// StartAnimation begins a new animation
	// Called by: Move/attack actions, Visual effects
	StartAnimation(animationType string, unit *Unit, from, to Position, duration float64)
	
	// UpdateAnimation updates animation progress
	// Called by: Animation loop, Frame updates
	// Returns: Whether animation is still active
	UpdateAnimation(deltaTime float64) bool
	
	// StopAnimation halts current animation
	// Called by: User input, Animation completion
	StopAnimation()
	
	// GetAnimationState returns current animation state
	// Called by: Rendering system, UI updates
	GetAnimationState() AnimationState
	
	// IsAnimating returns whether animation is in progress
	// Called by: Input handling, State management
	IsAnimating() bool
	
	// Animation Types
	// AnimateUnitMovement animates unit movement
	// Called by: Move command execution
	AnimateUnitMovement(unit *Unit, path []Position, duration float64)
	
	// AnimateUnitAttack animates combat action
	// Called by: Attack command execution
	AnimateUnitAttack(attacker, defender *Unit, result *CombatResult, duration float64)
	
	// AnimateExplosion animates explosion effect
	// Called by: Unit destruction, Special effects
	AnimateExplosion(position Position, intensity float64, duration float64)
}

// =============================================================================
// Theme Interface
// =============================================================================

// ThemeInterface provides visual theming capabilities
type ThemeInterface interface {
	// Theme Management
	// SetTheme changes current visual theme
	// Called by: User preferences, Game settings
	SetTheme(themeName string) error
	
	// GetTheme returns current theme name
	// Called by: Settings display, Theme persistence
	GetTheme() string
	
	// GetAvailableThemes returns list of available themes
	// Called by: Theme selection UI, Settings
	GetAvailableThemes() []string
	
	// Color and Style
	// GetPlayerColor returns color for specific player
	// Called by: Unit rendering, UI elements
	GetPlayerColor(playerID int) (r, g, b, a uint8)
	
	// GetTerrainColor returns color for terrain type
	// Called by: Tile rendering, Map display
	GetTerrainColor(terrainType int) (r, g, b, a uint8)
	
	// GetUIColor returns color for UI elements
	// Called by: UI rendering, Button styling
	GetUIColor(elementType string) (r, g, b, a uint8)
}

// =============================================================================
// UI Utility Functions
// =============================================================================

// ClickType constants for click result types
const (
	ClickTile  = "tile"
	ClickUnit  = "unit"
	ClickUI    = "ui"
	ClickEmpty = "empty"
)

// Action constants for suggested actions
const (
	ActionSelect = "select"
	ActionMove   = "move"
	ActionAttack = "attack"
	ActionInfo   = "info"
	ActionCancel = "cancel"
)

// Animation type constants
const (
	AnimationMovement  = "movement"
	AnimationAttack    = "attack"
	AnimationExplosion = "explosion"
	AnimationFade      = "fade"
	AnimationPulse     = "pulse"
)

// Cursor style constants
const (
	CursorDefault = "default"
	CursorPointer = "pointer"
	CursorMove    = "move"
	CursorAttack  = "crosshair"
	CursorInfo    = "help"
)