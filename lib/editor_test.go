package weewar

import (
	"os"
	"testing"
)

func TestMapEditorCreation(t *testing.T) {
	editor := NewMapEditor()
	
	if editor == nil {
		t.Fatal("NewMapEditor returned nil")
	}
	
	if editor.GetCurrentMap() != nil {
		t.Error("New editor should have no map loaded")
	}
	
	if editor.IsModified() {
		t.Error("New editor should not be modified")
	}
	
	if editor.GetFilename() != "" {
		t.Error("New editor should have empty filename")
	}
}

func TestMapEditorNewMap(t *testing.T) {
	editor := NewMapEditor()
	
	// Test valid map creation
	err := editor.NewMap(5, 8)
	if err != nil {
		t.Fatalf("Failed to create new map: %v", err)
	}
	
	currentMap := editor.GetCurrentMap()
	if currentMap == nil {
		t.Fatal("Map should be loaded after NewMap")
	}
	
	if currentMap.NumRows != 5 || currentMap.NumCols != 8 {
		t.Errorf("Expected map size 5x8, got %dx%d", currentMap.NumRows, currentMap.NumCols)
	}
	
	// Check that map is filled with grass (terrain type 1)
	tileCount := 0
	for _, tile := range currentMap.Tiles {
		if tile != nil {
			tileCount++
			if tile.TileType != 1 {
				t.Errorf("Expected all tiles to be grass (type 1), found type %d at (%d, %d)", 
					tile.TileType, tile.Row, tile.Col)
			}
		}
	}
	
	expectedTiles := 5 * 8
	if tileCount != expectedTiles {
		t.Errorf("Expected %d tiles, found %d", expectedTiles, tileCount)
	}
	
	// Test invalid map sizes
	err = editor.NewMap(0, 10)
	if err == nil {
		t.Error("Should fail with invalid rows")
	}
	
	err = editor.NewMap(101, 10)
	if err == nil {
		t.Error("Should fail with too many rows")
	}
	
	err = editor.NewMap(10, 0)
	if err == nil {
		t.Error("Should fail with invalid cols")
	}
	
	err = editor.NewMap(10, 101)
	if err == nil {
		t.Error("Should fail with too many cols")
	}
}

func TestMapEditorBrushSettings(t *testing.T) {
	editor := NewMapEditor()
	
	// Test brush terrain setting
	err := editor.SetBrushTerrain(3) // Water
	if err != nil {
		t.Errorf("Failed to set brush terrain: %v", err)
	}
	
	// Test invalid terrain type
	err = editor.SetBrushTerrain(-1)
	if err == nil {
		t.Error("Should fail with invalid terrain type")
	}
	
	err = editor.SetBrushTerrain(999)
	if err == nil {
		t.Error("Should fail with invalid terrain type")
	}
	
	// Test brush size setting
	err = editor.SetBrushSize(2)
	if err != nil {
		t.Errorf("Failed to set brush size: %v", err)
	}
	
	// Test invalid brush size
	err = editor.SetBrushSize(-1)
	if err == nil {
		t.Error("Should fail with negative brush size")
	}
	
	err = editor.SetBrushSize(10)
	if err == nil {
		t.Error("Should fail with too large brush size")
	}
}

func TestMapEditorPaintTerrain(t *testing.T) {
	editor := NewMapEditor()
	
	// Test painting without map
	err := editor.PaintTerrain(0, 0)
	if err == nil {
		t.Error("Should fail to paint without map loaded")
	}
	
	// Create a map
	err = editor.NewMap(5, 5)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Set brush to water
	err = editor.SetBrushTerrain(3)
	if err != nil {
		t.Fatalf("Failed to set brush terrain: %v", err)
	}
	
	// Paint a tile
	err = editor.PaintTerrain(2, 2)
	if err != nil {
		t.Errorf("Failed to paint terrain: %v", err)
	}
	
	// Check that tile was painted
	tile := editor.GetCurrentMap().TileAt(2, 2)
	if tile == nil {
		t.Fatal("Tile should exist after painting")
	}
	
	if tile.TileType != 3 {
		t.Errorf("Expected water (type 3), got type %d", tile.TileType)
	}
	
	if !editor.IsModified() {
		t.Error("Map should be marked as modified after painting")
	}
}

func TestMapEditorBrushSize(t *testing.T) {
	editor := NewMapEditor()
	err := editor.NewMap(10, 10)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Set brush size to 1 (7 hex area)
	err = editor.SetBrushSize(1)
	if err != nil {
		t.Fatalf("Failed to set brush size: %v", err)
	}
	
	// Set brush to water
	err = editor.SetBrushTerrain(3)
	if err != nil {
		t.Fatalf("Failed to set brush terrain: %v", err)
	}
	
	// Paint center of map
	err = editor.PaintTerrain(5, 5)
	if err != nil {
		t.Errorf("Failed to paint with brush: %v", err)
	}
	
	// Count water tiles (should be 7: center + 6 neighbors)
	waterCount := 0
	for _, tile := range editor.GetCurrentMap().Tiles {
		if tile != nil && tile.TileType == 3 {
			waterCount++
		}
	}
	
	if waterCount != 7 {
		t.Errorf("Expected 7 water tiles with brush size 1, got %d", waterCount)
	}
}

