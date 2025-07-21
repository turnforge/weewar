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
	m := NewMapRect(10, 15)
	if m.Tiles == nil {
		t.Errorf("Expected Tiles map to be initialized")
	}
}

func TestMapTileOperations(t *testing.T) {
	m := NewMapRect(5, 5)

	// Test TileAt with no tiles
	coord := AxialCoord{Q: 2, R: 3}
	tile := m.TileAt(coord)
	if tile != nil {
		t.Errorf("Expected nil tile, got %v", tile)
	}

	// Test AddTile and TileAt
	newTile := NewTile(coord, 1)
	m.AddTile(newTile)

	retrievedTile := m.TileAt(coord)
	if retrievedTile == nil {
		t.Errorf("Expected tile at %v, got nil", coord)
	}
	if retrievedTile.Coord != coord {
		t.Errorf("Expected tile at %v, got %v", coord, retrievedTile.Coord)
	}
	if retrievedTile.TileType != 1 {
		t.Errorf("Expected tile type 1, got %d", retrievedTile.TileType)
	}

	// Test tile replacement
	replacementTile := NewTile(coord, 2)
	m.AddTile(replacementTile)

	retrievedTile = m.TileAt(coord)
	if retrievedTile.TileType != 2 {
		t.Errorf("Expected tile type 2 after replacement, got %d", retrievedTile.TileType)
	}

	// Test DeleteTile
	m.DeleteTile(coord)
	retrievedTile = m.TileAt(coord)
	if retrievedTile != nil {
		t.Errorf("Expected nil after deletion, got %v", retrievedTile)
	}
}

func TestGameCreationAndBasicOperations(t *testing.T) {
	// Create a test world with map
	gameMap := NewMapRect(3, 3)
	
	// Add some tiles with different types
	coords := []AxialCoord{
		{Q: 0, R: 0}, {Q: 0, R: 1}, {Q: 1, R: 0},
		{Q: 1, R: 1}, {Q: 2, R: 0}, {Q: 2, R: 1},
	}
	
	tileTypes := []int{1, 2, 3, 1, 4, 5} // Grass, Desert, Water, Grass, Mountain, Rock

	for i, coord := range coords {
		tile := NewTile(coord, tileTypes[i])
		gameMap.AddTile(tile)
	}

	world, err := NewWorld(2, gameMap)
	if err != nil {
		t.Fatalf("Failed to create world: %v", err)
	}

	// Load rules engine first
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	// Create game with rules engine
	game, err := NewGame(world, rulesEngine, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Add some units manually
	unit1 := NewUnit(1, 0)
	unit1.Coord = AxialCoord{Q: 0, R: 0}
	// Initialize unit stats from rules engine
	unitData1, err := rulesEngine.GetUnitData(unit1.UnitType)
	if err != nil {
		t.Fatalf("Failed to get unit data: %v", err)
	}
	unit1.AvailableHealth = unitData1.Health
	unit1.DistanceLeft = unitData1.MovementPoints
	err = game.AddUnit(unit1, 0)
	if err != nil {
		t.Errorf("Failed to add unit1: %v", err)
	}

	unit2 := NewUnit(1, 1)
	unit2.Coord = AxialCoord{Q: 1, R: 1}
	// Initialize unit stats from rules engine
	unitData2, err := rulesEngine.GetUnitData(unit2.UnitType)
	if err != nil {
		t.Fatalf("Failed to get unit data: %v", err)
	}
	unit2.AvailableHealth = unitData2.Health
	unit2.DistanceLeft = unitData2.MovementPoints
	err = game.AddUnit(unit2, 1)
	if err != nil {
		t.Errorf("Failed to add unit2: %v", err)
	}

	// Test basic game operations
	if game.CurrentPlayer != 0 {
		t.Errorf("Expected current player 0, got %d", game.CurrentPlayer)
	}

	if game.TurnCounter != 1 {
		t.Errorf("Expected turn counter 1, got %d", game.TurnCounter)
	}

	// Test unit retrieval
	retrievedUnit := game.GetUnitAt(AxialCoord{Q: 0, R: 0})
	if retrievedUnit == nil {
		t.Error("Failed to retrieve unit at (0,0)")
	} else if retrievedUnit.UnitType != 1 {
		t.Errorf("Expected unit type 1, got %d", retrievedUnit.UnitType)
	}

	t.Logf("Game creation and basic operations test completed successfully")
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

func TestGameMovementAndCombat(t *testing.T) {
	// Create a world with map and units
	gameMap := NewMapRect(3, 3)
	world, err := NewWorld(2, gameMap)
	if err != nil {
		t.Fatalf("Failed to create world: %v", err)
	}

	// Add some tiles in a 3x3 pattern
	coords := []AxialCoord{
		{Q: 0, R: 0}, {Q: 0, R: 1}, {Q: 0, R: 2},
		{Q: 1, R: 0}, {Q: 1, R: 1}, {Q: 1, R: 2},
		{Q: 2, R: 0}, {Q: 2, R: 1}, {Q: 2, R: 2},
	}
	
	tileTypes := []int{1, 2, 3, 2, 3, 4, 3, 4, 5}

	for i, coord := range coords {
		tile := NewTile(coord, tileTypes[i])
		world.Map.AddTile(tile)
	}

	// Load rules engine for movement/combat
	rulesEngine, err := LoadRulesEngineFromFile("../data/rules-data.json")
	if err != nil {
		t.Fatalf("Failed to load rules engine: %v", err)
	}

	game, err := NewGame(world, rulesEngine, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Add units
	unit1 := NewUnit(1, 0) // Soldier
	unit1.Coord = AxialCoord{Q: 0, R: 0}
	// Initialize unit stats from rules engine
	unitData1, err := rulesEngine.GetUnitData(unit1.UnitType)
	if err != nil {
		t.Fatalf("Failed to get unit data: %v", err)
	}
	unit1.AvailableHealth = unitData1.Health
	unit1.DistanceLeft = unitData1.MovementPoints
	game.AddUnit(unit1, 0)

	unit2 := NewUnit(1, 1) // Soldier
	unit2.Coord = AxialCoord{Q: 2, R: 2}
	// Initialize unit stats from rules engine
	unitData2, err := rulesEngine.GetUnitData(unit2.UnitType)
	if err != nil {
		t.Fatalf("Failed to get unit data: %v", err)
	}
	unit2.AvailableHealth = unitData2.Health
	unit2.DistanceLeft = unitData2.MovementPoints
	game.AddUnit(unit2, 1)

	// Test movement validation
	from := AxialCoord{Q: 0, R: 0}
	to := AxialCoord{Q: 0, R: 1}
	canMove := game.IsValidMove(from, to)
	if !canMove {
		t.Error("Expected valid move from (0,0) to (0,1)")
	}

	// Test movement execution
	err = game.MoveUnit(unit1, AxialCoord{Q: 0, R: 1})
	if err != nil {
		t.Errorf("Failed to move unit: %v", err)
	}

	// Verify unit moved
	movedUnit := game.GetUnitAt(AxialCoord{Q: 0, R: 1})
	if movedUnit == nil {
		t.Error("Unit not found at new position")
	}

	// Test attack validation
	canAttack := game.CanAttackUnit(unit1, unit2)
	if canAttack {
		t.Log("Units can attack each other")
	} else {
		t.Log("Units cannot attack each other (likely out of range)")
	}

	t.Logf("Game movement and combat test completed successfully")
}
