package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

func (r *RootViewsHandler) setupMapsMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ViewRenderer(Copier(&MapListingPage{}), ""))
	mux.HandleFunc("/new", r.createNewMapHandler)
	mux.HandleFunc("/{mapId}/view", r.ViewRenderer(Copier(&MapDetailsPage{}), ""))
	mux.HandleFunc("/{mapId}/edit", r.ViewRenderer(Copier(&MapEditorPage{}), ""))
	mux.HandleFunc("/{mapId}/start", func(w http.ResponseWriter, r *http.Request) {
		mapId := r.PathValue("mapId")
		redirectURL := fmt.Sprintf("/games/new?mapId=%s", mapId)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})
	mux.HandleFunc("/{mapId}/copy", func(w http.ResponseWriter, r *http.Request) {
		notationId := r.PathValue("notationId")
		http.Redirect(w, r, fmt.Sprintf("/appitems/new?copyFrom=%s", notationId), http.StatusFound)
	})
	mux.HandleFunc("/{mapId}", r.handleMapActions)
	return mux
}

// createNewMapHandler creates a new map and redirects to the edit page
func (r *RootViewsHandler) createNewMapHandler(w http.ResponseWriter, req *http.Request) {
	// Get logged in user ID
	loggedInUserId := r.Context.AuthMiddleware.GetLoggedInUserId(req)
	
	// For now, allow anonymous map creation (following existing pattern)
	// if loggedInUserId == "" {
	//     http.Redirect(w, req, "/login?callbackURL=/maps/new", http.StatusSeeOther)
	//     return
	// }
	
	// Get maps service client
	client, err := r.Context.ClientMgr.GetMapsSvcClient()
	if err != nil {
		log.Printf("Failed to get maps service client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Create a new map with minimal data
	createReq := &protos.CreateMapRequest{
		Map: &protos.Map{
			Name:        "Untitled Map",
			Description: "",
			CreatorId:   loggedInUserId,
			Tags:        []string{},
			Difficulty:  "",
		},
	}
	
	// Call CreateMap service (will generate new ID automatically)
	resp, err := client.CreateMap(context.Background(), createReq)
	if err != nil {
		log.Printf("Failed to create map: %v", err)
		http.Error(w, "Failed to create map", http.StatusInternalServerError)
		return
	}
	
	// Redirect to the edit page for the newly created map
	editURL := fmt.Sprintf("/maps/%s/edit", resp.Map.Id)
	http.Redirect(w, req, editURL, http.StatusFound)
}

// handleMapActions handles different HTTP methods for map operations
func (r *RootViewsHandler) handleMapActions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		r.deleteMapHandler(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// deleteMapHandler deletes a map and redirects to the listing page
func (r *RootViewsHandler) deleteMapHandler(w http.ResponseWriter, req *http.Request) {
	
	// Get map ID from URL path
	mapId := req.PathValue("mapId")
	if mapId == "" {
		http.Error(w, "Map ID is required", http.StatusBadRequest)
		return
	}
	
	// Get logged in user ID for authorization (optional for now)
	loggedInUserId := r.Context.AuthMiddleware.GetLoggedInUserId(req)
	log.Printf("Delete map request: mapId=%s, userId=%s", mapId, loggedInUserId)
	
	// Get maps service client
	client, err := r.Context.ClientMgr.GetMapsSvcClient()
	if err != nil {
		log.Printf("Failed to get maps service client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Create delete request
	deleteReq := &protos.DeleteMapRequest{
		Id: mapId,
	}
	
	// Call DeleteMap service
	_, err = client.DeleteMap(context.Background(), deleteReq)
	if err != nil {
		log.Printf("Failed to delete map %s: %v", mapId, err)
		http.Error(w, "Failed to delete map", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Successfully deleted map: %s", mapId)
	
	// Redirect back to the maps listing page
	http.Redirect(w, req, "/maps/", http.StatusFound)
}
