//go:build js && wasm
// +build js,wasm

package main

import "fmt"

func main() {
	fmt.Println("This is commented out")
	c := make(chan struct{})
	<-c
}

/*
// Global CLI instance for WASM
var globalCLI *WeeWarCLI

// WebCLI wraps the CLI for web use
type WebCLI struct {
	cli  *weewar.WeeWarCLI
	game *weewar.Game
}

// JSResponse represents a JavaScript-friendly response
type JSResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func main() {
	// Keep the program running
	c := make(chan struct{})

	// Register JavaScript functions
	js.Global().Set("weewarCreateGame", js.FuncOf(createGame))
	js.Global().Set("weewarLoadGame", js.FuncOf(loadGame))
	js.Global().Set("weewarExecuteCommand", js.FuncOf(executeCommand))
	js.Global().Set("weewarGetGameState", js.FuncOf(getGameState))
	js.Global().Set("weewarRenderGame", js.FuncOf(renderGame))
	js.Global().Set("weewarSaveGame", js.FuncOf(saveGame))
	js.Global().Set("weewarSetVerbose", js.FuncOf(setVerbose))
	js.Global().Set("weewarSetDisplayMode", js.FuncOf(setDisplayMode))

	fmt.Println("WeeWar WASM CLI loaded")
	<-c
}

// createGame creates a new game with specified number of players
func createGame(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return createJSResponse(false, "", "Missing player count argument", nil)
	}

	playerCount := args[0].Int()
	if playerCount < 2 || playerCount > 6 {
		return createJSResponse(false, "", fmt.Sprintf("Invalid player count: %d (must be 2-6)", playerCount), nil)
	}

	// Create test map
	testMap := createTestMap()

	// Create game
	seed := time.Now().UnixNano()
	game, err := weewar.NewGame(playerCount, testMap, seed)
	if err != nil {
		return createJSResponse(false, "", fmt.Sprintf("Failed to create game: %v", err), nil)
	}

	// Create CLI
	globalCLI = weewar.NewWeeWarCLI(game)

	return createJSResponse(true, fmt.Sprintf("Game created with %d players", playerCount), "", map[string]any{
		"playerCount":   playerCount,
		"turnNumber":    game.GetTurnNumber(),
		"currentPlayer": game.GetCurrentPlayer(),
	})
}

// loadGame loads a game from JSON data
func loadGame(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return createJSResponse(false, "", "Missing game data argument", nil)
	}

	gameDataStr := args[0].String()
	gameData := []byte(gameDataStr)

	game, err := weewar.LoadGame(gameData)
	if err != nil {
		return createJSResponse(false, "", fmt.Sprintf("Failed to load game: %v", err), nil)
	}

	// Create CLI
	globalCLI = weewar.NewWeeWarCLI(game)

	return createJSResponse(true, "Game loaded successfully", "", map[string]any{
		"playerCount":   game.PlayerCount,
		"turnNumber":    game.GetTurnNumber(),
		"currentPlayer": game.GetCurrentPlayer(),
	})
}

// executeCommand executes a game command
func executeCommand(this js.Value, args []js.Value) any {
	if globalCLI == nil {
		return createJSResponse(false, "", "No game loaded", nil)
	}

	if len(args) < 1 {
		return createJSResponse(false, "", "Missing command argument", nil)
	}

	command := args[0].String()
	response := globalCLI.ExecuteCommand(command)

	// Convert CLI response to JS response
	return createJSResponse(response.Success, response.Message, response.Error, response.Data)
}

// getGameState returns the current game state
func getGameState(this js.Value, args []js.Value) any {
	if globalCLI == nil {
		return createJSResponse(false, "", "No game loaded", nil)
	}

	game := globalCLI.GetGame()
	if game == nil {
		return createJSResponse(false, "", "No game available", nil)
	}

	// Get basic game state
	state := map[string]any{
		"playerCount":   game.PlayerCount,
		"turnNumber":    game.GetTurnNumber(),
		"currentPlayer": game.GetCurrentPlayer(),
		"status":        string(game.GetGameStatus()),
		"mapSize": map[string]int{
			"rows": game.Map.NumRows(),
			"cols": game.Map.NumCols(),
		},
	}

	// Add player unit counts
	playerInfo := make(map[string]any)
	for i := 0; i < game.PlayerCount; i++ {
		units := game.GetUnitsForPlayer(i)
		playerInfo[fmt.Sprintf("player%d", i)] = map[string]any{
			"unitCount": len(units),
		}
	}
	state["players"] = playerInfo

	return createJSResponse(true, "Game state retrieved", "", state)
}

// renderGame renders the game to a data URL (base64 PNG)
func renderGame(this js.Value, args []js.Value) any {
	if globalCLI == nil {
		return createJSResponse(false, "", "No game loaded", nil)
	}

	// Default dimensions
	width, height := 800, 600

	// Parse optional dimensions
	if len(args) >= 2 {
		width = args[0].Int()
		height = args[1].Int()
	}

	// Render to buffer
	buffer := weewar.NewBuffer(width, height)
	game := globalCLI.GetGame()

	// Calculate tile sizes
	tileWidth := float64(width) / float64(game.Map.NumCols())
	tileHeight := float64(height) / float64(game.Map.NumRows())
	yIncrement := tileHeight * 0.75

	err := game.RenderToBuffer(buffer, tileWidth, tileHeight, yIncrement)
	if err != nil {
		return createJSResponse(false, "", fmt.Sprintf("Failed to render game: %v", err), nil)
	}

	// Convert buffer to base64 data URL
	dataURL, err := buffer.ToDataURL()
	if err != nil {
		return createJSResponse(false, "", fmt.Sprintf("Failed to create data URL: %v", err), nil)
	}

	return createJSResponse(true, "Game rendered successfully", "", map[string]any{
		"dataURL": dataURL,
		"width":   width,
		"height":  height,
	})
}

// saveGame saves the current game to JSON
func saveGame(this js.Value, args []js.Value) any {
	if globalCLI == nil {
		return createJSResponse(false, "", "No game loaded", nil)
	}

	saveData, err := globalCLI.GetGame().SaveGame()
	if err != nil {
		return createJSResponse(false, "", fmt.Sprintf("Failed to save game: %v", err), nil)
	}

	return createJSResponse(true, "Game saved successfully", "", map[string]any{
		"saveData": string(saveData),
		"size":     len(saveData),
	})
}

// setVerbose sets verbose mode
func setVerbose(this js.Value, args []js.Value) any {
	if globalCLI == nil {
		return createJSResponse(false, "", "No game loaded", nil)
	}

	verbose := true
	if len(args) >= 1 {
		verbose = args[0].Bool()
	}

	globalCLI.SetVerbose(verbose)

	status := "enabled"
	if !verbose {
		status = "disabled"
	}

	return createJSResponse(true, fmt.Sprintf("Verbose mode %s", status), "", nil)
}

// setDisplayMode sets the display mode
func setDisplayMode(this js.Value, args []js.Value) any {
	if globalCLI == nil {
		return createJSResponse(false, "", "No game loaded", nil)
	}

	if len(args) < 1 {
		return createJSResponse(false, "", "Missing display mode argument", nil)
	}

	modeStr := args[0].String()
	var mode weewar.CLIDisplayMode

	switch modeStr {
	case "compact":
		mode = weewar.DisplayCompact
	case "detailed":
		mode = weewar.DisplayDetailed
	default:
		return createJSResponse(false, "", fmt.Sprintf("Invalid display mode: %s", modeStr), nil)
	}

	globalCLI.SetDisplayMode(mode)

	return createJSResponse(true, fmt.Sprintf("Display mode set to %s", modeStr), "", nil)
}

// createTestMap creates a test map with varied terrain
func createTestMap() *weewar.Map {
	testMap := weewar.NewMap(8, 12, false)

	// Add varied terrain
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tileType := 1 // Default to grass
			if (row+col)%4 == 0 {
				tileType = 2 // Some desert
			} else if (row+col)%7 == 0 {
				tileType = 3 // Some water
			} else if (row+col)%11 == 0 {
				tileType = 4 // Some mountains
			}

			tile := weewar.NewTile(row, col, tileType)
			testMap.AddTile(tile)
		}
	}

	return testMap
}

// createJSResponse creates a JavaScript-compatible response object
func createJSResponse(success bool, message, error string, data any) js.Value {
	response := JSResponse{
		Success: success,
		Message: message,
		Error:   error,
		Data:    data,
	}

	// Convert to JS object
	responseBytes, _ := json.Marshal(response)
	return js.Global().Get("JSON").Call("parse", string(responseBytes))
}
*/
