//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/panyam/turnengine/games/weewar/assets"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// =============================================================================
// Global State (Initialized in main)
// =============================================================================

var globalEditor *weewar.WorldEditor
var globalWorld *weewar.World
var globalAssetProvider weewar.AssetProvider

// =============================================================================
// Response Types
// =============================================================================

// WASMResponse represents a standardized JavaScript-friendly response
type WASMResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// =============================================================================
// Generic Wrapper Infrastructure
// =============================================================================

// WASMFunction represents a function that takes js.Value args and returns (data, error)
type WASMFunction func(args []js.Value) (any, error)

// createWrapper creates a generic wrapper for WASM functions with validation and error handling
func createWrapper(minArgs, maxArgs int, fn WASMFunction) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		// Validate argument count
		if len(args) < minArgs {
			return createErrorResponse(fmt.Sprintf("Expected at least %d arguments, got %d", minArgs, len(args)))
		}
		if maxArgs >= 0 && len(args) > maxArgs {
			return createErrorResponse(fmt.Sprintf("Expected at most %d arguments, got %d", maxArgs, len(args)))
		}

		// Call the function and handle response
		result, err := fn(args)
		if err != nil {
			return createErrorResponse(err.Error())
		}

		return createSuccessResponse(result)
	})
}

// =============================================================================
// Response Helpers
// =============================================================================

func createSuccessResponse(data any) js.Value {
	response := WASMResponse{
		Success: true,
		Message: "Operation completed successfully",
		Data:    data,
	}
	return marshalToJS(response)
}

func createErrorResponse(error string) js.Value {
	response := WASMResponse{
		Success: false,
		Error:   error,
	}
	return marshalToJS(response)
}

func createMessageResponse(message string, data any) js.Value {
	response := WASMResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	return marshalToJS(response)
}

func marshalToJS(obj any) js.Value {
	bytes, _ := json.Marshal(obj)
	return js.Global().Get("JSON").Call("parse", string(bytes))
}

// =============================================================================
// Main Function - Initialize Everything
// =============================================================================

func main() {
	fmt.Println("WeeWar Map Editor WASM initializing...")

	// Initialize World (2 players by default)
	globalWorld, _ = weewar.NewWorld(2, weewar.NewMapRect(5, 5))

	// Initialize WorldEditor with the World
	globalEditor = weewar.NewWorldEditor()
	globalEditor.NewWorld() // This creates a 1x1 world internally
	
	// Set default map size to 5x5
	_, err := setMapSize(5, 5)
	if err != nil {
		fmt.Printf("Warning: Failed to set default map size: %v\n", err)
	}

	// Initialize and preload assets
	globalAssetProvider = assets.NewEmbeddedAssetManager()
	if globalAssetProvider != nil {
		err := globalAssetProvider.PreloadCommonAssets()
		if err != nil {
			fmt.Printf("Warning: Failed to preload assets: %v\n", err)
		} else {
			fmt.Println("Assets preloaded successfully")
		}
		globalEditor.SetAssetProvider(globalAssetProvider)
	}

	// Register all editor functions with clean wrappers
	registerEditorFunctions()

	// Register utility functions
	registerUtilityFunctions()

	fmt.Println("WeeWar Map Editor WASM loaded and ready")

	// Keep the program running
	<-make(chan struct{})
}

// =============================================================================
// Function Registration
// =============================================================================

func registerEditorFunctions() {
	// Map management
	js.Global().Set("editorNewMap", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return newMap(args[0].Int(), args[1].Int())
	}))
	js.Global().Set("editorSetMapSize", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return setMapSize(args[0].Int(), args[1].Int())
	}))

	// Terrain editing
	js.Global().Set("editorPaintTerrain", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return nil, paintTerrain(args[0].Int(), args[1].Int())
	}))
	js.Global().Set("editorRemoveTerrain", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return nil, removeTerrain(args[0].Int(), args[1].Int())
	}))
	js.Global().Set("editorFloodFill", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return nil, floodFill(args[0].Int(), args[1].Int())
	}))

	// Brush settings
	js.Global().Set("editorSetBrushTerrain", createWrapper(1, 1, func(args []js.Value) (any, error) {
		return nil, setBrushTerrain(args[0].Int())
	}))
	js.Global().Set("editorSetBrushSize", createWrapper(1, 1, func(args []js.Value) (any, error) {
		return setBrushSize(args[0].Int())
	}))

	// Visual settings
	js.Global().Set("editorSetShowGrid", createWrapper(1, 1, func(args []js.Value) (any, error) {
		return nil, setShowGrid(args[0].Bool())
	}))
	js.Global().Set("editorSetShowCoordinates", createWrapper(1, 1, func(args []js.Value) (any, error) {
		return nil, setShowCoordinates(args[0].Bool())
	}))

	// Rendering
	js.Global().Set("editorRender", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return nil, renderEditor()
	}))
	js.Global().Set("editorSetCanvas", createWrapper(3, 3, func(args []js.Value) (any, error) {
		return setCanvas(args[0].String(), args[1].Int(), args[2].Int())
	}))
	js.Global().Set("editorSetViewPort", createWrapper(4, 4, func(args []js.Value) (any, error) {
		return nil, setViewPort(args[0].Int(), args[1].Int(), args[2].Int(), args[3].Int())
	}))

	// Information
	js.Global().Set("editorGetMapInfo", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return getMapInfo()
	}))
	js.Global().Set("editorValidateMap", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return validateMap()
	}))
	js.Global().Set("editorGetTerrainTypes", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return getTerrainTypes()
	}))
	js.Global().Set("editorGetTileDimensions", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return getTileDimensions()
	}))
	js.Global().Set("editorGetMapBounds", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return getMapBounds()
	}))
}

