//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/panyam/turnengine/games/weewar"
)

// Global editor instance for WASM
var globalEditor *weewar.MapEditor

// EditorResponse represents a JavaScript-friendly response
type EditorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func main() {
	// Keep the program running
	c := make(chan struct{})

	// Register JavaScript functions
	js.Global().Set("editorCreate", js.FuncOf(createEditor))
	js.Global().Set("editorNewMap", js.FuncOf(newMap))
	js.Global().Set("editorLoadMap", js.FuncOf(loadMap))
	js.Global().Set("editorSaveMap", js.FuncOf(saveMap))
	js.Global().Set("editorPaintTerrain", js.FuncOf(paintTerrain))
	js.Global().Set("editorRemoveTerrain", js.FuncOf(removeTerrain))
	js.Global().Set("editorFloodFill", js.FuncOf(floodFill))
	js.Global().Set("editorSetBrushTerrain", js.FuncOf(setBrushTerrain))
	js.Global().Set("editorSetBrushSize", js.FuncOf(setBrushSize))
	js.Global().Set("editorUndo", js.FuncOf(undo))
	js.Global().Set("editorRedo", js.FuncOf(redo))
	js.Global().Set("editorCanUndo", js.FuncOf(canUndo))
	js.Global().Set("editorCanRedo", js.FuncOf(canRedo))
	js.Global().Set("editorGetMapInfo", js.FuncOf(getMapInfo))
	js.Global().Set("editorValidateMap", js.FuncOf(validateMap))
	js.Global().Set("editorExportToGame", js.FuncOf(exportToGame))
	js.Global().Set("editorGetTerrainTypes", js.FuncOf(getTerrainTypes))

	// Canvas initialization and management
	js.Global().Set("editorSetCanvas", js.FuncOf(setCanvas))
	js.Global().Set("editorSetCanvasSize", js.FuncOf(setCanvasSize))

	// New World-Renderer architecture functions
	js.Global().Set("worldCreate", js.FuncOf(worldCreate))
	js.Global().Set("worldCreateTestMap", js.FuncOf(worldCreateTestMap))
	js.Global().Set("viewStateCreate", js.FuncOf(viewStateCreate))
	js.Global().Set("canvasRendererCreate", js.FuncOf(canvasRendererCreate))
	js.Global().Set("worldRendererRender", js.FuncOf(worldRendererRender))

	// Legacy function (for backward compatibility during transition)
	js.Global().Set("editorRenderMap", js.FuncOf(renderMap))

	fmt.Println("WeeWar Map Editor WASM loaded")
	<-c
}

// createEditor creates a new map editor instance
func createEditor(this js.Value, args []js.Value) any {
	globalEditor = weewar.NewMapEditor()

	return createEditorResponse(true, "Map editor created", "", map[string]any{
		"version": "1.0.0",
		"ready":   true,
	})
}

// newMap creates a new map with specified dimensions
func newMap(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing width/height arguments", nil)
	}

	rows := args[0].Int()
	cols := args[1].Int()

	err := globalEditor.NewMap(rows, cols)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to create map: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("New map created (%dx%d)", rows, cols), "", map[string]any{
		"width":  cols,
		"height": rows,
	})
}

// loadMap loads a map from JSON data
func loadMap(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 1 {
		return createEditorResponse(false, "", "Missing map data argument", nil)
	}

	// For now, return not implemented since LoadMap is a placeholder
	return createEditorResponse(false, "", "Map loading not yet implemented", nil)
}

// saveMap saves the current map to JSON
func saveMap(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	filename := "map.json"
	if len(args) >= 1 {
		filename = args[0].String()
	}

	// For now, return not implemented since SaveMap is a placeholder
	err := globalEditor.SaveMap(filename)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to save map: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Map saved as %s", filename), "", nil)
}

// paintTerrain paints terrain at specified coordinates
func paintTerrain(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing row/col arguments", nil)
	}

	row := args[0].Int()
	col := args[1].Int()

	err := globalEditor.PaintTerrain(row, col)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to paint terrain: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Terrain painted at (%d, %d)", row, col), "", nil)
}

