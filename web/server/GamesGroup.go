package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// GamesGroup implements goal.PageGroup for /games routes.
type GamesGroup struct {
	lilbattleApp *LilBattleApp
	goalApp   *goal.App[*LilBattleApp]
}

// RegisterRoutes registers all game-related routes using goal.Register.
func (g *GamesGroup) RegisterRoutes(app *goal.App[*LilBattleApp]) *http.ServeMux {
	mux := http.NewServeMux()

	// Register pages using goal's generic registration
	goal.Register[*GameListingPage](app, mux, "/")
	goal.Register[*StartGamePage](app, mux, "/new")

	// Game viewer uses custom handler for layout detection
	mux.HandleFunc("/{gameId}/view", g.gameViewerHandler)
	mux.HandleFunc("/{gameId}/copy", gameCopyHandler)
	// Screenshot handler delegates to LilBattleApp's ViewsRoot method
	mux.HandleFunc("/{gameId}/screenshot/live", g.lilbattleApp.ViewsRoot.handleGameScreenshotLive)
	mux.HandleFunc("/{gameId}", gameActionsHandler(app))

	return mux
}

// gameViewerHandler detects layout preference and serves appropriate template
func (g *GamesGroup) gameViewerHandler(w http.ResponseWriter, req *http.Request) {
	layout := detectLayoutPreference(req)

	templateName := "GameViewerPageDockView"
	switch layout {
	case "dockview":
		templateName = "GameViewerPageDockView"
	case "mobile":
		templateName = "GameViewerPageMobile"
	case "grid":
		templateName = "GameViewerPageGrid"
	}

	log.Printf("GameViewer: Using layout=%s, template=%s", layout, templateName)

	// Create and load view
	view := &GameViewerPage{}
	err, finished := view.Load(req, w, g.goalApp)
	if finished {
		return
	}
	if err != nil {
		log.Printf("View load error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render with selected template
	fileName, blockName := goal.ParseTemplateSpec(templateName)
	if renderErr := g.goalApp.RenderTemplate(w, fileName, blockName, view); renderErr != nil {
		http.Error(w, "Template render error", http.StatusInternalServerError)
	}
}

// detectLayoutPreference determines which layout to use based on request
func detectLayoutPreference(r *http.Request) string {
	if layout := r.URL.Query().Get("layout"); layout != "" {
		return layout
	}
	if cookie, err := r.Cookie("layout_preference"); err == nil {
		return cookie.Value
	}
	return "dockview"
}

func gameCopyHandler(w http.ResponseWriter, r *http.Request) {
	gameId := r.PathValue("gameId")
	http.Redirect(w, r, fmt.Sprintf("/games/new?copyFrom=%s", gameId), http.StatusFound)
}

func gameActionsHandler(app *goal.App[*LilBattleApp]) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodDelete:
			deleteGameHandler(app, w, req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func deleteGameHandler(app *goal.App[*LilBattleApp], w http.ResponseWriter, req *http.Request) {
	gameId := req.PathValue("gameId")
	if gameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	ctx := app.Context
	loggedInUserId := ctx.AuthMiddleware.GetLoggedInUserId(req)
	log.Printf("Delete game request: gameId=%s, userId=%s", gameId, loggedInUserId)

	client := ctx.ClientMgr.GetGamesSvcClient()
	deleteReq := &protos.DeleteGameRequest{Id: gameId}

	_, err := client.DeleteGame(context.Background(), deleteReq)
	if err != nil {
		log.Printf("Failed to delete game %s: %v", gameId, err)
		http.Error(w, "Failed to delete game", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted game: %s", gameId)
	http.Redirect(w, req, "/games/", http.StatusFound)
}
