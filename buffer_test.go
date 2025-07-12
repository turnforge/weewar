package weewar

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Global variable to control test file cleanup (shared with core_test.go)
var cleanupBufferTestFiles = false

// Helper function to create organized test output directory for buffer tests
func createBufferTestOutputDir(testName string) string {
	timestamp := time.Now().Format("20060102_150405")
	testDir := filepath.Join("/tmp/turnengine", testName, timestamp)
	os.MkdirAll(testDir, 0755)
	return testDir
}

// Helper function to get organized test output path for buffer tests
func getBufferTestOutputPath(testName, filename string) string {
	testDir := createBufferTestOutputDir(testName)
	return filepath.Join(testDir, filename)
}

func TestBufferDrawImageScaling(t *testing.T) {
	// Create a buffer
	buffer := NewBuffer(200, 200)

	// Create a test image with a specific pattern
	testImg := image.NewRGBA(image.Rect(0, 0, 4, 4))

	// Create a checkerboard pattern
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if (x+y)%2 == 0 {
				testImg.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
			} else {
				testImg.Set(x, y, color.RGBA{0, 255, 0, 255}) // Green
			}
		}
	}

	// Draw at original size
	buffer.DrawImage(10, 10, 4, 4, testImg)

	// Draw scaled up
	buffer.DrawImage(50, 50, 40, 40, testImg)

	// Draw scaled down
	buffer.DrawImage(150, 150, 20, 20, testImg)

	// Draw with different aspect ratio
	buffer.DrawImage(100, 10, 60, 30, testImg)

	// Save result
	imagePath := getBufferTestOutputPath("TestBufferDrawImageScaling", "scaling.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save scaling test: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Scaling test PNG was not created")
	} else {
		t.Logf("Scaling test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupBufferTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferAlphaCompositing(t *testing.T) {
	// Create a buffer with white background
	buffer := NewBuffer(150, 150)

	// Create a white background
	whiteImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	whiteImg.Set(0, 0, color.RGBA{255, 255, 255, 255})
	buffer.DrawImage(0, 0, 150, 150, whiteImg)

	// Create semi-transparent images
	redImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	redImg.Set(0, 0, color.RGBA{255, 0, 0, 128}) // 50% transparent red

	blueImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	blueImg.Set(0, 0, color.RGBA{0, 0, 255, 128}) // 50% transparent blue

	greenImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	greenImg.Set(0, 0, color.RGBA{0, 255, 0, 128}) // 50% transparent green

	// Draw overlapping semi-transparent rectangles
	buffer.DrawImage(25, 25, 50, 50, redImg)   // Red square
	buffer.DrawImage(50, 50, 50, 50, blueImg)  // Blue square (overlaps red)
	buffer.DrawImage(75, 25, 50, 50, greenImg) // Green square (overlaps blue)

	// Save result
	imagePath := getBufferTestOutputPath("TestBufferAlphaCompositing", "alpha_compositing.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save alpha test: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Alpha test PNG was not created")
	} else {
		t.Logf("Alpha compositing test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupBufferTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferFillPath(t *testing.T) {
	// Create a buffer
	buffer := NewBuffer(300, 300)

	// Create a triangle path
	triangle := []Point{
		{X: 50, Y: 50},
		{X: 250, Y: 50},
		{X: 150, Y: 200},
	}

	// Fill triangle with red color
	buffer.FillPath(triangle, Color{R: 255, G: 0, B: 0, A: 255})

	// Create a pentagon path
	pentagon := []Point{
		{X: 100, Y: 220},
		{X: 140, Y: 210},
		{X: 160, Y: 240},
		{X: 140, Y: 270},
		{X: 100, Y: 270},
	}

	// Fill pentagon with semi-transparent blue
	buffer.FillPath(pentagon, Color{R: 0, G: 0, B: 255, A: 128})

	// Save result
	imagePath := getBufferTestOutputPath("TestBufferFillPath", "filled_paths.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save fill path test: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Fill path test PNG was not created")
	} else {
		t.Logf("Fill path test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupBufferTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferStrokePath(t *testing.T) {
	// Create a buffer
	buffer := NewBuffer(300, 300)

	// Create a zigzag path
	zigzag := []Point{
		{X: 50, Y: 50},
		{X: 100, Y: 100},
		{X: 150, Y: 50},
		{X: 200, Y: 100},
		{X: 250, Y: 50},
	}

	// Stroke with thick line
	thickStroke := StrokeProperties{
		Width:    10.0,
		LineCap:  "round",
		LineJoin: "round",
	}
	buffer.StrokePath(zigzag, Color{R: 255, G: 0, B: 0, A: 255}, thickStroke)

	// Create a dashed line
	dashedLine := []Point{
		{X: 50, Y: 150},
		{X: 250, Y: 150},
	}

	// Stroke with dashed pattern
	dashedStroke := StrokeProperties{
		Width:       5.0,
		LineCap:     "butt",
		LineJoin:    "miter",
		DashPattern: []float64{10, 5, 3, 5},
		DashOffset:  0,
	}
	buffer.StrokePath(dashedLine, Color{R: 0, G: 255, B: 0, A: 255}, dashedStroke)

	// Create a curved path
	curved := []Point{
		{X: 50, Y: 200},
		{X: 100, Y: 180},
		{X: 150, Y: 220},
		{X: 200, Y: 180},
		{X: 250, Y: 200},
	}

	// Stroke with square caps and bevel joins
	squareStroke := StrokeProperties{
		Width:    8.0,
		LineCap:  "square",
		LineJoin: "bevel",
	}
	buffer.StrokePath(curved, Color{R: 0, G: 0, B: 255, A: 180}, squareStroke)

	// Save result
	imagePath := getBufferTestOutputPath("TestBufferStrokePath", "stroked_paths.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save stroke path test: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Stroke path test PNG was not created")
	} else {
		t.Logf("Stroke path test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupBufferTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferPathAlphaCompositing(t *testing.T) {
	// Create a buffer with white background
	buffer := NewBuffer(200, 200)

	// Fill with white background
	background := []Point{
		{X: 0, Y: 0},
		{X: 200, Y: 0},
		{X: 200, Y: 200},
		{X: 0, Y: 200},
	}
	buffer.FillPath(background, Color{R: 255, G: 255, B: 255, A: 255})

	// Create overlapping circles (approximated with polygons)
	circle1 := createCirclePoints(70, 70, 40, 20)
	circle2 := createCirclePoints(130, 70, 40, 20)
	circle3 := createCirclePoints(100, 110, 40, 20)

	// Fill with semi-transparent colors
	buffer.FillPath(circle1, Color{R: 255, G: 0, B: 0, A: 100}) // Red
	buffer.FillPath(circle2, Color{R: 0, G: 255, B: 0, A: 100}) // Green
	buffer.FillPath(circle3, Color{R: 0, G: 0, B: 255, A: 100}) // Blue

	// Save result
	imagePath := getBufferTestOutputPath("TestBufferPathAlphaCompositing", "alpha_compositing_paths.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save alpha compositing test: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Alpha compositing test PNG was not created")
	} else {
		t.Logf("Alpha compositing test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupBufferTestFiles {
		os.Remove(imagePath)
	}
}

func TestBufferPathEdgeCases(t *testing.T) {
	// Create a buffer
	buffer := NewBuffer(100, 100)

	// Test empty path
	emptyPath := []Point{}
	buffer.FillPath(emptyPath, Color{R: 255, G: 0, B: 0, A: 255})

	// Test single point
	singlePoint := []Point{{X: 50, Y: 50}}
	buffer.FillPath(singlePoint, Color{R: 255, G: 0, B: 0, A: 255})

	// Test two points (should create a line for stroke)
	twoPoints := []Point{{X: 10, Y: 10}, {X: 90, Y: 90}}
	strokeProps := StrokeProperties{Width: 5.0}
	buffer.StrokePath(twoPoints, Color{R: 0, G: 255, B: 0, A: 255}, strokeProps)

	// Save result
	imagePath := getBufferTestOutputPath("TestBufferPathEdgeCases", "edge_cases.png")
	err := buffer.Save(imagePath)
	if err != nil {
		t.Errorf("Failed to save edge cases test: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Errorf("Edge cases test PNG was not created")
	} else {
		t.Logf("Edge cases test saved to: %s", imagePath)
	}

	// Clean up (conditional)
	if cleanupBufferTestFiles {
		os.Remove(imagePath)
	}
}
