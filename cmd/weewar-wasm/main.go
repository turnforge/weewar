//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/panyam/turnengine/games/weewar/assets"
	"github.com/panyam/turnengine/games/weewar/cmd/wasmutils"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// Global game instance
var globalGame *weewar.Game
var globalRulesEngine *weewar.RulesEngine

func main() {
	fmt.Println("WeeWar WASM module loading...")

	// Initialize rules engine from embedded data
	var err error
	globalRulesEngine, err = weewar.LoadRulesEngineFromJSON(assets.RulesDataJSON)
	if err != nil {
		fmt.Printf("Failed to load rules engine: %v", err)
		return
	}
	fmt.Println("Rules engine loaded successfully")

	// Register JavaScript functions
	js.Global().Set("weewarCreateGameFromMap", wasmutils.CreateWrapper(1, 1, func(args []js.Value) (any, error) {
		return createGameFromMap(args[0].String())
	}))
	js.Global().Set("weewarGetMovementOptions", wasmutils.CreateWrapper(2, 2, func(args []js.Value) (any, error) {
		return getMovementOptions(args[0].Int(), args[1].Int())
	}))
	js.Global().Set("weewarGetAttackOptions", wasmutils.CreateWrapper(2, 2, func(args []js.Value) (any, error) {
		return getAttackOptions(args[0].Int(), args[1].Int())
	}))
	js.Global().Set("weewarCanSelectUnit", wasmutils.CreateWrapper(2, 2, func(args []js.Value) (any, error) {
		return canSelectUnit(args[0].Int(), args[1].Int())
	}))

	js.Global().Set("weewarGetGameState", js.FuncOf(getGameState))
	js.Global().Set("weewarSelectUnit", js.FuncOf(selectUnit))
	js.Global().Set("weewarMoveUnit", js.FuncOf(moveUnit))
	js.Global().Set("weewarAttackUnit", js.FuncOf(attackUnit))
	js.Global().Set("weewarEndTurn", js.FuncOf(endTurn))
	js.Global().Set("weewarGetTerrainStatsAt", js.FuncOf(getTerrainStatsAt))
	js.Global().Set("weewarGetTileInfo", js.FuncOf(getTileInfo))

	fmt.Println("WeeWar WASM module loaded successfully")

	// Keep the program running
	c := make(chan struct{})
	<-c
}

// =============================================================================
// WASM API Functions - Thin Delegation Layer
// =============================================================================

// createGameFromMap creates a new game from web map data
func createGameFromMap(mapDataStr string) (gameState any, err error) {
	// Create world using unified JSON format from frontend
	world := weewar.NewWorld("test") // &weewar.World{}
	// world := &weewar.World{}
	if err = world.UnmarshalJSON([]byte(mapDataStr)); err != nil {
		return
	}

	// Create game using existing NewGame method
	seed := time.Now().UnixNano()
	game, err := weewar.NewGame(world, globalRulesEngine, seed)
	if err != nil {
		return nil, err
	}

	// Store game globally
	globalGame = game

	// Return game state using new UI method
	return game.GetGameStateForUI(), nil
}

// getGameState returns current game state for UI
func getGameState(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	gameState := globalGame.GetGameStateForUI()
	return createJSResponse(true, "Game state retrieved", gameState)
}

// selectUnit selects unit and returns movement/attack options
func selectUnit(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	if len(args) < 2 {
		return createJSResponse(false, "Missing coordinate arguments", nil)
	}

	q := args[0].Int()
	r := args[1].Int()
	coord := weewar.AxialCoord{Q: q, R: r}

	// Use new UI method from lib/ui.go
	unit, movable, attackable, err := globalGame.SelectUnit(coord)
	if err != nil {
		return createJSResponse(false, err.Error(), nil)
	}

	// Return data using existing types
	data := map[string]any{
		"unit":             unit,       // *Unit (already JSON-tagged)
		"movableCoords":    movable,    // []TileOption from RulesEngine
		"attackableCoords": attackable, // []AxialCoord from RulesEngine
	}

	return createJSResponse(true, "Unit selected", data)
}

