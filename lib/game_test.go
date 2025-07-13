package weewar

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// Game Creation Tests
// =============================================================================

// Helper function to create test map
func createTestMapForTest() *Map {
	gameMap := NewMap(8, 12, false)

	// Add some test tiles
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tileType := 1 // Default to grass
			if (row+col)%4 == 0 {
				tileType = 2 // Some desert
			}
			tile := NewTile(row, col, tileType)
			gameMap.AddTile(tile)
		}
	}

	// Note: Neighbor connections calculated on-demand
	return gameMap
}

func TestNewGame(t *testing.T) {
	// Create a test map first
	testMap := createTestMapForTest()

	// Test successful game creation
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Verify game state
	if game.GetMapName() != "DefaultMap" {
		t.Errorf("Expected map name 'DefaultMap', got '%s'", game.GetMapName())
	}

	if game.PlayerCount != 2 {
		t.Errorf("Expected player count 2, got %d", game.PlayerCount)
	}

	if game.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", game.Seed)
	}

	if game.CurrentPlayer != 0 {
		t.Errorf("Expected current player 0, got %d", game.CurrentPlayer)
	}

	if game.TurnCounter != 1 {
		t.Errorf("Expected turn counter 1, got %d", game.TurnCounter)
	}

	if game.Status != GameStatusPlaying {
		t.Errorf("Expected game status playing, got %s", game.Status)
	}

	// Verify map was created
	if game.Map == nil {
		t.Error("Game map is nil")
	}

	// Verify units were created
	if len(game.Units) != 2 {
		t.Errorf("Expected 2 player unit arrays, got %d", len(game.Units))
	}

	// Verify some units were placed
	totalUnits := 0
	for _, playerUnits := range game.Units {
		totalUnits += len(playerUnits)
	}

	if totalUnits == 0 {
		t.Error("No units were created")
	}

	t.Logf("Game created successfully with %d total units", totalUnits)
}

func TestNewGameValidation(t *testing.T) {
	testMap := createTestMapForTest()

	// Test invalid player count
	_, err := NewGame(1, testMap, 12345)
	if err == nil {
		t.Error("Expected error for invalid player count, got nil")
	}

	_, err = NewGame(7, testMap, 12345)
	if err == nil {
		t.Error("Expected error for invalid player count, got nil")
	}

	// Test nil map
	_, err = NewGame(2, nil, 12345)
	if err == nil {
		t.Error("Expected error for nil map, got nil")
	}
}

// =============================================================================
// Game Interface Tests
// =============================================================================

func TestGameController(t *testing.T) {
	testMap := createTestMapForTest()
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Test GameController methods
	if game.GetCurrentPlayer() != 0 {
		t.Errorf("Expected current player 0, got %d", game.GetCurrentPlayer())
	}

	if game.GetTurnNumber() != 1 {
		t.Errorf("Expected turn number 1, got %d", game.GetTurnNumber())
	}

	if game.GetGameStatus() != GameStatusPlaying {
		t.Errorf("Expected game status playing, got %s", game.GetGameStatus())
	}

	winner, hasWinner := game.GetWinner()
	if hasWinner {
		t.Errorf("Expected no winner, got winner %d", winner)
	}

	// Test turn advancement
	err = game.NextTurn()
	if err != nil {
		t.Errorf("Failed to advance turn: %v", err)
	}

	if game.GetCurrentPlayer() != 1 {
		t.Errorf("Expected current player 1 after turn advance, got %d", game.GetCurrentPlayer())
	}

	// Test CanEndTurn
	if !game.CanEndTurn() {
		t.Error("Expected CanEndTurn to return true")
	}
}

func TestMapInterface(t *testing.T) {
	testMap := createTestMapForTest()
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Test map interface methods
	rows, cols := game.GetMapSize()
	if rows <= 0 || cols <= 0 {
		t.Errorf("Invalid map size: %dx%d", rows, cols)
	}

	if game.GetMapName() != "DefaultMap" {
		t.Errorf("Expected map name 'DefaultMap', got '%s'", game.GetMapName())
	}

	// Test bounds
	minX, minY, maxX, maxY := game.GetMapBounds()
	if maxX <= minX || maxY <= minY {
		t.Errorf("Invalid map bounds: (%f,%f) to (%f,%f)", minX, minY, maxX, maxY)
	}

	// Test tile access
	tile := game.GetTileAt(0, 0)
	if tile == nil {
		t.Error("Expected tile at (0,0), got nil")
	}

	tileType := game.GetTileType(0, 0)
	if tileType < 0 {
		t.Errorf("Invalid tile type: %d", tileType)
	}

	// Test coordinate conversion
	x, y := game.RowColToPixel(0, 0)
	if x < 0 || y < 0 {
		t.Errorf("Invalid pixel coordinates: (%f,%f)", x, y)
	}

	row, col, valid := game.PixelToRowCol(x, y)
	if !valid {
		t.Error("Pixel to row/col conversion failed")
	}

	t.Logf("Map size: %dx%d, bounds: (%f,%f) to (%f,%f)", rows, cols, minX, minY, maxX, maxY)
	t.Logf("Coordinate test: (0,0) -> (%f,%f) -> (%d,%d)", x, y, row, col)
}

