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

	// Map dimension functions
	js.Global().Set("editorSetMapSize", js.FuncOf(setMapSize))

	// Debug functions
	js.Global().Set("debugAssetLoading", js.FuncOf(debugAssetLoading))

	// Asset management functions
	js.Global().Set("loadEmbeddedAssets", js.FuncOf(loadEmbeddedAssets))
	js.Global().Set("testEmbeddedAssets", js.FuncOf(testEmbeddedAssets))
	js.Global().Set("loadFetchAssets", js.FuncOf(loadFetchAssets))
	js.Global().Set("testFetchAssets", js.FuncOf(testFetchAssets))

	// Legacy function (for backward compatibility during transition)
	js.Global().Set("editorRenderMap", js.FuncOf(renderMap))

	// Canvas rendering function
	js.Global().Set("editorRenderToCanvas", js.FuncOf(renderEditorToCanvasJS))

	fmt.Println("WeeWar Map Editor WASM loaded")
	<-c
}

// createEditor creates a new map editor instance with canvas binding
func createEditor(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return createEditorResponse(false, "", "Missing canvas ID argument", nil)
	}

	canvasID := args[0].String()

	// Start with default 8x8 map for initial canvas size
	defaultRows, defaultCols := 8, 8
	width, height := calculateCanvasSize(defaultRows, defaultCols)

	globalEditor = weewar.NewMapEditor()

	// Bind the editor to the canvas with auto-calculated size
	err := globalEditor.SetCanvas(canvasID, width, height)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to bind canvas: %v", err), nil)
	}

	// Create the initial default map
	err = globalEditor.NewMap(defaultRows, defaultCols)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to create initial map: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Map editor created and bound to canvas '%s' (%dx%d)", canvasID, width, height), "", map[string]any{
		"version":      "1.0.0",
		"ready":        true,
		"canvasID":     canvasID,
		"canvasWidth":  width,
		"canvasHeight": height,
		"mapWidth":     defaultCols,
		"mapHeight":    defaultRows,
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

	// Automatically render to the bound canvas
	err = renderEditorToCanvas()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to render updated terrain: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Terrain painted at (%d, %d)", row, col), "", nil)
}

// setMapSize sets the map size and immediately renders to the bound canvas
// calculateCanvasSize calculates the optimal canvas size for a given map size
// Returns width and height in pixels for 1:1 tile ratio
func calculateCanvasSize(rows, cols int) (width, height int) {
	const tileSize = 40     // pixels per hex tile
	const hexSpacing = 0.75 // vertical spacing factor for hex grid

	// Calculate canvas dimensions based on hex grid layout
	width = cols*tileSize + tileSize                                    // extra space for hex offset
	height = int(float64(rows)*float64(tileSize)*hexSpacing) + tileSize // extra space for hex height

	// Ensure minimum size for usability
	if width < 200 {
		width = 200
	}
	if height < 200 {
		height = 200
	}

	return width, height
}

func setMapSize(this js.Value, args []js.Value) any {
	if globalEditor == nil {
		return createEditorResponse(false, "", "Editor not initialized", nil)
	}

	if len(args) < 2 {
		return createEditorResponse(false, "", "Missing rows/cols arguments", nil)
	}

	rows := args[0].Int()
	cols := args[1].Int()

	// Calculate optimal canvas size for the map dimensions
	canvasWidth, canvasHeight := calculateCanvasSize(rows, cols)

	// Resize the canvas to match the map size
	err := globalEditor.SetCanvasSize(canvasWidth, canvasHeight)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to resize canvas: %v", err), nil)
	}

	// Create a new map with the new dimensions
	err = globalEditor.NewMap(rows, cols)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to set map size: %v", err), nil)
	}

	// Automatically render to the bound canvas
	err = renderEditorToCanvas()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to render map: %v", err), nil)
	}

	return createEditorResponse(true, fmt.Sprintf("Map size set to %dx%d", rows, cols), "", map[string]any{
		"mapWidth":     cols,
		"mapHeight":    rows,
		"canvasWidth":  canvasWidth,
		"canvasHeight": canvasHeight,
	})
}

