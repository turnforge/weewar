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
var cleanupBufferTestFiles = true

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