func TestMapEditorFloodFill(t *testing.T) {
	editor := NewMapEditor()
	err := editor.NewMap(5, 5)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Create a small area of water tiles manually
	editor.SetBrushTerrain(3) // Water
	editor.SetBrushSize(0)    // Single tile
	editor.PaintTerrain(1, 1)
	editor.PaintTerrain(1, 2)
	editor.PaintTerrain(2, 1)
	
	// Set brush to mountain
	editor.SetBrushTerrain(4) // Mountain
	
	// Flood fill starting from grass area
	err = editor.FloodFill(0, 0)
	if err != nil {
		t.Errorf("Failed to flood fill: %v", err)
	}
	
	// Count terrain types
	grassCount := 0
	waterCount := 0
	mountainCount := 0
	
	for _, tile := range editor.GetCurrentMap().Tiles {
		if tile != nil {
			switch tile.TileType {
			case 1:
				grassCount++
			case 3:
				waterCount++
			case 4:
				mountainCount++
			}
		}
	}
	
	// Should have 3 water tiles unchanged, and rest converted to mountain
	if waterCount != 3 {
		t.Errorf("Expected 3 water tiles, got %d", waterCount)
	}
	
	if grassCount != 0 {
		t.Errorf("Expected 0 grass tiles after flood fill, got %d", grassCount)
	}
	
	expectedMountain := 25 - 3 // Total tiles minus water tiles
	if mountainCount != expectedMountain {
		t.Errorf("Expected %d mountain tiles, got %d", expectedMountain, mountainCount)
	}
}

func TestMapEditorUndoRedo(t *testing.T) {
	editor := NewMapEditor()
	err := editor.NewMap(3, 3)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Initially should not be able to undo/redo
	if editor.CanUndo() {
		t.Error("Should not be able to undo initially")
	}
	
	if editor.CanRedo() {
		t.Error("Should not be able to redo initially")
	}
	
	// Make a change
	editor.SetBrushTerrain(3) // Water
	err = editor.PaintTerrain(1, 1)
	if err != nil {
		t.Errorf("Failed to paint: %v", err)
	}
	
	// Now should be able to undo
	if !editor.CanUndo() {
		t.Error("Should be able to undo after making change")
	}
	
	// Check that tile is water
	tile := editor.GetCurrentMap().TileAt(1, 1)
	if tile == nil || tile.TileType != 3 {
		t.Error("Tile should be water after painting")
	}
	
	// Undo the change
	err = editor.Undo()
	if err != nil {
		t.Errorf("Failed to undo: %v", err)
	}
	
	// Check that tile is back to grass
	tile = editor.GetCurrentMap().TileAt(1, 1)
	if tile == nil || tile.TileType != 1 {
		t.Error("Tile should be grass after undo")
	}
	
	// Now should be able to redo
	if !editor.CanRedo() {
		t.Error("Should be able to redo after undo")
	}
	
	// Redo the change
	err = editor.Redo()
	if err != nil {
		t.Errorf("Failed to redo: %v", err)
	}
	
	// Check that tile is water again
	tile = editor.GetCurrentMap().TileAt(1, 1)
	if tile == nil {
		t.Error("Tile should exist after redo")
	} else if tile.TileType != 3 {
		t.Errorf("Tile should be water (type 3) after redo, got type %d", tile.TileType)
	}
}

func TestMapEditorRemoveTerrain(t *testing.T) {
	editor := NewMapEditor()
	err := editor.NewMap(3, 3)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Remove a tile
	err = editor.RemoveTerrain(1, 1)
	if err != nil {
		t.Errorf("Failed to remove terrain: %v", err)
	}
	
	// Check that tile is gone
	tile := editor.GetCurrentMap().TileAt(1, 1)
	if tile != nil {
		t.Error("Tile should be removed")
	}
	
	if !editor.IsModified() {
		t.Error("Map should be marked as modified after removing terrain")
	}
}