func registerUtilityFunctions() {
	// Coordinate conversion
	js.Global().Set("pixelToCoords", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return pixelToCoords(args[0].Float(), args[1].Float())
	}))
	js.Global().Set("calculateCanvasSize", createWrapper(2, 2, func(args []js.Value) (any, error) {
		return calculateCanvasSize(args[0].Int(), args[1].Int())
	}))

	// Asset testing
	js.Global().Set("testAssets", createWrapper(0, 0, func(args []js.Value) (any, error) {
		return testAssets()
	}))
}

// =============================================================================
// Editor Function Implementations (Clean, No Boilerplate)
// =============================================================================

func newMap(rows, cols int) (map[string]any, error) {
	// Calculate optimal canvas size for the new map
	width, height := calculateCanvasSizeInternal()

	// Create new map in the editor
	err := globalEditor.NewWorld() // Creates 1x1, we'll expand it
	if err != nil {
		return nil, err
	}

	// TODO: Expand map to rows x cols using Add/Remove methods
	// For now, just use the 1x1 map

	return map[string]any{
		"width":        cols,
		"height":       rows,
		"canvasWidth":  width,
		"canvasHeight": height,
	}, nil
}

func setMapSize(rows, cols int) (map[string]any, error) {
	return newMap(rows, cols)
}

func paintTerrain(q, r int) error {
	// Create cube coordinate directly from Q, R values
	coord := weewar.CubeCoord{Q: q, R: r}
	return globalEditor.PaintTerrain(coord)
}

func removeTerrain(q, r int) error {
	coord := weewar.CubeCoord{Q: q, R: r}
	return globalEditor.RemoveTerrain(coord)
}

func floodFill(q, r int) error {
	coord := weewar.CubeCoord{Q: q, R: r}
	return globalEditor.FloodFill(coord)
}

func setBrushTerrain(terrainType int) error {
	return globalEditor.SetBrushTerrain(terrainType)
}

func setBrushSize(size int) (map[string]any, error) {
	err := globalEditor.SetBrushSize(size)
	if err != nil {
		return nil, err
	}

	hexCount := 1
	if size > 0 {
		hexCount = 1 + 6*size*(size+1)/2 // Formula for hex area
	}

	return map[string]any{
		"size":     size,
		"hexCount": hexCount,
	}, nil
}

func setShowGrid(showGrid bool) error {
	return globalEditor.SetShowGrid(showGrid)
}

func setShowCoordinates(showCoordinates bool) error {
	return globalEditor.SetShowCoordinates(showCoordinates)
}

func renderEditor() error {
	return globalEditor.RenderFull()
}

func setCanvas(canvasID string, width, height int) (map[string]any, error) {
	// Create canvas drawable for the editor
	canvasDrawable := weewar.NewCanvasBuffer(canvasID, width, height)
	err := globalEditor.SetDrawable(canvasDrawable, width, height)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"canvasID": canvasID,
		"width":    width,
		"height":   height,
	}, nil
}

func setViewPort(x, y, width, height int) error {
	fmt.Printf("WASM setViewPort called with: x=%d, y=%d, width=%d, height=%d\n", x, y, width, height)
	return globalEditor.SetViewPort(x, y, width, height)
}

func getMapInfo() (map[string]any, error) {
	info := globalEditor.GetMapInfo()
	if info == nil {
		return nil, fmt.Errorf("no map loaded")
	}

	return map[string]any{
		"filename":      info.Filename,
		"width":         info.Width,
		"height":        info.Height,
		"totalTiles":    info.TotalTiles,
		"terrainCounts": info.TerrainCounts,
		"modified":      info.Modified,
	}, nil
}