// renderEditorToCanvasJS is the JavaScript wrapper for renderEditorToCanvas
func renderEditorToCanvasJS(this js.Value, args []js.Value) any {
	err := renderEditorToCanvas()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to render: %v", err), nil)
	}
	return createEditorResponse(true, "Map rendered successfully", "", nil)
}

// renderEditorToCanvas renders the current editor state to the bound canvas
func renderEditorToCanvas() error {
	if globalEditor == nil {
		return fmt.Errorf("editor not initialized")
	}

	// Use the layered renderer for fast prototyping rendering
	layeredRenderer := globalEditor.GetLayeredRenderer()
	if layeredRenderer != nil {
		fmt.Println("Using layered renderer for rendering")

		// Force render to update all layers with error handling
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("PANIC in layered renderer: %v\n", r)
			}
		}()

		// Ensure units are marked as dirty before render (batching system issue)
		layeredRenderer.MarkAllUnitsDirty()
		layeredRenderer.ForceRender()

		// Blit each layer to canvas sequentially with alpha blending
		canvasID := "map-canvas"

		// Blit all layers: terrain, units, and UI
		terrainBuffer := layeredRenderer.GetTerrainBuffer()
		if terrainBuffer != nil {
			fmt.Println("Blitting terrain buffer to canvas")
			err := blitBufferToCanvas(terrainBuffer, canvasID)
			if err != nil {
				return fmt.Errorf("failed to blit terrain to canvas: %v", err)
			}
		}

		unitBuffer := layeredRenderer.GetUnitBuffer()
		if unitBuffer != nil {
			fmt.Println("Blitting unit buffer to canvas")
			err := blitBufferToCanvas(unitBuffer, canvasID)
			if err != nil {
				return fmt.Errorf("failed to blit units to canvas: %v", err)
			}
		}

		uiBuffer := layeredRenderer.GetUIBuffer()
		if uiBuffer != nil {
			fmt.Println("Blitting UI buffer to canvas")
			err := blitBufferToCanvas(uiBuffer, canvasID)
			if err != nil {
				return fmt.Errorf("failed to blit UI to canvas: %v", err)
			}
		}

		return nil
	}

	// Fallback to old rendering if layered renderer not available
	fmt.Println("Layered renderer not available, using fallback")
	return renderEditorToCanvasFallback()
}

