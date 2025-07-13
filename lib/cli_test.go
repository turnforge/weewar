package weewar

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// CLI Tests
// =============================================================================

func TestCLIBasicOperations(t *testing.T) {
	// Create test map
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tileType := 1
			if (row+col)%4 == 0 {
				tileType = 2
			}
			tile := NewTile(row, col, tileType)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	// Create game
	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create CLI
	cli := NewWeeWarCLI(game)

	// Test command parsing
	cmd := cli.ParseCommand("move A1 B2")
	if cmd == nil {
		t.Error("Failed to parse move command")
	}
	if cmd.Command != "move" {
		t.Errorf("Expected command 'move', got '%s'", cmd.Command)
	}
	if len(cmd.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(cmd.Arguments))
	}
	if cmd.Arguments[0] != "A1" || cmd.Arguments[1] != "B2" {
		t.Errorf("Expected arguments [A1, B2], got %v", cmd.Arguments)
	}

	// Test command validation
	valid, msg := cli.ValidateCommand(cmd)
	if !valid {
		t.Errorf("Move command should be valid: %s", msg)
	}

	// Test available commands
	commands := cli.GetAvailableCommands()
	if len(commands) == 0 {
		t.Error("No available commands returned")
	}

	// Test help
	help := cli.GetCommandHelp("move")
	if help == "" {
		t.Error("No help returned for move command")
	}
}

func TestCLICommands(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Test status command
	response := cli.ExecuteCommand("status")
	if !response.Success {
		t.Errorf("Status command failed: %s", response.Error)
	}

	// Test map command
	response = cli.ExecuteCommand("map")
	if !response.Success {
		t.Errorf("Map command failed: %s", response.Error)
	}

	// Test units command
	response = cli.ExecuteCommand("units")
	if !response.Success {
		t.Errorf("Units command failed: %s", response.Error)
	}

	// Test player command
	response = cli.ExecuteCommand("player")
	if !response.Success {
		t.Errorf("Player command failed: %s", response.Error)
	}

	// Test help command
	response = cli.ExecuteCommand("help")
	if !response.Success {
		t.Errorf("Help command failed: %s", response.Error)
	}

	// Test invalid command
	response = cli.ExecuteCommand("invalid")
	if response.Success {
		t.Error("Invalid command should fail")
	}
}

func TestCLIGameOperations(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Test end turn
	response := cli.ExecuteCommand("end")
	if !response.Success {
		t.Errorf("End turn command failed: %s", response.Error)
	}

	// Verify turn changed
	if game.GetCurrentPlayer() != 1 {
		t.Errorf("Expected current player 1, got %d", game.GetCurrentPlayer())
	}

	// Test new game command
	response = cli.ExecuteCommand("new 3")
	if !response.Success {
		t.Errorf("New game command failed: %s", response.Error)
	}

	// Verify new game created
	if cli.game.PlayerCount != 3 {
		t.Errorf("Expected 3 players, got %d", cli.game.PlayerCount)
	}
}

func TestCLISaveLoad(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Create temporary file
	tempDir := "/tmp/turnengine/cli_test"
	os.MkdirAll(tempDir, 0755)
	timestamp := time.Now().Format("20060102_150405")
	saveFile := filepath.Join(tempDir, "test_save_"+timestamp+".json")

	// Test save
	response := cli.ExecuteCommand("save " + saveFile)
	if !response.Success {
		t.Errorf("Save command failed: %s", response.Error)
	}

	// Verify file exists
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}

	// Test load
	response = cli.ExecuteCommand("load " + saveFile)
	if !response.Success {
		t.Errorf("Load command failed: %s", response.Error)
	}

	// Verify game loaded
	if cli.game == nil {
		t.Error("Game was not loaded")
	}

	// Cleanup
	os.Remove(saveFile)
}

