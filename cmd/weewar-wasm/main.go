//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// Global game instance
var globalGame *weewar.Game
var globalRulesEngine *weewar.RulesEngine

func main() {
	fmt.Println("WeeWar WASM module loading...")
	
	// Initialize rules engine from embedded data
	var err error
	globalRulesEngine, err = weewar.LoadRulesEngineFromFile(weewar.DevDataPath("data/rules-data.json"))
	if err != nil {
		fmt.Printf("Failed to load rules engine: %v", err)
		return
	}
	fmt.Println("Rules engine loaded successfully")
	
	// Register JavaScript functions
	js.Global().Set("weewarCreateGameFromMap", js.FuncOf(createGameFromMap))
	js.Global().Set("weewarGetGameState", js.FuncOf(getGameState))
	js.Global().Set("weewarSelectUnit", js.FuncOf(selectUnit))
	js.Global().Set("weewarMoveUnit", js.FuncOf(moveUnit))
	js.Global().Set("weewarAttackUnit", js.FuncOf(attackUnit))
	js.Global().Set("weewarEndTurn", js.FuncOf(endTurn))
	
	fmt.Println("WeeWar WASM module loaded successfully")
	
	// Keep the program running
	c := make(chan struct{})
	<-c
}

// =============================================================================
// WASM API Functions - Thin Delegation Layer
// =============================================================================

// createGameFromMap creates a new game from web map data
func createGameFromMap(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return createJSResponse(false, "Missing mapData or playerCount arguments", nil)
	}

	// mapDataStr := args[0].String() // TODO: Use this when implementing real map parsing
	playerCount := args[1].Int()

	if playerCount < 2 || playerCount > 6 {
		return createJSResponse(false, fmt.Sprintf("Invalid player count: %d", playerCount), nil)
	}

	// For now, create a simple test world
	// TODO: Parse mapDataStr and create World from it
	world, err := createTestWorld(playerCount)
	if err != nil {
		return createJSResponse(false, fmt.Sprintf("Failed to create test world: %v", err), nil)
	}

	// Create game using existing NewGame method
	seed := time.Now().UnixNano()
	game, err := weewar.NewGame(world, globalRulesEngine, seed)
	if err != nil {
		return createJSResponse(false, fmt.Sprintf("Failed to create game: %v", err), nil)
	}

	// Store game globally
	globalGame = game

	// Return game state using new UI method
	gameState := game.GetGameStateForUI()
	return createJSResponse(true, "Game created successfully", gameState)
}

// getGameState returns current game state for UI
func getGameState(this js.Value, args []js.Value) interface{} {
	if globalGame == nil {
		return createJSResponse(false, "No game loaded", nil)
	}

	gameState := globalGame.GetGameStateForUI()
	return createJSResponse(true, "Game state retrieved", gameState)
}

// selectUnit selects unit and returns movement/attack options
func selectUnit(this js.Value, args []js.Value) interface{} {
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
	data := map[string]interface{}{
		"unit":            unit,       // *Unit (already JSON-tagged)
		"movableCoords":   movable,    // []TileOption from RulesEngine
		"attackableCoords": attackable, // []AxialCoord from RulesEngine
	}

	return createJSResponse(true, "Unit selected", data)
}

// moveUnit moves a unit from one position to another
func moveUnit(this js.Value, args []js.Value) interface{} {
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
	unit := globalGame.GetUnitAt(to)
	return createJSResponse(true, "Unit moved successfully", map[string]interface{}{
		"unit": unit,
	})
}

// attackUnit performs combat between two units
func attackUnit(this js.Value, args []js.Value) interface{} {
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
func endTurn(this js.Value, args []js.Value) interface{} {
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

// =============================================================================
// Helper Functions
// =============================================================================

// createJSResponse creates a JavaScript-compatible response
func createJSResponse(success bool, message string, data interface{}) interface{} {
	response := map[string]interface{}{
		"success": success,
		"message": message,
		"data":    data,
	}

	// Convert to JS Value
	responseBytes, _ := json.Marshal(response)
	return js.Global().Get("JSON").Call("parse", string(responseBytes))
}

// createTestWorld creates a simple test world for now
// TODO: Replace with proper map data parsing
func createTestWorld(playerCount int) (*weewar.World, error) {
	// Create a simple 6x6 hex map for testing
	gameMap := weewar.NewMapRect(6, 6)

	// Add some test tiles
	for q := -2; q <= 2; q++ {
		for r := -2; r <= 2; r++ {
			if q+r >= -2 && q+r <= 2 { // Valid hex coordinates
				coord := weewar.AxialCoord{Q: q, R: r}
				tile := weewar.NewTile(coord, 1) // Grass terrain
				gameMap.AddTile(tile)
			}
		}
	}

	// Create world with playerCount and map
	world, err := weewar.NewWorld(playerCount, gameMap)
	if err != nil {
		return nil, err
	}

	// Add some test units
	unit1 := weewar.NewUnit(1, 0) // Infantry for player 0
	unit1.SetPosition(weewar.AxialCoord{Q: 0, R: 0})
	world.AddUnit(unit1)

	unit2 := weewar.NewUnit(1, 1) // Infantry for player 1 
	unit2.SetPosition(weewar.AxialCoord{Q: 1, R: -1})
	world.AddUnit(unit2)

	return world, nil
}