// renderEditorToCanvasFallback provides old rendering method as fallback
func renderEditorToCanvasFallback() error {
	// Export the current editor state to a game for rendering
	game, err := globalEditor.ExportToGame(2)
	if err != nil {
		return fmt.Errorf("failed to export editor state: %v", err)
	}

	// Create/update the world with the new map
	globalWorld = weewar.NewWorld(2, game.Map, int(game.Seed))

	// Ensure we have a view state
	if globalViewState == nil {
		globalViewState = weewar.NewViewState()
	}

	// Set up the game with embedded assets
	if globalEmbeddedAssetManager != nil {
		game.SetAssetProvider(globalEmbeddedAssetManager)
	}

	// Get canvas dimensions from the editor
	canvasWidth, canvasHeight := globalEditor.GetCanvasSize()

	// Render using BufferRenderer
	bufferRenderer := weewar.NewBufferRenderer()
	buffer := weewar.NewBuffer(canvasWidth, canvasHeight)

	baseRenderer := &weewar.BaseRenderer{}
	options := baseRenderer.CalculateRenderOptions(canvasWidth, canvasHeight, globalWorld)

	bufferRenderer.RenderWorldWithAssets(globalWorld, globalViewState, buffer, options, game)

	// Push the buffer directly to the bound canvas
	err = blitBufferToCanvas(buffer, "map-canvas")
	if err != nil {
		return fmt.Errorf("failed to push buffer to canvas: %v", err)
	}

	return nil
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
var globalEmbeddedAssetManager *assets.EmbeddedAssetManager

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
	seed := args[2].Int()

	// Use the globalEditor's current map if available, otherwise create a default map
	var testMap *weewar.Map
	if globalEditor != nil && globalEditor.GetMapInfo() != nil {
		// Export the current editor map to use for rendering
		mapInfo := globalEditor.GetMapInfo()
		testMap = weewar.NewMap(mapInfo.Height, mapInfo.Width, false)

		// Fill with current terrain data from the editor
		for row := 0; row < mapInfo.Height; row++ {
			for col := 0; col < mapInfo.Width; col++ {
				terrainType := 1 // Default grass
				// In a full implementation, we'd get the actual terrain from the editor
				// For now, just use grass with some variety
				if row == 1 && (col == 1 || col == 2) {
					terrainType = 3 // Water
				} else if row == 2 && col == 3 {
					terrainType = 4 // Mountain
				}
				tile := weewar.NewTile(row, col, terrainType)
				testMap.AddTile(tile)
			}
		}
	} else {
		// Create a default 5x8 map if no editor map is available
		testMap = weewar.NewMap(5, 8, false)

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
	}

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

// worldRendererRender renders a World using BufferRenderer and blits to canvas
func worldRendererRender(this js.Value, args []js.Value) any {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in worldRendererRender: %v\n", r)
		}
	}()

	if globalWorld == nil {
		return createEditorResponse(false, "", "globalWorld is nil - run worldCreate first", nil)
	}
	if globalViewState == nil {
		return createEditorResponse(false, "", "globalViewState is nil - run viewStateCreate first", nil)
	}
	fmt.Printf("DEBUG: Global objects are non-nil\n")

	if len(args) < 3 {
		return createEditorResponse(false, "", "Missing canvasID/width/height arguments", nil)
	}

	canvasID := args[0].String()
	width := args[1].Int()
	height := args[2].Int()

	// Create a Game instance with embedded assets (same as CLI)
	testGame, err := weewar.NewGame(2, globalWorld.Map, int64(globalWorld.Seed))
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to create game for rendering: %v", err), nil)
	}
	fmt.Printf("DEBUG: testGame created successfully\n")

	// Use embedded assets if available
	if globalEmbeddedAssetManager != nil {
		testGame.SetAssetProvider(globalEmbeddedAssetManager)
		fmt.Printf("DEBUG: embedded asset manager set on testGame\n")
	} else {
		fmt.Printf("DEBUG: globalEmbeddedAssetManager is nil\n")
	}

	// Use BufferRenderer (same as CLI) to avoid canvas deadlock issues
	bufferRenderer := weewar.NewBufferRenderer()
	buffer := weewar.NewBuffer(width, height)

	// Calculate render options based on world and canvas size
	baseRenderer := &weewar.BaseRenderer{}
	options := baseRenderer.CalculateRenderOptions(width, height, globalWorld)
	fmt.Printf("DEBUG: render options calculated - tileWidth=%d, tileHeight=%d\n", options.TileWidth, options.TileHeight)

	// Render using BufferRenderer with embedded assets (same as CLI)
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic in BufferRenderer: %v\n", r)
			}
		}()
		bufferRenderer.RenderWorldWithAssets(globalWorld, globalViewState, buffer, options, testGame)
	}()

	// Convert buffer to canvas ImageData and blit to HTML canvas
	err = blitBufferToCanvas(buffer, canvasID)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to blit buffer to canvas: %v", err), nil)
	}

	return createEditorResponse(true, "World rendered with BufferRenderer", "", map[string]any{
		"canvasID":   canvasID,
		"width":      width,
		"height":     height,
		"tileWidth":  options.TileWidth,
		"tileHeight": options.TileHeight,
		"yIncrement": options.YIncrement,
	})
}

// loadEmbeddedAssets switches global game instance to use embedded assets
func loadEmbeddedAssets(this js.Value, args []js.Value) any {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in loadEmbeddedAssets: %v\n", r)
		}
	}()

	// Create embedded asset manager
	embeddedAssets := assets.NewEmbeddedAssetManager()
	if embeddedAssets == nil {
		return createEditorResponse(false, "", "Failed to create embedded asset manager", nil)
	}

	// Preload common assets
	err := embeddedAssets.PreloadCommonAssets()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to preload embedded assets: %v", err), nil)
	}

	// Store globally for use in rendering
	globalEmbeddedAssetManager = embeddedAssets

	// Update editor to use the loaded assets for terrain sprites
	if globalEditor != nil {
		globalEditor.SetAssetProvider(embeddedAssets)
	}

	tileCount, unitCount := embeddedAssets.GetCacheStats()

	return createEditorResponse(true, "Embedded assets loaded successfully", "", map[string]any{
		"tilesLoaded": tileCount,
		"unitsLoaded": unitCount,
		"ready":       true,
	})
}