func TestCLIRender(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Create temporary file
	tempDir := "/tmp/turnengine/cli_test"
	os.MkdirAll(tempDir, 0755)
	timestamp := time.Now().Format("20060102_150405")
	renderFile := filepath.Join(tempDir, "test_render_"+timestamp+".png")

	// Test render
	response := cli.ExecuteCommand("render " + renderFile)
	if !response.Success {
		t.Errorf("Render command failed: %s", response.Error)
	}

	// Verify file exists
	if _, err := os.Stat(renderFile); os.IsNotExist(err) {
		t.Error("Render file was not created")
	} else {
		t.Logf("Render test saved to: %s", renderFile)
	}

	// Test render with size
	renderFile2 := filepath.Join(tempDir, "test_render_sized_"+timestamp+".png")
	response = cli.ExecuteCommand("render " + renderFile2 + " 400 300")
	if !response.Success {
		t.Errorf("Render with size command failed: %s", response.Error)
	}

	// Cleanup
	os.Remove(renderFile)
	os.Remove(renderFile2)
}

func TestCLIBatchProcessing(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Create temporary batch file
	tempDir := "/tmp/turnengine/cli_test"
	os.MkdirAll(tempDir, 0755)
	timestamp := time.Now().Format("20060102_150405")
	batchFile := filepath.Join(tempDir, "test_batch_"+timestamp+".txt")

	batchContent := `# Test batch commands
status
map
units
end
status
`

	err = os.WriteFile(batchFile, []byte(batchContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create batch file: %v", err)
	}

	// Test batch execution
	err = cli.ExecuteBatchCommands(batchFile)
	if err != nil {
		t.Errorf("Batch execution failed: %v", err)
	}

	// Cleanup
	os.Remove(batchFile)
}

func TestCLIPositionParsing(t *testing.T) {
	// Test position parsing
	testCases := []struct {
		input     string
		expectRow int
		expectCol int
		expectValid bool
	}{
		{"A1", 0, 0, true},
		{"B2", 1, 1, true},
		{"Z26", 25, 25, true},
		{"a1", 0, 0, true}, // Should handle lowercase
		{"", 0, 0, false},
		{"1A", 0, 0, false},
		{"AA", 0, 0, false},
		{"A0", 0, 0, false},
		{"A100", 0, 0, false},
	}

	for _, tc := range testCases {
		row, col, valid := ParsePositionFromString(tc.input)
		if valid != tc.expectValid {
			t.Errorf("Position '%s': expected valid=%v, got valid=%v", tc.input, tc.expectValid, valid)
		}
		if valid && (row != tc.expectRow || col != tc.expectCol) {
			t.Errorf("Position '%s': expected (%d,%d), got (%d,%d)", tc.input, tc.expectRow, tc.expectCol, row, col)
		}
	}

	// Test position formatting
	testCases2 := []struct {
		row, col int
		expect   string
	}{
		{0, 0, "A1"},
		{1, 1, "B2"},
		{25, 25, "Z26"},
		{-1, -1, "??"},
		{0, 26, "??"},
	}

	for _, tc := range testCases2 {
		result := FormatPositionToString(tc.row, tc.col)
		if result != tc.expect {
			t.Errorf("Format position (%d,%d): expected '%s', got '%s'", tc.row, tc.col, tc.expect, result)
		}
	}
}

func TestCLIDisplayModes(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Test display mode changes
	cli.SetDisplayMode(DisplayCompact)
	if cli.GetDisplayMode() != DisplayCompact {
		t.Error("Display mode was not set to compact")
	}

	cli.SetDisplayMode(DisplayDetailed)
	if cli.GetDisplayMode() != DisplayDetailed {
		t.Error("Display mode was not set to detailed")
	}

	// Test verbose mode
	cli.SetVerbose(true)
	if !cli.IsVerbose() {
		t.Error("Verbose mode was not enabled")
	}

	cli.SetVerbose(false)
	if cli.IsVerbose() {
		t.Error("Verbose mode was not disabled")
	}

	// Test compact command
	response := cli.ExecuteCommand("compact")
	if !response.Success {
		t.Errorf("Compact command failed: %s", response.Error)
	}

	// Test verbose command
	response = cli.ExecuteCommand("verbose")
	if !response.Success {
		t.Errorf("Verbose command failed: %s", response.Error)
	}
}

func TestCLIREPLCommands(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)
	
	// Enable interactive mode to test REPL commands
	cli.interactive = true

	// Test REPL-specific commands
	testCases := []struct {
		command string
		shouldSucceed bool
	}{
		{"state", true},
		{"s", true},
		{"refresh", true},
		{"r", true},
		{"turn", true},
		{"actions", true},
	}

	for _, tc := range testCases {
		response := cli.executeREPLCommand(tc.command)
		if response.Success != tc.shouldSucceed {
			t.Errorf("REPL command '%s': expected success=%v, got success=%v", 
				tc.command, tc.shouldSucceed, response.Success)
		}
	}

	// Test REPL command availability
	commands := cli.GetAvailableCommands()
	replCommands := []string{"state", "refresh", "turn", "actions"}
	
	for _, replCmd := range replCommands {
		found := false
		for _, cmd := range commands {
			if cmd == replCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("REPL command '%s' not found in available commands", replCmd)
		}
	}

	// Test REPL help
	for _, replCmd := range replCommands {
		help := cli.GetCommandHelp(replCmd)
		if help == "Unknown command. Use 'help' to see all available commands." {
			t.Errorf("No help found for REPL command '%s'", replCmd)
		}
	}
}

