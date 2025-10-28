package rendering

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

// createHexPoints creates points for a hexagon centered at (cx, cy) with given radius
func createHexPoints(cx, cy, radius float64) []Point {
	points := make([]Point, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * 60.0 * 3.14159 / 180.0 // Convert to radians
		x := cx + radius*math.Cos(angle)
		y := cy + radius*math.Sin(angle)
		points[i] = Point{X: x, Y: y}
	}
	return points
}

// createHexagonPath creates a hexagon path centered at (cx, cy) with given radius
func CreateHexagonPath(cx, cy, radius float64) *canvas.Path {
	path := &canvas.Path{}

	// Create hexagon with 6 sides
	for i := 0; i < 6; i++ {
		// Angle for each vertex (60 degrees apart)
		angle := float64(i) * 60.0 * 3.14159 / 180.0
		x := cx + radius*math.Cos(angle)
		y := cy + radius*math.Sin(angle)

		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()

	return path
}