// testEmbeddedAssets tests embedded asset loading
func testEmbeddedAssets(this js.Value, args []js.Value) any {
	embeddedAssets := assets.NewEmbeddedAssetManager()

	// Test asset existence checks
	hasTile := embeddedAssets.HasTileAsset(1)    // Grass tile
	hasUnit := embeddedAssets.HasUnitAsset(1, 0) // Basic unit, player 0

	// Test actual asset loading
	var tileError, unitError string
	_, err := embeddedAssets.GetTileImage(1)
	if err != nil {
		tileError = err.Error()
	}

	_, err = embeddedAssets.GetUnitImage(1, 0)
	if err != nil {
		unitError = err.Error()
	}

	tileCount, unitCount := embeddedAssets.GetCacheStats()

	return createEditorResponse(true, "Embedded asset test complete", "", map[string]any{
		"hasTileAsset":  hasTile,
		"hasUnitAsset":  hasUnit,
		"tileLoadError": tileError,
		"unitLoadError": unitError,
		"tilesInCache":  tileCount,
		"unitsInCache":  unitCount,
		"note":          "Using embedded assets instead of os.Open()",
	})
}

// debugAssetLoading tests asset loading to identify the root cause
func debugAssetLoading(this js.Value, args []js.Value) any {
	// Test 1: Create a Game with AssetManager
	testGame, err := weewar.NewGame(2, weewar.NewMap(3, 3, false), 12345)
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to create test game: %v", err), nil)
	}

	assetManager := testGame.GetAssetManager()
	if assetManager == nil {
		return createEditorResponse(false, "", "AssetManager is nil", nil)
	}

	// Test 2: Try to check if an asset exists (this uses os.Stat internally)
	hasTile := assetManager.HasTileAsset(1)    // Grass tile
	hasUnit := assetManager.HasUnitAsset(1, 0) // Basic unit, player 0

	// Test 3: Try to actually load an asset (this uses os.Open internally)
	var tileError, unitError string
	_, err = assetManager.GetTileImage(1)
	if err != nil {
		tileError = err.Error()
	}

	_, err = assetManager.GetUnitImage(1, 0)
	if err != nil {
		unitError = err.Error()
	}

	return createEditorResponse(true, "Asset loading debug complete", "", map[string]any{
		"hasTileAsset":  hasTile,
		"hasUnitAsset":  hasUnit,
		"tileLoadError": tileError,
		"unitLoadError": unitError,
		"assetPath":     "data", // The hardcoded path used
		"note":          "In WASM/browser, os.Open() and os.Stat() fail on local files",
	})
}

// loadFetchAssets switches global game instance to use fetch-based assets
func loadFetchAssets(this js.Value, args []js.Value) any {
	// Base URL for assets (current directory by default)
	baseURL := "."
	if len(args) >= 1 {
		baseURL = args[0].String()
	}

	// Create fetch asset manager
	fetchAssets := weewar.NewFetchAssetManager(baseURL)

	// Preload common assets
	err := fetchAssets.PreloadCommonAssets()
	if err != nil {
		return createEditorResponse(false, "", fmt.Sprintf("Failed to preload fetch assets: %v", err), nil)
	}

	// Update global world's game to use fetch assets if it exists
	if globalWorld != nil {
		// Create a new game with fetch assets
		testGame, err := weewar.NewGame(globalWorld.PlayerCount, globalWorld.Map, int64(globalWorld.Seed))
		if err != nil {
			return createEditorResponse(false, "", fmt.Sprintf("Failed to create game with fetch assets: %v", err), nil)
		}

		// Switch to fetch asset provider
		testGame.SetAssetProvider(fetchAssets)
	}

	tileCount, unitCount := fetchAssets.GetCacheStats()

	return createEditorResponse(true, "Fetch assets loaded successfully", "", map[string]any{
		"tilesLoaded": tileCount,
		"unitsLoaded": unitCount,
		"baseURL":     baseURL,
		"ready":       true,
	})
}

