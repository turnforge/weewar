package server

import (
	"context"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type WorldViewerPage struct {
	BasePage
	Header    Header
	World     *protos.World
	WorldData *protos.WorldData
	WorldId   string
}

func (p *WorldViewerPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	p.WorldId = r.PathValue("worldId")
	if p.WorldId == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return nil, true
	}

	p.Title = "World Details"
	p.Header.Load(r, w, app)

	ctx := app.Context
	client := ctx.ClientMgr.GetWorldsSvcClient()
	req := &protos.GetWorldRequest{Id: p.WorldId}

	resp, err := client.GetWorld(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching World %s: %v", p.WorldId, err)
		return HandleGRPCError(err, w, r, app)
	}

	if resp.World != nil {
		p.World = resp.World
		p.WorldData = resp.WorldData
		p.Title = p.World.Name
	}

	return nil, false
}
