package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

type GameViewerPage struct {
	BasePage
	Header  Header
	World   *protos.World
	WorldId string

	// Game creation parameters from URL
	PlayerCount      int
	MaxTurns         int
	UnitRestrictions map[string]string // unitId -> restriction level
}

func (p *GameViewerPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.WorldId = r.PathValue("gameId") // gameId is actually worldId for now
	if p.WorldId == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return nil, true
	}

	p.Title = "Game Viewer"
	p.Header.Load(r, w, vc)

	// Parse game creation parameters from query string
	p.parseGameParameters(r)

	// Load the world (same as WorldEditorPage)
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
		p.World = resp.World
		p.Title = "Playing: " + p.World.Name
	}

	log.Printf("GameViewerPage loaded - WorldId: %s, Players: %d, MaxTurns: %d",
		p.WorldId, p.PlayerCount, p.MaxTurns)

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

// GetTerrainDataJSON returns terrain data from rules engine as JSON string
func (p *GameViewerPage) GetTerrainDataJSON() string {
	rulesEngine := weewar.DefaultRulesEngine()
	terrainData, err := json.Marshal(rulesEngine.Terrains)
	// fmt.Println("RuleEngine: ", rulesEngine)
	// fmt.Println("TerrainData: ", terrainData)
	if err != nil {
		log.Printf("Error marshaling terrain data: %v", err)
		return "{}"
	}
	return string(terrainData)
}

// GetUnitDataJSON returns unit data from rules engine as JSON string
func (p *GameViewerPage) GetUnitDataJSON() string {
	rulesEngine := weewar.DefaultRulesEngine()
	unitData, err := json.Marshal(rulesEngine.Units)
	if err != nil {
		log.Printf("Error marshaling unit data: %v", err)
		return "{}"
	}
	return string(unitData)
}

// GetMovementMatrixJSON returns movement cost matrix as JSON string
func (p *GameViewerPage) GetMovementMatrixJSON() string {
	rulesEngine := weewar.DefaultRulesEngine()
	movementData, err := json.Marshal(rulesEngine.MovementMatrix)
	if err != nil {
		log.Printf("Error marshaling movement matrix: %v", err)
		return "{}"
	}
	return string(movementData)
}

func (p *GameViewerPage) Copy() View {
	return &GameViewerPage{}
}