func TestMapEditorMapInfo(t *testing.T) {
	editor := NewMapEditor()
	
	// Test with no map
	info := editor.GetMapInfo()
	if info != nil {
		t.Error("Should return nil info with no map")
	}
	
	// Create a map
	err := editor.NewMap(4, 5)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	info = editor.GetMapInfo()
	if info == nil {
		t.Fatal("Should return map info with loaded map")
	}
	
	if info.Width != 5 || info.Height != 4 {
		t.Errorf("Expected size 5x4, got %dx%d", info.Width, info.Height)
	}
	
	if info.TotalTiles != 20 {
		t.Errorf("Expected 20 total tiles, got %d", info.TotalTiles)
	}
	
	// Check terrain counts
	grassCount, exists := info.TerrainCounts[1]
	if !exists || grassCount != 20 {
		t.Errorf("Expected 20 grass tiles, got %d", grassCount)
	}
	
	if info.Filename != "" {
		t.Errorf("Expected empty filename, got %s", info.Filename)
	}
}

func TestMapEditorValidation(t *testing.T) {
	editor := NewMapEditor()
	
	// Test with no map
	issues := editor.ValidateMap()
	if len(issues) != 1 || issues[0] != "No map loaded" {
		t.Errorf("Expected 'No map loaded' issue, got %v", issues)
	}
	
	// Create a small map
	err := editor.NewMap(2, 2)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	issues = editor.ValidateMap()
	
	// Should have warning about small size
	foundSizeWarning := false
	for _, issue := range issues {
		if issue == "Map is very small (recommended minimum 3x3)" {
			foundSizeWarning = true
			break
		}
	}
	
	if !foundSizeWarning {
		t.Error("Expected warning about small map size")
	}
	
	// Remove a tile to create a hole
	err = editor.RemoveTerrain(0, 0)
	if err != nil {
		t.Errorf("Failed to remove terrain: %v", err)
	}
	
	issues = editor.ValidateMap()
	
	// Should have warning about holes
	foundHoleWarning := false
	for _, issue := range issues {
		if issue == "Map has holes: 1 tiles missing" {
			foundHoleWarning = true
			break
		}
	}
	
	if !foundHoleWarning {
		t.Error("Expected warning about map holes")
	}
}

func TestMapEditorExportToGame(t *testing.T) {
	editor := NewMapEditor()
	
	// Test with no map
	game, err := editor.ExportToGame(2)
	if err == nil {
		t.Error("Should fail to export with no map")
	}
	
	// Create a map
	err = editor.NewMap(5, 5)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Test invalid player counts
	game, err = editor.ExportToGame(1)
	if err == nil {
		t.Error("Should fail with too few players")
	}
	
	game, err = editor.ExportToGame(7)
	if err == nil {
		t.Error("Should fail with too many players")
	}
	
	// Test valid export
	game, err = editor.ExportToGame(2)
	if err != nil {
		t.Errorf("Failed to export to game: %v", err)
	}
	
	if game == nil {
		t.Fatal("Game should not be nil after successful export")
	}
	
	if game.PlayerCount != 2 {
		t.Errorf("Expected 2 players, got %d", game.PlayerCount)
	}
	
	if game.Map == nil {
		t.Error("Game should have a map")
	}
	
	if game.Map.NumRows != 5 || game.Map.NumCols != 5 {
		t.Errorf("Expected map size 5x5, got %dx%d", game.Map.NumRows, game.Map.NumCols)
	}
}

func TestMapEditorRenderToFile(t *testing.T) {
	editor := NewMapEditor()
	
	// Test with no map
	err := editor.RenderToFile("/tmp/test.png", 400, 300)
	if err == nil {
		t.Error("Should fail to render with no map")
	}
	
	// Create a map
	err = editor.NewMap(4, 4)
	if err != nil {
		t.Fatalf("Failed to create map: %v", err)
	}
	
	// Paint some different terrain types
	editor.SetBrushTerrain(3) // Water
	editor.PaintTerrain(1, 1)
	editor.SetBrushTerrain(4) // Mountain
	editor.PaintTerrain(2, 2)
	
	// Render to file
	renderPath := getTestOutputPath("TestMapEditorRender", "editor_render.png")
	err = editor.RenderToFile(renderPath, 400, 300)
	if err != nil {
		t.Errorf("Failed to render to file: %v", err)
	}
	
	// Check that file was created
	if _, err := os.Stat(renderPath); os.IsNotExist(err) {
		t.Errorf("Render file was not created at %s", renderPath)
	} else {
		t.Logf("Map editor render saved to: %s", renderPath)
	}
	
	// Test without .png extension
	renderPath2 := getTestOutputPath("TestMapEditorRender", "editor_render_no_ext")
	err = editor.RenderToFile(renderPath2, 300, 200)
	if err != nil {
		t.Errorf("Failed to render to file without extension: %v", err)
	}
	
	// Should have added .png extension
	expectedPath := renderPath2 + ".png"
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Render file with .png extension was not created at %s", expectedPath)
	}
	
	// Clean up
	if cleanupTestFiles {
		os.Remove(renderPath)
		os.Remove(expectedPath)
	}
}