func TestCLIGameStateIntegration(t *testing.T) {
	// Create test game
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tile := NewTile(row, col, 1)
			testMap.AddTile(tile)
		}
	}
	// Note: Neighbor connections calculated on-demand

	game, err := NewGame(2, testMap, 12345)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	cli := NewWeeWarCLI(game)

	// Test GameInterface method integration
	
	// Test GetCurrentPlayer
	currentPlayer := cli.game.GetCurrentPlayer()
	if currentPlayer != 0 {
		t.Errorf("Expected current player 0, got %d", currentPlayer)
	}

	// Test GetTurnNumber
	turnNumber := cli.game.GetTurnNumber()
	if turnNumber != 1 {
		t.Errorf("Expected turn number 1, got %d", turnNumber)
	}

	// Test GetGameStatus
	gameStatus := cli.game.GetGameStatus()
	if gameStatus != GameStatusPlaying {
		t.Errorf("Expected game status playing, got %s", gameStatus)
	}

	// Test GetUnitsForPlayer
	units := cli.game.GetUnitsForPlayer(0)
	if len(units) != 2 {
		t.Errorf("Expected 2 units for player 0, got %d", len(units))
	}

	// Test CanEndTurn
	canEndTurn := cli.game.CanEndTurn()
	if !canEndTurn {
		t.Error("Expected to be able to end turn")
	}

	// Test MoveUnit (via CLI) - Use unit ID A1 instead of chess notation B2 to avoid conflict
	response := cli.ExecuteCommand("move A1 B3")
	if !response.Success {
		t.Errorf("Move command failed: %s", response.Error)
	}

	// Verify unit moved - A1 unit should now be at B3 (1,2) 
	unit := cli.game.GetUnitAt(2, 1) // B3 = row 2, col 1 (0-based)
	if unit == nil {
		t.Error("Unit not found at expected position after move")
	}

	// Test EndTurn (via CLI)
	response = cli.ExecuteCommand("end")
	if !response.Success {
		t.Errorf("End turn command failed: %s", response.Error)
	}

	// Verify turn changed
	newCurrentPlayer := cli.game.GetCurrentPlayer()
	if newCurrentPlayer != 1 {
		t.Errorf("Expected current player 1 after end turn, got %d", newCurrentPlayer)
	}
}