// testFetchAssets tests fetch-based asset loading
func testFetchAssets(this js.Value, args []js.Value) any {
	baseURL := "."
	if len(args) >= 1 {
		baseURL = args[0].String()
	}

	fetchAssets := weewar.NewFetchAssetManager(baseURL)

	// Test asset existence checks (always returns true for fetch-based)
	hasTile := fetchAssets.HasTileAsset(1)    // Grass tile
	hasUnit := fetchAssets.HasUnitAsset(1, 0) // Basic unit, player 0

	// Test actual asset loading
	var tileError, unitError string
	_, err := fetchAssets.GetTileImage(1)
	if err != nil {
		tileError = err.Error()
	}

	_, err = fetchAssets.GetUnitImage(1, 0)
	if err != nil {
		unitError = err.Error()
	}

	tileCount, unitCount := fetchAssets.GetCacheStats()

	return createEditorResponse(true, "Fetch asset test complete", "", map[string]any{
		"hasTileAsset":  hasTile,
		"hasUnitAsset":  hasUnit,
		"tileLoadError": tileError,
		"unitLoadError": unitError,
		"tilesInCache":  tileCount,
		"unitsInCache":  unitCount,
		"baseURL":       baseURL,
		"note":          "Using HTTP fetch instead of os.Open() - check console for fetch URLs",
	})
}

// blitBufferToCanvas converts a BufferRenderer output to HTML canvas ImageData and blits it
func blitBufferToCanvas(buffer *weewar.Buffer, canvasID string) error {
	// Get buffer image data
	imageData := buffer.GetImageData()
	width := buffer.GetWidth()
	height := buffer.GetHeight()

	// Get canvas element from DOM
	canvas := js.Global().Get("document").Call("getElementById", canvasID)
	if canvas.IsUndefined() {
		return fmt.Errorf("canvas element '%s' not found", canvasID)
	}

	// Get 2D context
	ctx := canvas.Call("getContext", "2d")
	if ctx.IsUndefined() {
		return fmt.Errorf("failed to get 2D context for canvas '%s'", canvasID)
	}

	// Create ImageData object
	imageDataJS := ctx.Call("createImageData", width, height)
	if imageDataJS.IsUndefined() {
		return fmt.Errorf("failed to create ImageData object")
	}

	// Get the data array from ImageData
	dataArray := imageDataJS.Get("data")
	if dataArray.IsUndefined() {
		return fmt.Errorf("failed to get data array from ImageData")
	}

	// Convert Go image.RGBA to JavaScript Uint8ClampedArray efficiently
	// The buffer is in RGBA format, and ImageData expects RGBA format
	pixels := imageData.Pix

	// Debug: Check if buffer contains any non-transparent pixels
	nonTransparentPixels := 0
	for i := 3; i < len(pixels); i += 4 { // Check alpha channel (every 4th byte)
		if pixels[i] > 0 {
			nonTransparentPixels++
		}
	}
	fmt.Printf("BlitBufferToCanvas: canvas='%s', size=%dx%d, pixels=%d, non-transparent=%d\n",
		canvasID, width, height, len(pixels)/4, nonTransparentPixels)

	// Use js.CopyBytesToJS for efficient bulk transfer instead of pixel-by-pixel
	js.CopyBytesToJS(dataArray, pixels)

	// Use drawImage with globalCompositeOperation for proper alpha blending
	// First create a temporary canvas with the ImageData
	tempCanvas := js.Global().Get("document").Call("createElement", "canvas")
	tempCanvas.Set("width", width)
	tempCanvas.Set("height", height)
	tempCtx := tempCanvas.Call("getContext", "2d")
	tempCtx.Call("putImageData", imageDataJS, 0, 0)

	// Then draw the temporary canvas onto the main canvas with alpha blending
	ctx.Set("globalCompositeOperation", "source-over") // This respects alpha
	ctx.Call("drawImage", tempCanvas, 0, 0)

	// Debug: Check canvas size
	canvasWidth := canvas.Get("width").Int()
	canvasHeight := canvas.Get("height").Int()
	fmt.Printf("Canvas '%s' size: %dx%d, ImageData size: %dx%d\n",
		canvasID, canvasWidth, canvasHeight, width, height)

	return nil
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
