package weewar

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Global variable to control test file cleanup
var cleanupTestFiles = false

// Helper function to create organized test output directory
func createTestOutputDir(testName string) string {
	timestamp := time.Now().Format("20060102_150405")
	testDir := filepath.Join("/tmp/turnengine", testName, timestamp)
	os.MkdirAll(testDir, 0755)
	return testDir
}

// Helper function to get organized test output path
func getTestOutputPath(testName, filename string) string {
	testDir := createTestOutputDir(testName)
	return filepath.Join(testDir, filename)
}

func TestNewMap(t *testing.T) {
	m := NewMap(10, 15, true)

	if m.NumRows != 10 {
		t.Errorf("Expected NumRows to be 10, got %d", m.NumRows)
	}
	if m.NumCols != 15 {
		t.Errorf("Expected NumCols to be 15, got %d", m.NumCols)
	}
	if !m.EvenRowsOffset {
		t.Errorf("Expected EvenRowsOffset to be true")
	}
	if m.Tiles == nil {
		t.Errorf("Expected Tiles map to be initialized")
	}
}

func TestMapTileOperations(t *testing.T) {
	m := NewMap(5, 5, false)

	// Test TileAt with no tiles
	tile := m.TileAt(2, 3)
	if tile != nil {
		t.Errorf("Expected nil tile, got %v", tile)
	}

	// Test AddTile and TileAt
	newTile := NewTile(2, 3, 1)
	m.AddTile(newTile)

	retrievedTile := m.TileAt(2, 3)
	if retrievedTile == nil {
		t.Errorf("Expected tile at (2,3), got nil")
	}
	if retrievedTile.Row != 2 || retrievedTile.Col != 3 {
		t.Errorf("Expected tile at (2,3), got (%d,%d)", retrievedTile.Row, retrievedTile.Col)
	}
	if retrievedTile.TileType != 1 {
		t.Errorf("Expected tile type 1, got %d", retrievedTile.TileType)
	}

	// Test tile replacement
	replacementTile := NewTile(2, 3, 2)
	m.AddTile(replacementTile)

	retrievedTile = m.TileAt(2, 3)
	if retrievedTile.TileType != 2 {
		t.Errorf("Expected tile type 2 after replacement, got %d", retrievedTile.TileType)
	}

	// Test DeleteTile
	m.DeleteTile(2, 3)
	retrievedTile = m.TileAt(2, 3)
	if retrievedTile != nil {
		t.Errorf("Expected nil after deletion, got %v", retrievedTile)
	}
}

