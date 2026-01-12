package server

import (
	"context"
	"log"
	"net/http"

	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/web/assets/themes"
)

// handleWorldScreenshotLive dynamically renders a world screenshot using the specified theme
// GET /worlds/{worldId}/screenshot/live?theme=fantasy
func (r *RootViewsHandler) handleWorldScreenshotLive(w http.ResponseWriter, req *http.Request) {
	worldId := req.PathValue("worldId")
	if worldId == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return
	}

	themeName := req.URL.Query().Get("theme")
	if themeName == "" {
		themeName = "default"
	}

	// Get world data from service
	client := r.LilBattleApp.ClientMgr.GetWorldsSvcClient()
	resp, err := client.GetWorld(context.Background(), &protos.GetWorldRequest{Id: worldId})
	if err != nil {
		log.Printf("Failed to get world %s: %v", worldId, err)
		http.Error(w, "World not found", http.StatusNotFound)
		return
	}

	if resp.World == nil || resp.WorldData == nil {
		http.Error(w, "World has no data", http.StatusNotFound)
		return
	}

	// Render the screenshot
	r.renderScreenshot(w, resp.WorldData.TilesMap, resp.WorldData.UnitsMap, themeName)
}

// handleGameScreenshotLive dynamically renders a game screenshot using the specified theme
// GET /games/{gameId}/screenshot/live?theme=fantasy
func (r *RootViewsHandler) handleGameScreenshotLive(w http.ResponseWriter, req *http.Request) {
	gameId := req.PathValue("gameId")
	if gameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	themeName := req.URL.Query().Get("theme")
	if themeName == "" {
		themeName = "default"
	}

	// Get game data from service
	client := r.LilBattleApp.ClientMgr.GetGamesSvcClient()
	resp, err := client.GetGame(context.Background(), &protos.GetGameRequest{Id: gameId})
	if err != nil {
		log.Printf("Failed to get game %s: %v", gameId, err)
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if resp.Game == nil || resp.State == nil || resp.State.WorldData == nil {
		http.Error(w, "Game has no state data", http.StatusNotFound)
		return
	}

	// Render the screenshot
	r.renderScreenshot(w, resp.State.WorldData.TilesMap, resp.State.WorldData.UnitsMap, themeName)
}

// renderScreenshot renders tiles and units using the specified theme
func (r *RootViewsHandler) renderScreenshot(w http.ResponseWriter, tiles map[string]*protos.Tile, units map[string]*protos.Unit, themeName string) {
	// Create theme
	re := lib.DefaultRulesEngine()
	theme, err := themes.CreateTheme(themeName, re.GetCityTerrains())
	if err != nil {
		log.Printf("Failed to create theme %s: %v", themeName, err)
		http.Error(w, "Invalid theme", http.StatusBadRequest)
		return
	}

	// Create renderer for this theme
	renderer, err := themes.CreateWorldRenderer(theme)
	if err != nil {
		log.Printf("Failed to create renderer for theme %s: %v", themeName, err)
		http.Error(w, "Failed to create renderer", http.StatusInternalServerError)
		return
	}

	// Render the image
	imageBytes, contentType, err := renderer.Render(tiles, units, nil)
	if err != nil {
		log.Printf("Failed to render screenshot: %v", err)
		http.Error(w, "Failed to render screenshot", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache") // Don't cache live screenshots
	w.WriteHeader(http.StatusOK)
	w.Write(imageBytes)
}
