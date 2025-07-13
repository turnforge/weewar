package weewar

import (
	"testing"
)

func TestCubeCoordinates(t *testing.T) {
	// Test cube coordinate creation and validation
	coord := NewCubeCoord(1, 2)
	if coord.Q != 1 || coord.R != 2 || coord.S() != -3 {
		t.Errorf("NewCubeCoord(1, 2) = %+v, expected Q=1, R=2, S=-3", coord)
	}
	
	if !coord.IsValid() {
		t.Error("Coordinate should be valid")
	}
	
	// Note: With S() as method, all coordinates are valid by construction
	// This test is no longer needed since invalid coordinates can't be created
}

func TestCubeNeighbors(t *testing.T) {
	center := NewCubeCoord(0, 0)
	
	// Test individual neighbor directions
	right := center.Neighbor(RIGHT)
	expected := CubeCoord{Q: 1, R: 0}
	if right != expected {
		t.Errorf("RIGHT neighbor = %+v, expected %+v", right, expected)
	}
	
	left := center.Neighbor(LEFT)
	expected = CubeCoord{Q: -1, R: 0}
	if left != expected {
		t.Errorf("LEFT neighbor = %+v, expected %+v", left, expected)
	}
	
	// Test all neighbors
	neighbors := center.Neighbors()
	if len(neighbors) != 6 {
		t.Errorf("Expected 6 neighbors, got %d", len(neighbors))
	}
	
	// All neighbors should be valid
	for i, neighbor := range neighbors {
		if !neighbor.IsValid() {
			t.Errorf("Neighbor %d (%+v) is invalid", i, neighbor)
		}
	}
}

func TestCubeDistance(t *testing.T) {
	a := NewCubeCoord(0, 0)
	b := NewCubeCoord(3, -1)
	
	distance := a.Distance(b)
	expected := 3
	if distance != expected {
		t.Errorf("Distance from %+v to %+v = %d, expected %d", a, b, distance, expected)
	}
	
	// Distance should be symmetric
	distance2 := b.Distance(a)
	if distance != distance2 {
		t.Errorf("Distance is not symmetric: %d vs %d", distance, distance2)
	}
	
	// Distance to self should be 0
	distance3 := a.Distance(a)
	if distance3 != 0 {
		t.Errorf("Distance to self should be 0, got %d", distance3)
	}
}

func TestArrayToCubeConversion(t *testing.T) {
	// Test round-trip conversion (the exact cube coordinates depend on implementation)
	testMaps := []*Map{
		NewMap(10, 10, false), // Odd rows offset
		NewMap(10, 10, true),  // Even rows offset
	}
	
	for mapIdx, testMap := range testMaps {
		t.Logf("Testing map %d (EvenRowsOffset=%v)", mapIdx, testMap.EvenRowsOffset())
		
		// Test several positions
		testPositions := [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 0}, {2, 1}, {5, 5}}
		
		for _, pos := range testPositions {
			row, col := pos[0], pos[1]
			
			// Convert array to cube and back
			cubeCoord := testMap.ArrayToHex(row, col)
			backRow, backCol := testMap.HexToArray(cubeCoord)
			
			// Round trip should work
			if backRow != row || backCol != col {
				t.Errorf("Round trip failed for (%d, %d): got (%d, %d)", 
					row, col, backRow, backCol)
			}
			
			// Cube coordinate should be valid
			if !cubeCoord.IsValid() {
				t.Errorf("Invalid cube coordinate for (%d, %d): %+v", 
					row, col, cubeCoord)
			}
		}
	}
}

func TestMapCubeIntegration(t *testing.T) {
	// Create a test map
	testMap := NewMap(5, 5, false)
	
	// Add some tiles
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	
	// Test TileAtCube
	coord := NewCubeCoord(0, 2)
	tile := testMap.TileAtCube(coord)
	if tile == nil {
		t.Error("TileAtCube returned nil for valid coordinate")
	}
	
	// Verify the tile is at the right position
	row, col := testMap.HexToArray(coord)
	expectedTile := testMap.TileAt(row, col)
	if tile != expectedTile {
		t.Error("TileAtCube returned different tile than TileAt")
	}
	
	// Test cube neighbors
	neighbors := testMap.GetTileNeighborsCube(coord)
	if len(neighbors) != 6 {
		t.Errorf("Expected 6 neighbors, got %d", len(neighbors))
	}
	
	// Test that cube neighbors are valid tiles (don't compare with old method due to different layouts)
	for i, neighbor := range neighbors {
		if neighbor != nil {
			// Verify it's a real tile with valid coordinates
			if neighbor.Row < 0 || neighbor.Col < 0 {
				t.Errorf("Cube neighbor %d has invalid coordinates: (%d, %d)", i, neighbor.Row, neighbor.Col)
			}
		}
		// Note: Some neighbors may be nil if they're outside the map bounds, which is valid
	}
}

func TestCubeRange(t *testing.T) {
	center := NewCubeCoord(0, 0)
	
	// Test range 0 (just the center)
	range0 := center.Range(0)
	if len(range0) != 1 || range0[0] != center {
		t.Errorf("Range 0 should contain only center, got %v", range0)
	}
	
	// Test range 1 (center + 6 neighbors)
	range1 := center.Range(1)
	if len(range1) != 7 { // 1 + 6
		t.Errorf("Range 1 should contain 7 tiles, got %d", len(range1))
	}
	
	// Test range 2 (should be 19 tiles: 1 + 6 + 12)
	range2 := center.Range(2)
	if len(range2) != 19 {
		t.Errorf("Range 2 should contain 19 tiles, got %d", len(range2))
	}
	
	// All coordinates in range should be valid
	for _, coord := range range2 {
		if !coord.IsValid() {
			t.Errorf("Invalid coordinate in range: %+v", coord)
		}
	}
}

func TestCubeRing(t *testing.T) {
	center := NewCubeCoord(0, 0)
	
	// Test ring 0 (just the center)
	ring0 := center.Ring(0)
	if len(ring0) != 1 || ring0[0] != center {
		t.Errorf("Ring 0 should contain only center, got %v", ring0)
	}
	
	// Test ring 1 (6 neighbors)
	ring1 := center.Ring(1)
	if len(ring1) != 6 {
		t.Errorf("Ring 1 should contain 6 tiles, got %d", len(ring1))
	}
	
	// Test ring 2 (12 tiles)
	ring2 := center.Ring(2)
	if len(ring2) != 12 {
		t.Errorf("Ring 2 should contain 12 tiles, got %d", len(ring2))
	}
	
	// All coordinates in ring should be valid and at correct distance
	for _, coord := range ring2 {
		if !coord.IsValid() {
			t.Errorf("Invalid coordinate in ring: %+v", coord)
		}
		distance := center.Distance(coord)
		if distance != 2 {
			t.Errorf("Coordinate %+v should be distance 2, got %d", coord, distance)
		}
	}
}