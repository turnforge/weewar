package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

func (r *RootViewsHandler) setupWorldsMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ViewRenderer(Copier(&WorldListingPage{}), ""))
	mux.HandleFunc("/new", r.createNewWorldHandler)
	mux.HandleFunc("/{worldId}/view", r.ViewRenderer(Copier(&WorldViewerPage{}), ""))
	mux.HandleFunc("/{worldId}/edit", r.ViewRenderer(Copier(&WorldEditorPage{}), ""))
	mux.HandleFunc("/{worldId}/start", func(w http.ResponseWriter, r *http.Request) {
		worldId := r.PathValue("worldId")
		redirectURL := fmt.Sprintf("/games/new?worldId=%s", worldId)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})
	mux.HandleFunc("/{worldId}/copy", func(w http.ResponseWriter, r *http.Request) {
		notationId := r.PathValue("notationId")
		http.Redirect(w, r, fmt.Sprintf("/appitems/new?copyFrom=%s", notationId), http.StatusFound)
	})
	mux.HandleFunc("/{worldId}", r.handleWorldActions)
	return mux
}

// createNewWorldHandler creates a new world and redirects to the edit page
func (r *RootViewsHandler) createNewWorldHandler(w http.ResponseWriter, req *http.Request) {
	// Get logged in user ID
	loggedInUserId := r.Context.AuthMiddleware.GetLoggedInUserId(req)

	// For now, allow anonymous world creation (following existing pattern)
	// if loggedInUserId == "" {
	//     http.Redirect(w, req, "/login?callbackURL=/worlds/new", http.StatusSeeOther)
	//     return
	// }

	// Get worlds service client
	client, err := r.Context.ClientMgr.GetWorldsSvcClient()
	if err != nil {
		log.Printf("Failed to get worlds service client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create a new world with minimal data
	createReq := &protos.CreateWorldRequest{
		World: &protos.World{
			Name:        "Untitled World",
			Description: "",
			CreatorId:   loggedInUserId,
			Tags:        []string{},
			Difficulty:  "",
		},
	}

	// Call CreateWorld service (will generate new ID automatically)
	resp, err := client.CreateWorld(context.Background(), createReq)
	if err != nil {
		log.Printf("Failed to create world: %v", err)
		http.Error(w, "Failed to create world", http.StatusInternalServerError)
		return
	}

	// Redirect to the edit page for the newly created world
	editURL := fmt.Sprintf("/worlds/%s/edit", resp.World.Id)
	http.Redirect(w, req, editURL, http.StatusFound)
}

// handleWorldActions handles different HTTP methods for world operations
func (r *RootViewsHandler) handleWorldActions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		r.deleteWorldHandler(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// deleteWorldHandler deletes a world and redirects to the listing page
func (r *RootViewsHandler) deleteWorldHandler(w http.ResponseWriter, req *http.Request) {

	// Get world ID from URL path
	worldId := req.PathValue("worldId")
	if worldId == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return
	}

	// Get logged in user ID for authorization (optional for now)
	loggedInUserId := r.Context.AuthMiddleware.GetLoggedInUserId(req)
	log.Printf("Delete world request: worldId=%s, userId=%s", worldId, loggedInUserId)

	// Get worlds service client
	client, err := r.Context.ClientMgr.GetWorldsSvcClient()
	if err != nil {
		log.Printf("Failed to get worlds service client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create delete request
	deleteReq := &protos.DeleteWorldRequest{
		Id: worldId,
	}

	// Call DeleteWorld service
	_, err = client.DeleteWorld(context.Background(), deleteReq)
	if err != nil {
		log.Printf("Failed to delete world %s: %v", worldId, err)
		http.Error(w, "Failed to delete world", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted world: %s", worldId)

	// Redirect back to the worlds listing page
	http.Redirect(w, req, "/worlds/", http.StatusFound)
}
