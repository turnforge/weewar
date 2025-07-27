package server

import (
	"context"
	"log"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type WorldDetailsPage struct {
	BasePage
	Header  Header
	World   *protos.World // Use the same type as WorldEditorPage for consistency
	WorldId string
}

func (p *WorldDetailsPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.WorldId = r.PathValue("worldId")
	if p.WorldId == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return nil, true
	}

	p.Title = "World Details"
	p.Header.Load(r, w, vc)

	// Fetch the World using the client manager
	client, err := vc.ClientMgr.GetWorldsSvcClient()
	if err != nil {
		log.Printf("Error getting Worlds client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil, true
	}

	req := &protos.GetWorldRequest{
		Id: p.WorldId,
	}

	resp, err := client.GetWorld(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching World %s: %v", p.WorldId, err)
		http.Error(w, "World not found", http.StatusNotFound)
		return nil, true
	}

	if resp.World != nil {
		// Use the World data for display
		p.World = resp.World
		p.Title = p.World.Name
	}

	return nil, false
}


func (p *WorldDetailsPage) Copy() View {
	return &WorldDetailsPage{}
}