// removeTerrain removes terrain at specified coordinates
func removeTerrain(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing row/col arguments", nil)
	}

	row := args[0].Int()
	col := args[1].Int()

	err := globalEditor.RemoveTerrain(row, col)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to remove terrain: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Terrain removed at (%d, %d)", row, col), "", nil)
}

// floodFill performs flood fill at specified coordinates
func floodFill(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing row/col arguments", nil)
	}

	row := args[0].Int()
	col := args[1].Int()

	err := globalEditor.FloodFill(row, col)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to flood fill: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Flood fill applied at (%d, %d)", row, col), "", nil)
}

// setBrushTerrain sets the current brush terrain type
func setBrushTerrain(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 1 {
		return createEditorResponse(false, "", "Missing terrain type argument", nil)
	}

	terrainType := args[0].Int()

	err := globalEditor.SetBrushTerrain(terrainType)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to set brush terrain: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Brush terrain set to type %d", terrainType), "", nil)
}

// setBrushSize sets the current brush size
func setBrushSize(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 1 {
		return createEditorResponse(false, "", "Missing brush size argument", nil)
	}

	size := args[0].Int()

	err := globalEditor.SetBrushSize(size)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to set brush size: %v", err), nil)
	}

	hexCount := 1
	if size > 0 {
		hexCount = 1 + 6*size*(size+1)/2 // Formula for hex area
	}

	return createEditorResponse(true, fmt.Sprintf("Brush size set to %d (affects %d hexes)", size, hexCount), "", map[string]any{
		"size":     size,
		"hexCount": hexCount,
	})
}

// undo undoes the last operation
func undo(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	err := globalEditor.Undo()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Cannot undo: %v", err), nil)
	}

	return createEditorResponse(true, "Undo successful", "", map[string]any{
		"canUndo": globalEditor.CanUndo(),
		"canRedo": globalEditor.CanRedo(),
	})
}

// redo redoes the last undone operation
func redo(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	err := globalEditor.Redo()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Cannot redo: %v", err), nil)
	}

	return createEditorResponse(true, "Redo successful", "", map[string]any{
		"canUndo": globalEditor.CanUndo(),
		"canRedo": globalEditor.CanRedo(),
	})
}

// canUndo checks if undo is available
func canUndo(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	return createEditorResponse(true, "", "", map[string]any{
		"canUndo": globalEditor.CanUndo(),
	})
}

// canRedo checks if redo is available
func canRedo(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	return createEditorResponse(true, "", "", map[string]any{
		"canRedo": globalEditor.CanRedo(),
	})
}

// getMapInfo returns information about the current map
func getMapInfo(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	info := globalEditor.GetMapInfo()
	if info == nil {
		return createEditorResponse(false, "", "No map loaded", nil)
	}

	return createEditorResponse(true, "Map info retrieved", "", map[string]any{
		"filename":      info.Filename,
		"width":         info.Width,
		"height":        info.Height,
		"totalTiles":    info.TotalTiles,
		"terrainCounts": info.TerrainCounts,
		"modified":      info.Modified,
	})
}

// validateMap validates the current map
func validateMap(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	issues := globalEditor.ValidateMap()

	isValid := len(issues) == 0
	message := "Map is valid"
	if !isValid {
		message = fmt.Sprintf("Map has %d issue(s)", len(issues))
	}

	return createEditorResponse(true, message, "", map[string]any{
		"valid":  isValid,
		"issues": issues,
	})
}

// renderMap renders the current map to a data URL
func renderMap(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	// Default dimensions
	width, height := 800, 600

	// Parse optional dimensions
	if len(args) >= 2 {
		width = args[0].Int()
		height = args[1].Int()
	}

	// Create a temporary game for rendering using new World-Renderer architecture
	game, err := globalEditor.ExportToGame(2)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to export for rendering: %v", err), nil)
	}

	// Convert Game to World for new architecture
	world := weewar.NewWorld(2, game.Map, int(game.Seed))
	// Copy units from game to world
	for _, playerUnits := range game.Units {
		for _, unit := range playerUnits {
			if unit != nil {
				world.AddUnit(unit)
			}
		}
	}
	world.CurrentPlayer = game.CurrentPlayer
	// TurnNumber doesn't exist on Game, use a default
	world.TurnNumber = 1

	// Create ViewState for rendering
	viewState := weewar.NewViewState()

	// Create BufferRenderer and render using new architecture
	renderer := weewar.NewBufferRenderer()
	buffer := weewar.NewBuffer(width, height)

	// Calculate proper render options
	options := renderer.CalculateRenderOptions(width, height, world)

	// Render using new World-Renderer architecture with AssetManager support
	renderer.RenderWorldWithAssets(world, viewState, buffer, options, game)

	// Convert buffer to base64 data URL
	dataURL, err := buffer.ToDataURL()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to create data URL: %v", err), nil)
	}

	return createEditorResponse(true, "Map rendered successfully", "", map[string]any{
		"dataURL": dataURL,
		"width":   width,
		"height":  height,
	})
}