func TestUnitInterface(t *testing.T) {
	testMap := createTestMapForTest()
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Test unit queries
	allUnits := game.GetAllUnits()
	if len(allUnits) == 0 {
		t.Error("No units found")
	}

	player0Units := game.GetUnitsForPlayer(0)
	if len(player0Units) == 0 {
		t.Error("No units found for player 0")
	}

	player1Units := game.GetUnitsForPlayer(1)
	if len(player1Units) == 0 {
		t.Error("No units found for player 1")
	}

	// Test unit properties
	unit := player0Units[0]
	unitType := game.GetUnitType(unit)
	if unitType <= 0 {
		t.Errorf("Invalid unit type: %d", unitType)
	}

	health := game.GetUnitHealth(unit)
	if health <= 0 {
		t.Errorf("Invalid unit health: %d", health)
	}

	movement := game.GetUnitMovementLeft(unit)
	if movement < 0 {
		t.Errorf("Invalid unit movement: %d", movement)
	}

	attackRange := game.GetUnitAttackRange(unit)
	if attackRange < 0 {
		t.Errorf("Invalid attack range: %d", attackRange)
	}

	// Test unit at position
	unitAtPos := game.GetUnitAt(unit.Row, unit.Col)
	if unitAtPos != unit {
		t.Error("GetUnitAt returned wrong unit")
	}

	t.Logf("Found %d total units: %d for player 0, %d for player 1",
		len(allUnits), len(player0Units), len(player1Units))
	t.Logf("Unit properties: type=%d, health=%d, movement=%d, range=%d",
		unitType, health, movement, attackRange)
}

// =============================================================================
// Rendering Tests
// =============================================================================

func TestRenderToBuffer(t *testing.T) {
	testMap := createTestMapForTest()
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create a buffer for rendering
	buffer := NewBuffer(800, 600)

	// Test rendering
	err = game.RenderToBuffer(buffer, 64, 64, 51)
	if err != nil {
		t.Errorf("Failed to render to buffer: %v", err)
	}

	// Save rendered image to file for visual verification
	timestamp := time.Now().Format("20060102_150405")
	testDir := filepath.Join("/tmp/turnengine", "TestRenderToBuffer", timestamp)
	os.MkdirAll(testDir, 0755)

	imagePath := filepath.Join(testDir, "game_render.png")
	err = buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save rendered image: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Error("Rendered image file was not created")
	} else {
		t.Logf("Rendered image saved to: %s", imagePath)
	}
}

// TestSparseMapRendering tests rendering with a non-rectangular tile pattern
func TestSparseMapRendering(t *testing.T) {
	// Create a 10x10 map with tiles only in specific pattern
	testMap := NewMap(10, 10, false)
	
	// Add tiles only in a sparse pattern (e.g., checkerboard or every even cell)
	for row := 0; row < 10; row++ {
		for col := 0; col < 10; col++ {
			// Only add tiles where both row and col are even, or in a specific pattern
			if (row%2 == 0 && col%2 == 0) || (row == 5 && col == 5) {
				tile := &Tile{
					Row:      row,
					Col:      col,
					TileType: 1, // Grass
					Unit:     nil,
				}
				testMap.AddTile(tile)
			}
		}
	}
	
	// Note: Neighbor connections calculated on-demand
	
	// Create a game with this sparse map
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	
	// Clear default starting units and place units on our sparse tiles
	game.Units[0] = []*Unit{}
	game.Units[1] = []*Unit{}
	
	// Create units at the sparse tile positions
	unit1, _ := game.CreateUnit(1, 0, 0, 0)  // Player 0 unit at (0,0)
	unit2, _ := game.CreateUnit(1, 0, 0, 2)  // Player 0 unit at (0,2)
	unit3, _ := game.CreateUnit(1, 1, 2, 0)  // Player 1 unit at (2,0)
	unit4, _ := game.CreateUnit(1, 1, 2, 2)  // Player 1 unit at (2,2)
	
	// Place units on tiles
	testMap.TileAt(0, 0).Unit = unit1
	testMap.TileAt(0, 2).Unit = unit2
	testMap.TileAt(2, 0).Unit = unit3
	testMap.TileAt(2, 2).Unit = unit4
	
	// Create a buffer for rendering
	buffer := NewBuffer(800, 600)
	
	// Debug: Print coordinates for some tiles to see the pattern
	t.Logf("Tile coordinate debugging:")
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if testMap.TileAt(row, col) != nil {
				x, y := testMap.XYForTile(row, col, 64, 64, 51)
				t.Logf("Tile (%d,%d) -> pixel (%.1f, %.1f)", row, col, x, y)
			}
		}
	}
	
	// Debug: Print unit positions
	t.Logf("Unit coordinate debugging:")
	for _, unit := range game.GetAllUnits() {
		x, y := testMap.XYForTile(unit.Row, unit.Col, 64, 64, 51)
		t.Logf("Unit at (%d,%d) -> pixel (%.1f, %.1f)", unit.Row, unit.Col, x, y)
	}
	
	// Test rendering
	err = game.RenderToBuffer(buffer, 64, 64, 51)
	if err != nil {
		t.Errorf("Failed to render to buffer: %v", err)
	}
	
	// Save rendered image to file for visual verification
	timestamp := time.Now().Format("20060102_150405")
	testDir := filepath.Join("/tmp/turnengine", "TestSparseMapRendering", timestamp)
	os.MkdirAll(testDir, 0755)
	
	imagePath := filepath.Join(testDir, "sparse_map_render.png")
	err = buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save rendered image: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Error("Rendered image file was not created")
	} else {
		t.Logf("Sparse map rendered image saved to: %s", imagePath)
	}
}

