package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	"google.golang.org/protobuf/encoding/protojson"

	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
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

	WasmExecJsPath string
	WasmBundlePath string
}

func (p *GameViewerPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	p.WasmExecJsPath = "/static/wasm/wasm_exec.js"
	p.WasmBundlePath = "/static/wasm/lilbattle-cli.wasm"
	useTinyGo := getQueryOrDefaultStr(r.URL.Query(), "tinygo", "")
	if useTinyGo == "true" { // true for TinyGo
		p.WasmExecJsPath = "/static/wasm/wasm_exec_tiny.js"
		p.WasmBundlePath = "/static/wasm/lilbattle-cli-tinygo.wasm"
	}

	p.GameId = r.PathValue("gameId") // gameId is actually worldId for now
	if p.GameId == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return nil, true
	}

	// Load the world (same as WorldEditorPage)
	ctx := app.Context
	client := ctx.ClientMgr.GetGamesSvcClient()

	req := &protos.GetGameRequest{Id: p.GameId}

	resp, err := client.GetGame(context.Background(), req)
	if err != nil {
		log.Printf("Error fetching Game %s: %v", p.GameId, err)
		return HandleGRPCError(err, w, r, app)
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
	rulesEngine := lib.DefaultRulesEngine()

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
	rulesEngine := lib.DefaultRulesEngine()

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

// GetTerrainUnitPropertiesJSON returns terrain-unit interaction properties as JSON string
func (p *GameViewerPage) GetTerrainUnitPropertiesJSON() string {
	rulesEngine := lib.DefaultRulesEngine()

	// Marshal each terrain-unit property using protojson with camelCase
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false, // Use JSON names (camelCase)
	}

	terrainUnitMap := make(map[string]json.RawMessage)
	for key, props := range rulesEngine.TerrainUnitProperties {
		propsJSON, err := marshaler.Marshal(props)
		if err != nil {
			log.Printf("Error marshaling terrain-unit property %s: %v", key, err)
			continue
		}
		terrainUnitMap[key] = json.RawMessage(propsJSON)
	}

	terrainUnitData, err := json.Marshal(terrainUnitMap)
	if err != nil {
		log.Printf("Error marshaling terrain-unit properties: %v", err)
		return "{}"
	}
	return string(terrainUnitData)
}

// GetUnitUnitPropertiesJSON returns unit-vs-unit combat properties as JSON string
func (p *GameViewerPage) GetUnitUnitPropertiesJSON() string {
	rulesEngine := lib.DefaultRulesEngine()

	// Marshal each unit-unit property using protojson with camelCase
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false, // Use JSON names (camelCase)
	}

	unitUnitMap := make(map[string]json.RawMessage)
	for key, props := range rulesEngine.UnitUnitProperties {
		propsJSON, err := marshaler.Marshal(props)
		if err != nil {
			log.Printf("Error marshaling unit-unit property %s: %v", key, err)
			continue
		}
		unitUnitMap[key] = json.RawMessage(propsJSON)
	}

	unitUnitData, err := json.Marshal(unitUnitMap)
	if err != nil {
		log.Printf("Error marshaling unit-unit properties: %v", err)
		return "{}"
	}
	return string(unitUnitData)
}