// exportToGame exports the current map as a playable game
func exportToGame(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	playerCount := 2
	if len(args) >= 1 {
		playerCount = args[0].Int()
	}

	game, err := globalEditor.ExportToGame(playerCount)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to export to game: %v", err), nil)
	}

	// Save the game data
	saveData, err := game.SaveGame()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to serialize game: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Map exported as %d-player game", playerCount), "", map[string]any{
		"gameData":    string(saveData),
		"playerCount": playerCount,
		"size":        len(saveData),
	})
}

// getTerrainTypes returns available terrain types
func getTerrainTypes(this js.Value, args []js.Value) any {
	// Get terrain data from the weewar package
	terrainTypes := []map[string]any{
		{"id": 0, "name": "Unknown", "moveCost": 1, "defenseBonus": 0},
		{"id": 1, "name": "Grass", "moveCost": 1, "defenseBonus": 0},
		{"id": 2, "name": "Desert", "moveCost": 1, "defenseBonus": 0},
		{"id": 3, "name": "Water", "moveCost": 2, "defenseBonus": 0},
		{"id": 4, "name": "Mountain", "moveCost": 2, "defenseBonus": 10},
		{"id": 5, "name": "Rock", "moveCost": 3, "defenseBonus": 20},
	}

	return createEditorResponse(true, "Terrain types retrieved", "", map[string]any{
		"terrainTypes": terrainTypes,
	})
}

// =============================================================================
// Canvas Management Functions
// =============================================================================

// setCanvas initializes the canvas for rendering
func setCanvas(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 1 {
		return createEditorResponse(false, "", "Missing canvas ID argument", nil)
	}

	canvasID := args[0].String()

	// Default canvas size
	width, height := 800, 600
	if len(args) >= 3 {
		width = args[1].Int()
		height = args[2].Int()
	}

	err := globalEditor.SetCanvas(canvasID, width, height)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to set canvas: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Canvas '%s' initialized (%dx%d)", canvasID, width, height), "", map[string]any{
		"canvasID": canvasID,
		"width":    width,
		"height":   height,
	})
}

// setCanvasSize resizes the canvas
func setCanvasSize(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing width/height arguments", nil)
	}

	width := args[0].Int()
	height := args[1].Int()

	err := globalEditor.SetCanvasSize(width, height)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to resize canvas: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Canvas resized to %dx%d", width, height), "", map[string]any{
		"width":  width,
		"height": height,
	})
}

// =============================================================================
// World-Renderer Architecture Functions
// =============================================================================

// Global instances for testing the new architecture
var globalWorld *weewar.World
var globalViewState *weewar.ViewState
var globalCanvasRenderer *weewar.CanvasRenderer

// worldCreateTestMap creates a test map for World creation
func worldCreateTestMap(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing rows/cols arguments", nil)
	}

	rows := args[0].Int()
	cols := args[1].Int()

	// Create test map
	testMap := weewar.NewMap(rows, cols, false)

	// Fill with default terrain (grass) and add some variety
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			terrainType := 1 // Default to grass

			// Add some variety for testing
			if row == 1 && col == 1 {
				terrainType = 3 // Water
			} else if row == 2 && col == 2 {
				terrainType = 4 // Mountain
			} else if row == 1 && col == 2 {
				terrainType = 3 // Water
			}

			tile := weewar.NewTile(row, col, terrainType)
			testMap.AddTile(tile)
		}
	}

	return createEditorResponse(true, fmt.Sprintf("Test map created (%dx%d)", rows, cols), "", map[string]any{
		"map":  testMap,
		"rows": rows,
		"cols": cols,
	})
}