// =============================================================================
// Unit Movement Tests
// =============================================================================

func TestUnitMovement(t *testing.T) {
	testMap := createTestMapForTest()
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Get a unit to move
	units := game.GetUnitsForPlayer(0)
	if len(units) == 0 {
		t.Fatal("No units found for player 0")
	}

	unit := units[0]
	originalRow, originalCol := unit.Row, unit.Col

	// Find a valid move destination
	neighbors := game.GetTileNeighbors(unit.Row, unit.Col)
	var targetRow, targetCol int
	found := false

	for i, neighbor := range neighbors {
		if neighbor != nil && neighbor.Unit == nil {
			coords := game.Map.GetHexNeighborCoords(unit.Row, unit.Col)
			targetRow, targetCol = coords[i][0], coords[i][1]
			found = true
			break
		}
	}

	if !found {
		t.Skip("No valid move destination found")
	}

	// Test movement validation
	if !game.CanMoveUnit(unit, targetRow, targetCol) {
		t.Error("Unit should be able to move to target position")
	}

	// Execute the move
	err = game.MoveUnit(unit, targetRow, targetCol)
	if err != nil {
		t.Errorf("Failed to move unit: %v", err)
	}

	// Verify unit moved
	if unit.Row != targetRow || unit.Col != targetCol {
		t.Errorf("Unit position not updated: expected (%d,%d), got (%d,%d)",
			targetRow, targetCol, unit.Row, unit.Col)
	}

	// Verify unit is at new position on map
	unitAtNewPos := game.GetUnitAt(targetRow, targetCol)
	if unitAtNewPos != unit {
		t.Error("Unit not found at new position")
	}

	// Verify unit is no longer at old position
	unitAtOldPos := game.GetUnitAt(originalRow, originalCol)
	if unitAtOldPos == unit {
		t.Error("Unit still at old position")
	}

	t.Logf("Unit moved from (%d,%d) to (%d,%d)", originalRow, originalCol, targetRow, targetCol)
}

// =============================================================================
// Save/Load Tests
// =============================================================================

func TestSaveLoad(t *testing.T) {
	// Create original game
	testMap := createTestMapForTest()
	originalGame, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Advance a turn to create some state
	err = originalGame.NextTurn()
	if err != nil {
		t.Errorf("Failed to advance turn: %v", err)
	}

	// Save the game
	saveData, err := originalGame.SaveGame()
	if err != nil {
		t.Errorf("Failed to save game: %v", err)
	}

	if len(saveData) == 0 {
		t.Error("Save data is empty")
	}

	// Load the game
	loadedGame, err := LoadGame(saveData)
	if err != nil {
		t.Errorf("Failed to load game: %v", err)
	}

	// Verify loaded game state matches original
	if loadedGame.GetMapName() != originalGame.GetMapName() {
		t.Errorf("Map name mismatch: expected '%s', got '%s'",
			originalGame.GetMapName(), loadedGame.GetMapName())
	}

	if loadedGame.PlayerCount != originalGame.PlayerCount {
		t.Errorf("Player count mismatch: expected %d, got %d",
			originalGame.PlayerCount, loadedGame.PlayerCount)
	}

	if loadedGame.CurrentPlayer != originalGame.CurrentPlayer {
		t.Errorf("Current player mismatch: expected %d, got %d",
			originalGame.CurrentPlayer, loadedGame.CurrentPlayer)
	}

	if loadedGame.TurnCounter != originalGame.TurnCounter {
		t.Errorf("Turn counter mismatch: expected %d, got %d",
			originalGame.TurnCounter, loadedGame.TurnCounter)
	}

	if loadedGame.Seed != originalGame.Seed {
		t.Errorf("Seed mismatch: expected %d, got %d",
			originalGame.Seed, loadedGame.Seed)
	}

	t.Logf("Save/load test successful: %d bytes saved", len(saveData))
}