func validateMap() (map[string]any, error) {
	issues := globalEditor.ValidateMap()
	isValid := len(issues) == 0

	return map[string]any{
		"valid":  isValid,
		"issues": issues,
	}, nil
}

func getTerrainTypes() (map[string]any, error) {
	terrainTypes := []map[string]any{
		{"id": 0, "name": "Unknown", "moveCost": 1, "defenseBonus": 0},
		{"id": 1, "name": "Grass", "moveCost": 1, "defenseBonus": 0},
		{"id": 2, "name": "Desert", "moveCost": 1, "defenseBonus": 0},
		{"id": 3, "name": "Water", "moveCost": 2, "defenseBonus": 0},
		{"id": 4, "name": "Mountain", "moveCost": 2, "defenseBonus": 10},
		{"id": 5, "name": "Rock", "moveCost": 3, "defenseBonus": 20},
	}

	return map[string]any{
		"terrainTypes": terrainTypes,
	}, nil
}

func getTileDimensions() (map[string]any, error) {
	return map[string]any{
		"tileWidth":  int(weewar.DefaultTileWidth),
		"tileHeight": int(weewar.DefaultTileHeight),
		"yIncrement": int(weewar.DefaultYIncrement),
	}, nil
}

func getMapBounds() (map[string]any, error) {
	// Get map bounds from the editor
	minX, minY, maxX, maxY, minQ, maxQ, minR, maxR, startingCoord, startingX := globalEditor.GetMapBounds()

	return map[string]any{
		// Tile dimensions
		"tileWidth":  int(weewar.DefaultTileWidth),
		"tileHeight": int(weewar.DefaultTileHeight),
		"yIncrement": int(weewar.DefaultYIncrement),

		// Map bounds in pixels
		"minX": minX,
		"minY": minY,
		"maxX": maxX,
		"maxY": maxY,

		// Map bounds in hex coordinates
		"minQ": minQ,
		"maxQ": maxQ,
		"minR": minR,
		"maxR": maxR,

		// Starting position info
		"startingCoord": map[string]any{
			"q": startingCoord.Q,
			"r": startingCoord.R,
		},
		"startingX": startingX,
	}, nil
}

// =============================================================================
// Utility Function Implementations
// =============================================================================

func pixelToCoords(x, y float64) (map[string]any, error) {
	coord := globalWorld.Map.XYToQR(x, y, weewar.DefaultTileWidth, weewar.DefaultTileHeight, weewar.DefaultYIncrement)

	// Convert cube coordinates to row/col using proper conversion
	row, col := globalWorld.Map.HexToRowCol(coord)

	isWithinBounds := globalWorld.Map.IsWithinBoundsCube(coord)

	return map[string]any{
		"pixelX":       x,
		"pixelY":       y,
		"row":          row,
		"col":          col,
		"cubeQ":        coord.Q,
		"cubeR":        coord.R,
		"withinBounds": isWithinBounds,
	}, nil
}

func calculateCanvasSize(rows, cols int) (map[string]any, error) {
	width, height := calculateCanvasSizeInternal()

	return map[string]any{
		"width":  width,
		"height": height,
		"rows":   rows,
		"cols":   cols,
	}, nil
}

func calculateCanvasSizeInternal() (width, height int) {
	// Get map bounds and add padding for hover effects and potential expansion
	minX, minY, maxX, maxY, _, _, _, _, _, _ := globalEditor.GetMapBounds()

	// Add padding around the map bounds so we can show hexes being hovered
	// and allow for potential map expansion
	padding := 150.0
	width = int(maxX - minX + 2*padding)
	height = int(maxY - minY + 2*padding)

	// Ensure minimum canvas size
	width = weewar.Max(width, 400)
	height = weewar.Max(height, 300)

	return width, height
}

func testAssets() (map[string]any, error) {
	if globalAssetProvider == nil {
		return nil, fmt.Errorf("no asset provider loaded")
	}

	// Test terrain and unit asset availability
	hasTile := globalAssetProvider.HasTileAsset(1)    // Grass
	hasUnit := globalAssetProvider.HasUnitAsset(1, 0) // Basic unit, player 0

	// Test actual loading
	var tileError, unitError string
	_, err := globalAssetProvider.GetTileImage(1)
	if err != nil {
		tileError = err.Error()
	}

	_, err = globalAssetProvider.GetUnitImage(1, 0)
	if err != nil {
		unitError = err.Error()
	}

	return map[string]any{
		"hasTileAsset":  hasTile,
		"hasUnitAsset":  hasUnit,
		"tileLoadError": tileError,
		"unitLoadError": unitError,
	}, nil
}
