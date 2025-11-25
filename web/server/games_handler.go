package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	protos "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

func (r *RootViewsHandler) setupGamesMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ViewRenderer(Copier(&GameListingPage{}), ""))
	mux.HandleFunc("/new", r.ViewRenderer(Copier(&StartGamePage{}), ""))
	mux.HandleFunc("/{gameId}/view", r.gameViewerHandler) // Use custom handler for layout detection
	mux.HandleFunc("/{gameId}/copy", func(w http.ResponseWriter, r *http.Request) {
		gameId := r.PathValue("gameId")
		http.Redirect(w, r, fmt.Sprintf("/games/new?copyFrom=%s", gameId), http.StatusFound)
	})
	// Screenshots now served via /screenshots/ static handler
	mux.HandleFunc("/{gameId}/screenshot/live", r.handleGameScreenshotLive)
	mux.HandleFunc("/{gameId}", r.handleGameActions)
	return mux
}

// gameViewerHandler detects layout preference and serves appropriate template
func (r *RootViewsHandler) gameViewerHandler(w http.ResponseWriter, req *http.Request) {
	// Detect layout preference
	layout := detectLayoutPreference(req)

	// Map layout to template name
	templateName := "GameViewerPageDockView"
	if true {
		switch layout {
		case "dockview":
			templateName = "GameViewerPageDockView"
		case "mobile":
			templateName = "GameViewerPageMobile" // Future
		case "grid":
			templateName = "GameViewerPageGrid"
		}
	}

	log.Printf("GameViewer: Using layout=%s, template=%s", layout, templateName)

	// Render with selected template
	r.RenderView(Copier(&GameViewerPage{})(), templateName, req, w)
}

// detectLayoutPreference determines which layout to use based on request
func detectLayoutPreference(r *http.Request) string {
	// 1. Check query param (for testing/debugging)
	if layout := r.URL.Query().Get("layout"); layout != "" {
		return layout
	}

	// 2. Check cookie/session preference (future)
	if cookie, err := r.Cookie("layout_preference"); err == nil {
		return cookie.Value
	}

	// 3. User-Agent detection (future - mobile detection)
	// ua := r.Header.Get("User-Agent")
	// if isMobileDevice(ua) {
	//     return "base" // or "mobile" when ready
	// }

	// 4. Default to DockView for backward compatibility
	return "dockview"
}

// handleGameActions handles different HTTP methods for game operations
func (r *RootViewsHandler) handleGameActions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		r.deleteGameHandler(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// deleteGameHandler deletes a game and redirects to the listing page
func (r *RootViewsHandler) deleteGameHandler(w http.ResponseWriter, req *http.Request) {
	// Get game ID from URL path
	gameId := req.PathValue("gameId")
	if gameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	// Get logged in user ID for authorization (optional for now)
	loggedInUserId := r.Context.AuthMiddleware.GetLoggedInUserId(req)
	log.Printf("Delete game request: gameId=%s, userId=%s", gameId, loggedInUserId)

	// Get games service client
	client := r.Context.ClientMgr.GetGamesSvcClient()

	// Create delete request
	deleteReq := &protos.DeleteGameRequest{
		Id: gameId,
	}

	// Call DeleteGame service
	_, err := client.DeleteGame(context.Background(), deleteReq)
	if err != nil {
		log.Printf("Failed to delete game %s: %v", gameId, err)
		http.Error(w, "Failed to delete game", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted game: %s", gameId)

	// Redirect back to the games listing page
	http.Redirect(w, req, "/games/", http.StatusFound)
}
