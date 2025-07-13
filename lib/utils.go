package weewar

import (
	"math"

	"github.com/tdewolff/canvas"
)

// Helper function to create circle points
func createCirclePoints(centerX, centerY, radius float64, segments int) []Point {
	points := make([]Point, segments)
	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		points[i] = Point{X: x, Y: y}
	}
	return points
}

// Simple math helpers (since we can't import math in WASM easily)
func cos(angle float64) float64 {
	// Simple cosine approximation using Taylor series
	// cos(x) ≈ 1 - x²/2! + x⁴/4! - x⁶/6!
	x := angle
	for x > 3.14159*2 {
		x -= 3.14159 * 2
	}
	for x < 0 {
		x += 3.14159 * 2
	}

	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

func sin(angle float64) float64 {
	// sin(x) = cos(x - π/2)
	return cos(angle - 3.14159/2)
}

// createHexPoints creates points for a hexagon centered at (cx, cy) with given radius
func createHexPoints(cx, cy, radius float64) []Point {
	points := make([]Point, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * 60.0 * 3.14159 / 180.0 // Convert to radians
		x := cx + radius*cos(angle)
		y := cy + radius*sin(angle)
		points[i] = Point{X: x, Y: y}
	}
	return points
}

// createHexagonPath creates a hexagon path centered at (cx, cy) with given radius
func createHexagonPath(cx, cy, radius float64) *canvas.Path {
	path := &canvas.Path{}

	// Create hexagon with 6 sides
	for i := 0; i < 6; i++ {
		// Angle for each vertex (60 degrees apart)
		angle := float64(i) * 60.0 * 3.14159 / 180.0
		x := cx + radius*cos(angle)
		y := cy + radius*sin(angle)

		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()

	return path
}
