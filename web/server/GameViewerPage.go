package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

type GameViewerPage struct {
	BasePage
	Header      Header
	GameId      string
	WorldId     string
	Game        *protos.Game
	GameState   *protos.GameState
	GameHistory *protos.GameMoveHistory
	// World  *protos.World

	// Game creation parameters from URL
	PlayerCount      int
	MaxTurns         int
	UnitRestrictions map[string]string // unitId -> restriction level
}

func (p *GameViewerPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.GameId = r.PathValue("gameId") // gameId is actually worldId for now
	if p.GameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return nil, true
	}

	// Load the world (same as WorldEditorPage)
	client, err := vc.ClientMgr.GetGamesSvcClient()
	if err != nil {
		log.Printf("Error getting Games client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil, true
	}

	req := &protos.GetGameRequest{Id: p.GameId}

	resp, err := client.GetGame(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching Game %s: %v", p.GameId, err)
		http.Error(w, "Game not found", http.StatusNotFound)
		return nil, true
	}

	if resp.Game != nil {
		p.Game = resp.Game
		p.Title = "Playing: " + p.Game.Name
		// Set player count from game configuration
		if p.Game.Config != nil && p.Game.Config.Players != nil {
			p.PlayerCount = len(p.Game.Config.Players)
		}
	}

	if resp.State != nil {
		p.GameState = resp.State
	}

	if resp.History != nil {
		p.GameHistory = resp.History
	}

	log.Printf("GameViewerPage loaded - GameId: %s, Players: %d, MaxTurns: %d",
		p.WorldId, p.PlayerCount, p.MaxTurns)

	return nil, false
}

// GetTerrainDataJSON returns terrain data from rules engine as JSON string
func (p *GameViewerPage) GetTerrainDataJSON() string {
	rulesEngine := weewar.DefaultRulesEngine()
	
	// Marshal each terrain definition using protojson with EmitUnpopulated for all fields
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false, // Use JSON names (camelCase)
	}
	
	terrainMap := make(map[string]json.RawMessage)
	for id, terrain := range rulesEngine.Terrains {
		terrainJSON, err := marshaler.Marshal(terrain)
		if err != nil {
			log.Printf("Error marshaling terrain %d: %v", id, err)
			continue
		}
		terrainMap[fmt.Sprintf("%d", id)] = json.RawMessage(terrainJSON)
	}
	
	terrainData, err := json.Marshal(terrainMap)
	if err != nil {
		log.Printf("Error marshaling terrain data: %v", err)
		return "{}"
	}
	return string(terrainData)
}

// GetUnitDataJSON returns unit data from rules engine as JSON string
func (p *GameViewerPage) GetUnitDataJSON() string {
	rulesEngine := weewar.DefaultRulesEngine()
	
	// Marshal each unit definition using protojson with EmitUnpopulated for all fields
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false, // Use JSON names (camelCase)
	}
	
	unitMap := make(map[string]json.RawMessage)
	for id, unit := range rulesEngine.Units {
		unitJSON, err := marshaler.Marshal(unit)
		if err != nil {
			log.Printf("Error marshaling unit %d: %v", id, err)
			continue
		}
		unitMap[fmt.Sprintf("%d", id)] = json.RawMessage(unitJSON)
	}
	
	unitData, err := json.Marshal(unitMap)
	if err != nil {
		log.Printf("Error marshaling unit data: %v", err)
		return "{}"
	}
	return string(unitData)
}

// GetMovementMatrixJSON returns movement cost matrix as JSON string
func (p *GameViewerPage) GetMovementMatrixJSON() string {
	rulesEngine := weewar.DefaultRulesEngine()
	
	// Use protojson for consistent camelCase field names
	movementData, err := protojson.Marshal(rulesEngine.MovementMatrix)
	if err != nil {
		log.Printf("Error marshaling movement matrix: %v", err)
		return "{}"
	}
	return string(movementData)
}

func (p *GameViewerPage) Copy() View {
	return &GameViewerPage{}
}

/*
func (p *GameViewerPage) LoadPost(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.GameId = r.PathValue("gameId") // gameId is actually worldId for now
	if p.GameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
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
*/
