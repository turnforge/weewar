package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

func (r *RootViewsHandler) setupGamesMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ViewRenderer(Copier(&GameListingPage{}), ""))
	mux.HandleFunc("/new", r.ViewRenderer(Copier(&StartGamePage{}), ""))
	mux.HandleFunc("/{gameId}/view", r.ViewRenderer(Copier(&GameViewerPage{}), ""))
	mux.HandleFunc("/{gameId}/copy", func(w http.ResponseWriter, r *http.Request) {
		notationId := r.PathValue("notationId")
		http.Redirect(w, r, fmt.Sprintf("/appitems/new?copyFrom=%s", notationId), http.StatusFound)
	})
	mux.HandleFunc("/{gameId}", r.handleGameActions)
	return mux
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
	client, err := r.Context.ClientMgr.GetGamesSvcClient()
	if err != nil {
		log.Printf("Failed to get games service client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create delete request
	deleteReq := &protos.DeleteGameRequest{
		Id: gameId,
	}

	// Call DeleteGame service
	_, err = client.DeleteGame(context.Background(), deleteReq)
	if err != nil {
		log.Printf("Failed to delete game %s: %v", gameId, err)
		http.Error(w, "Failed to delete game", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted game: %s", gameId)

	// Redirect back to the games listing page
	http.Redirect(w, req, "/games/", http.StatusFound)
}
