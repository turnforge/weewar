package server

import (
	"context"
	"log"
	"net/http"

	protos "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

type GameDetailPage struct {
	BasePage
	Header Header
	Game   *protos.Game
	GameId string
}

func (p *GameDetailPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.GameId = r.PathValue("appItemId")
	if p.GameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return nil, true
	}

	p.Title = "Game Details"
	p.Header.Load(r, w, vc)

	// Fetch the Game using the client manager
	client := vc.ClientMgr.GetGamesSvcClient()

	req := &protos.GetGameRequest{
		Id: p.GameId,
	}

	resp, err := client.GetGame(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching Game %s: %v", p.GameId, err)
		http.Error(w, "Game not found", http.StatusNotFound)
		return nil, true
	}

	if resp.Game != nil {
		// Convert from GameProject to Game (assuming we need the basic info)
		p.Game = &protos.Game{
			Id:          resp.Game.Id,
			Name:        resp.Game.Name,
			Description: resp.Game.Description,
			CreatedAt:   resp.Game.CreatedAt,
			UpdatedAt:   resp.Game.UpdatedAt,
		}
		p.Title = p.Game.Name
	}

	return nil, false
}

func (p *GameDetailPage) Copy() View {
	return &GameDetailPage{}
}
