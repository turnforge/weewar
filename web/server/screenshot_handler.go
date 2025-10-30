package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// handleResourceScreenshot handles screenshot operations for any resource type (games, worlds, etc.)
func (r *RootViewsHandler) handleResourceScreenshot(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		resourceId := req.PathValue(resourceType + "Id")
		if resourceId == "" {
			http.Error(w, "Resource ID is required", http.StatusBadRequest)
			return
		}

		// Construct path to screenshot
		screenshotPath := filepath.Join(
			os.Getenv("HOME"),
			"dev-app-data",
			"weewar",
			"storage",
			resourceType+"s", // games or worlds
			resourceId,
			"screenshots",
			"screenshot.png",
		)

		switch req.Method {
		case http.MethodGet:
			r.getScreenshot(w, req, screenshotPath)
		case http.MethodPost:
			r.saveScreenshot(w, req, screenshotPath, resourceType, resourceId)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// getScreenshot serves the screenshot file
func (r *RootViewsHandler) getScreenshot(w http.ResponseWriter, req *http.Request, screenshotPath string) {
	// Check if screenshot exists
	if _, err := os.Stat(screenshotPath); os.IsNotExist(err) {
		http.Error(w, "Screenshot not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, req, screenshotPath)
}

// saveScreenshot saves the uploaded screenshot
func (r *RootViewsHandler) saveScreenshot(w http.ResponseWriter, req *http.Request, screenshotPath, resourceType, resourceId string) {
	// Parse multipart form (10MB max)
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, _, err := req.FormFile("screenshot")
	if err != nil {
		log.Printf("Failed to get screenshot file: %v", err)
		http.Error(w, "Screenshot file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create screenshots directory if it doesn't exist
	screenshotDir := filepath.Dir(screenshotPath)
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		log.Printf("Failed to create screenshots directory: %v", err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Create destination file
	dst, err := os.Create(screenshotPath)
	if err != nil {
		log.Printf("Failed to create screenshot file: %v", err)
		http.Error(w, "Failed to save screenshot", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Failed to write screenshot: %v", err)
		http.Error(w, "Failed to save screenshot", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved screenshot for %s: %s", resourceType, resourceId)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}
