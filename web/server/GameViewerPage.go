package server

import (
	"context"
	"log"
	"net/http"
	"strconv"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type GameViewerPage struct {
	BasePage
	Header Header
	Map    *protos.Map
	MapId  string
	
	// Game creation parameters from URL
	PlayerCount      int
	MaxTurns        int
	UnitRestrictions map[string]string // unitId -> restriction level
}

func (p *GameViewerPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.MapId = r.PathValue("gameId") // gameId is actually mapId for now
	if p.MapId == "" {
		http.Error(w, "Map ID is required", http.StatusBadRequest)
		return nil, true
	}

	p.Title = "Game Viewer"
	p.Header.Load(r, w, vc)

	// Parse game creation parameters from query string
	p.parseGameParameters(r)

	// Load the map (same as MapEditorPage)
	client, err := vc.ClientMgr.GetMapsSvcClient()
	if err != nil {
		log.Printf("Error getting Maps client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil, true
	}

	req := &protos.GetMapRequest{
		Id: p.MapId,
	}

	resp, err := client.GetMap(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching Map %s: %v", p.MapId, err)
		http.Error(w, "Map not found", http.StatusNotFound)
		return nil, true
	}

	if resp.Map != nil {
		p.Map = resp.Map
		p.Title = "Playing: " + p.Map.Name
	}

	log.Printf("GameViewerPage loaded - MapId: %s, Players: %d, MaxTurns: %d", 
		p.MapId, p.PlayerCount, p.MaxTurns)

	return nil, false
}

// parseGameParameters extracts game creation parameters from URL query string
func (p *GameViewerPage) parseGameParameters(r *http.Request) {
	query := r.URL.Query()
	
	// Parse player count
	if playerCountStr := query.Get("playerCount"); playerCountStr != "" {
		if count, err := strconv.Atoi(playerCountStr); err == nil {
			p.PlayerCount = count
		} else {
			p.PlayerCount = 2 // default
		}
	} else {
		p.PlayerCount = 2
	}
	
	// Parse max turns
	if maxTurnsStr := query.Get("maxTurns"); maxTurnsStr != "" {
		if turns, err := strconv.Atoi(maxTurnsStr); err == nil {
			p.MaxTurns = turns
		} else {
			p.MaxTurns = 0 // unlimited
		}
	}
	
	// Parse unit restrictions (format: unitId=restriction&unitId2=restriction2)
	p.UnitRestrictions = make(map[string]string)
	for key, values := range query {
		if key != "playerCount" && key != "maxTurns" && len(values) > 0 {
			p.UnitRestrictions[key] = values[0]
		}
	}
}

func (p *GameViewerPage) Copy() View {
	return &GameViewerPage{}
}