// moveUnit moves a unit from one position to another
func moveUnit(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	if len(args) < 4 {
		return createJSResponse(false, "Missing coordinate arguments", nil)
	}

	fromQ := args[0].Int()
	fromR := args[1].Int()
	toQ := args[2].Int()
	toR := args[3].Int()

	from := weewar.AxialCoord{Q: fromQ, R: fromR}
	to := weewar.AxialCoord{Q: toQ, R: toR}

	// Use existing Game method
	err := globalGame.MoveUnitAt(from, to)
	if err != nil {
		return createJSResponse(false, err.Error(), nil)
	}

	// Return updated unit
	unit := globalGame.World.UnitAt(to)
	return createJSResponse(true, "Unit moved successfully", map[string]any{
		"unit": unit,
	})
}

// attackUnit performs combat between two units
func attackUnit(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	if len(args) < 4 {
		return createJSResponse(false, "Missing coordinate arguments", nil)
	}

	attackerQ := args[0].Int()
	attackerR := args[1].Int()
	defenderQ := args[2].Int()
	defenderR := args[3].Int()

	attackerPos := weewar.AxialCoord{Q: attackerQ, R: attackerR}
	defenderPos := weewar.AxialCoord{Q: defenderQ, R: defenderR}

	// Use existing Game method
	result, err := globalGame.AttackUnitAt(attackerPos, defenderPos)
	if err != nil {
		return createJSResponse(false, err.Error(), nil)
	}

	// Return CombatResult (already perfectly structured for UI)
	return createJSResponse(true, "Attack completed", result)
}

// endTurn advances to next player's turn
func endTurn(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	// Use existing Game method
	err := globalGame.EndTurn()
	if err != nil {
		return createJSResponse(false, err.Error(), nil)
	}

	// Return updated game state
	gameState := globalGame.GetGameStateForUI()
	return createJSResponse(true, "Turn ended", gameState)
}

// getTerrainStatsAt returns detailed terrain stats for a tile
func getTerrainStatsAt(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	if len(args) < 2 {
		return createJSResponse(false, "Missing coordinate arguments", nil)
	}

	q := args[0].Int()
	r := args[1].Int()

	// Use UI method from lib/ui.go
	stats, err := globalGame.GetTerrainStatsAt(q, r)
	if err != nil {
		return createJSResponse(false, err.Error(), nil)
	}

	return createJSResponse(true, "Terrain stats retrieved", stats)
}

// canSelectUnit checks if unit at position can be selected by current player
func canSelectUnit(q, r int) (any, error) {
	if globalGame == nil {
		return nil, fmt.Errorf("No game loaded")
	}

	// Use UI method from lib/ui.go
	return globalGame.CanSelectUnit(q, r), nil
}

// getTileInfo returns basic tile information
func getTileInfo(this js.Value, args []js.Value) any {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	if len(args) < 2 {
		return createJSResponse(false, "Missing coordinate arguments", nil)
	}

	q := args[0].Int()
	r := args[1].Int()

	// Use UI method from lib/ui.go
	info, err := globalGame.GetTileInfo(q, r)
	if err != nil {
		return createJSResponse(false, err.Error(), nil)
	}

	return createJSResponse(true, "Tile info retrieved", info)
}

// getMovementOptions returns valid movement positions for a unit
func getMovementOptions(q, r int) (any, error) {
	if globalGame == nil {
		return nil, fmt.Errorf("No game loaded")
	}

	// Use game engine method that handles all validation
	return globalGame.GetUnitMovementOptionsFrom(q, r)
}

// getAttackOptions returns valid attack targets for a unit
func getAttackOptions(q, r int) (any, error) {
	if globalGame == nil {
		return nil, fmt.Errorf("No game loaded")
	}

	// Use game engine method that handles all validation
	return globalGame.GetUnitAttackOptionsFrom(q, r)
}

// =============================================================================
// Helper Functions
// =============================================================================

// createJSResponse creates a JavaScript-compatible response
func createJSResponse(success bool, message string, data any) any {
	response := map[string]any{
		"success": success,
		"message": message,
		"data":    data,
	}

	// Convert to JS Value
	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Failed to marshal JSON response: %v\n", err)
		// Return a simple error response
		errorResponse := map[string]any{
			"success": false,
			"message": fmt.Sprintf("JSON marshal error: %v", err),
			"data":    nil,
		}
		errorBytes, _ := json.Marshal(errorResponse)
		return js.Global().Get("JSON").Call("parse", string(errorBytes))
	}

	return js.Global().Get("JSON").Call("parse", string(responseBytes))
}