func TestGameRenderToBuffer(t *testing.T) {
	// Create a test map
	m := NewMap(3, 3, false)

	// Add some tiles with different types
	tiles := []*Tile{
		NewTile(0, 0, 1), // Grass
		NewTile(0, 1, 2), // Desert
		NewTile(1, 0, 3), // Water
		NewTile(1, 1, 1), // Grass
		NewTile(2, 0, 4), // Mountain
		NewTile(2, 1, 5), // Rock
	}

	for _, tile := range tiles {
		m.AddTile(tile)
	}

	// Connect neighbors
	m.ConnectHexNeighbors()

	// Create a game
	game, _ := NewGame(2, m, 12345)

	// Add some units
	unit1 := NewUnit(1, 0)
	unit1.Row = 0
	unit1.Col = 0
	game.AddUnit(unit1, 0)

	unit2 := NewUnit(2, 1)
	unit2.Row = 1
	unit2.Col = 1
	game.AddUnit(unit2, 1)

	// Create buffer and render
	buffer := NewBuffer(400, 300)
	game.RenderToBuffer(buffer, 60.0, 50.0, 40.0)

	// Save to PNG
	imagePath := getTestOutputPath("TestGameRenderToBuffer", "game_render.png")
	err := buffer.Save(imagePath)

	if err != nil {
		t.Errorf("Buffer Save failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("PNG file was not created at %s", imagePath)
	} else {
		t.Logf("Game render test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferOperations(t *testing.T) {
	// Test buffer creation
	buffer := NewBuffer(100, 80)
	if buffer == nil {
		t.Fatal("NewBuffer returned nil")
	}

	// Test size
	w, h := buffer.Size()
	if w != 100 || h != 80 {
		t.Errorf("Expected size (100, 80), got (%.0f, %.0f)", w, h)
	}

	// Test copy
	copy := buffer.Copy()
	if copy == nil {
		t.Fatal("Copy returned nil")
	}

	cw, ch := copy.Size()
	if cw != 100 || ch != 80 {
		t.Errorf("Copy size mismatch: expected (100, 80), got (%.0f, %.0f)", cw, ch)
	}

	// Test clear
	buffer.Clear()

	// Test save
	imagePath := getTestOutputPath("TestBufferOperations", "buffer.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Buffer Save failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("PNG file was not created at %s", imagePath)
	} else {
		t.Logf("Buffer operations test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferComposition(t *testing.T) {
	// Create two buffers
	buffer1 := NewBuffer(100, 100)
	buffer2 := NewBuffer(100, 100)

	// Create a simple test image
	testImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			testImg.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}

	// Draw to first buffer
	buffer1.DrawImage(10, 10, 20, 20, testImg)

	// Draw to second buffer
	buffer2.DrawImage(30, 30, 15, 15, testImg)

	// Composite buffers
	finalBuffer := NewBuffer(100, 100)
	finalBuffer.RenderBuffer(buffer1)
	finalBuffer.RenderBuffer(buffer2)

	// Save result
	imagePath := getTestOutputPath("TestBufferComposition", "buffer_composition.png")
	err := finalBuffer.Save(imagePath)
	if err != nil {
		t.Errorf("Buffer composition save failed: %v", err)
	} else {
		t.Logf("Buffer composition test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupTestFiles {
		os.Remove(imagePath)
	}
}

func TestMultiLayerRendering(t *testing.T) {
	// Create a game with map and units
	m := NewMap(3, 3, false)

	// Add some tiles
	tiles := []*Tile{
		NewTile(0, 0, 1), NewTile(0, 1, 2), NewTile(0, 2, 3),
		NewTile(1, 0, 2), NewTile(1, 1, 3), NewTile(1, 2, 4),
		NewTile(2, 0, 3), NewTile(2, 1, 4), NewTile(2, 2, 5),
	}

	for _, tile := range tiles {
		m.AddTile(tile)
	}

	game, _ := NewGame(2, m, 12345)

	// Add units
	unit1 := NewUnit(1, 0)
	unit1.Row = 0
	unit1.Col = 0
	game.AddUnit(unit1, 0)

	unit2 := NewUnit(2, 1)
	unit2.Row = 2
	unit2.Col = 2
	game.AddUnit(unit2, 1)

	// Create separate layer buffers
	terrainBuffer := NewBuffer(300, 250)
	unitBuffer := NewBuffer(300, 250)
	uiBuffer := NewBuffer(300, 250)

	// Render layers separately
	game.RenderTerrain(terrainBuffer, 80.0, 70.0, 50.0)
	game.RenderUnits(unitBuffer, 80.0, 70.0, 50.0)
	game.RenderUI(uiBuffer, 80.0, 70.0, 50.0)

	// Composite layers
	finalBuffer := NewBuffer(300, 250)
	finalBuffer.RenderBuffer(terrainBuffer)
	finalBuffer.RenderBuffer(unitBuffer)
	finalBuffer.RenderBuffer(uiBuffer)

	// Save final result and individual layers
	imagePath := getTestOutputPath("TestMultiLayerRendering", "multi_layer_composite.png")
	terrainPath := getTestOutputPath("TestMultiLayerRendering", "layer_terrain.png")
	unitsPath := getTestOutputPath("TestMultiLayerRendering", "layer_units.png")
	uiPath := getTestOutputPath("TestMultiLayerRendering", "layer_ui.png")

	err := finalBuffer.Save(imagePath)
	if err != nil {
		t.Errorf("Multi-layer rendering save failed: %v", err)
	}

	// Also save individual layers for comparison
	terrainBuffer.Save(terrainPath)
	unitBuffer.Save(unitsPath)
	uiBuffer.Save(uiPath)

	// Verify files exist
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Multi-layer PNG file was not created")
	} else {
		t.Logf("Multi-layer rendering completed successfully")
		t.Logf("Final composite: %s", imagePath)
		t.Logf("Terrain layer: %s", terrainPath)
		t.Logf("Units layer: %s", unitsPath)
		t.Logf("UI layer: %s", uiPath)
	}

	// Clean up (conditional)
	if cleanupTestFiles {
		os.Remove(imagePath)
		os.Remove(terrainPath)
		os.Remove(unitsPath)
		os.Remove(uiPath)
	}
}