// worldCreate creates a new World with the given parameters
func worldCreate(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return createEditorResponse(false, "", "Missing playerCount/map/seed arguments", nil)
	}

	playerCount := args[0].Int()
	// For now, create a simple test map - in practice this would use the map from args[1]
	testMap := weewar.NewMap(5, 8, false)

	// Fill with variety of terrains for testing
	for row := 0; row < 5; row++ {
		for col := 0; col < 8; col++ {
			terrainType := 1 // Default grass
			if row == 1 && (col == 1 || col == 2) {
				terrainType = 3 // Water
			} else if row == 2 && col == 3 {
				terrainType = 4 // Mountain
			}
			tile := weewar.NewTile(row, col, terrainType)
			testMap.AddTile(tile)
		}
	}

	seed := args[2].Int()

	// Create the world
	world := weewar.NewWorld(playerCount, testMap, seed)
	globalWorld = world

	return createEditorResponse(true, "World created successfully", "", map[string]any{
		"world":       world,
		"playerCount": playerCount,
		"seed":        seed,
		"mapRows":     testMap.NumRows,
		"mapCols":     testMap.NumCols,
	})
}

// viewStateCreate creates a new ViewState
func viewStateCreate(this js.Value, args []js.Value) any {
	viewState := weewar.NewViewState()
	globalViewState = viewState

	return createEditorResponse(true, "ViewState created successfully", "", map[string]any{
		"viewState":    viewState,
		"showGrid":     viewState.ShowGrid,
		"zoomLevel":    viewState.ZoomLevel,
		"brushTerrain": viewState.BrushTerrain,
		"brushSize":    viewState.BrushSize,
	})
}

// canvasRendererCreate creates a new CanvasRenderer
func canvasRendererCreate(this js.Value, args []js.Value) any {
	renderer := weewar.NewCanvasRenderer()
	globalCanvasRenderer = renderer

	return createEditorResponse(true, "CanvasRenderer created successfully", "", map[string]any{
		"renderer": "CanvasRenderer instance created",
	})
}

// worldRendererRender renders a World using the CanvasRenderer
func worldRendererRender(this js.Value, args []js.Value) any {
	if globalWorld == nil || globalViewState == nil || globalCanvasRenderer == nil {
		return createEditorResponse(false, "", "Missing World, ViewState, or CanvasRenderer - run creation functions first", nil)
	}

	if len(args) < 3 {
		return createEditorResponse(false, "", "Missing canvasID/width/height arguments", nil)
	}

	canvasID := args[0].String()
	width := args[1].Int()
	height := args[2].Int()

	// Create CanvasBuffer for the specified canvas
	canvasBuffer := weewar.NewCanvasBuffer(canvasID, width, height)
	if canvasBuffer == nil {
		return createEditorResponse(false, "", "Failed to create CanvasBuffer - canvas element not found", nil)
	}

	// Create a Game instance with AssetManager (like CLI does) to get asset support
	// For now, create a simple test game - in production this could be from globalEditor.ExportToGame()
	testGame, err := weewar.NewGame(2, globalWorld.Map, int64(globalWorld.Seed))
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to create game for rendering: %v", err), nil)
	}

	// Calculate render options based on world and canvas size
	baseRenderer := &weewar.BaseRenderer{}
	options := baseRenderer.CalculateRenderOptions(width, height, globalWorld)

	// Render using the SAME pattern as CLI: RenderWorldWithAssets with original game for AssetManager!
	globalCanvasRenderer.RenderWorldWithAssets(globalWorld, globalViewState, canvasBuffer, options, testGame)

	return createEditorResponse(true, "World rendered with CanvasRenderer", "", map[string]any{
		"canvasID":   canvasID,
		"width":      width,
		"height":     height,
		"tileWidth":  options.TileWidth,
		"tileHeight": options.TileHeight,
		"yIncrement": options.YIncrement,
	})
}

// createEditorResponse creates a JavaScript-compatible response object
func createEditorResponse(success bool, message, error string, data any) js.Value {
	response := EditorResponse{
		Success: success,
		Message: message,
		Error:   error,
		Data:    data,
	}

	// Convert to JS object
	responseBytes, _ := json.Marshal(response)
	return js.Global().Get("JSON").Call("parse", string(responseBytes))
}
