package weewar

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// CLI Implementation
// =============================================================================

// WeeWarCLI implements the CLIInterface for WeeWar games
type WeeWarCLI struct {
	game         *Game
	displayMode  CLIDisplayMode
	verbose      bool
	formatter    CLIFormatter
	recording    bool
	recordFile   *os.File
	interactive  bool
	autoRender   bool
	renderDir    string
	maxRenders   int
	commandCount int
}

// NewWeeWarCLI creates a new CLI instance
func NewWeeWarCLI(game *Game) *WeeWarCLI {
	return &WeeWarCLI{
		game:        game,
		displayMode: DisplayDetailed,
		verbose:     false,
		formatter:   NewDefaultFormatter(),
		recording:   false,
		interactive: false,
		autoRender:  false,
		renderDir:   "/tmp/turnengine/autorenders",
		maxRenders:  0, // Will be set by command line flags
		commandCount: 0,
	}
}

// =============================================================================
// Command Processing
// =============================================================================

// ExecuteCommand processes text commands and returns response
func (cli *WeeWarCLI) ExecuteCommand(command string) *CLIResponse {
	if cli.recording && cli.recordFile != nil {
		fmt.Fprintf(cli.recordFile, "%s\n", command)
	}

	cmd := cli.ParseCommand(command)
	if cmd == nil {
		return &CLIResponse{
			Success: false,
			Message: "Invalid command format",
			Error:   "Failed to parse command",
		}
	}

	// Validate command
	if valid, msg := cli.ValidateCommand(cmd); !valid {
		return &CLIResponse{
			Success: false,
			Message: msg,
			Error:   "Command validation failed",
		}
	}

	// Execute command
	return cli.executeCommand(cmd)
}

// ParseCommand parses command string into structured format
func (cli *WeeWarCLI) ParseCommand(commandStr string) *CLICommand {
	commandStr = strings.TrimSpace(commandStr)
	if commandStr == "" {
		return nil
	}

	parts := strings.Fields(commandStr)
	if len(parts) == 0 {
		return nil
	}

	cmd := &CLICommand{
		Command:   strings.ToLower(parts[0]),
		Arguments: []string{},
		Options:   make(map[string]string),
		Raw:       commandStr,
	}

	// Parse arguments and options
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if strings.HasPrefix(part, "--") {
			// Long option
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part[2:], "=", 2)
				cmd.Options[kv[0]] = kv[1]
			} else {
				cmd.Options[part[2:]] = "true"
			}
		} else if strings.HasPrefix(part, "-") {
			// Short option
			cmd.Options[part[1:]] = "true"
		} else {
			// Argument
			cmd.Arguments = append(cmd.Arguments, part)
		}
	}

	return cmd
}

// executeCommand handles the actual command execution
func (cli *WeeWarCLI) executeCommand(cmd *CLICommand) *CLIResponse {
	switch cmd.Command {
	case CmdMove:
		return cli.handleMove(cmd)
	case CmdAttack:
		return cli.handleAttack(cmd)
	case CmdStatus:
		return cli.handleStatus(cmd)
	case CmdMap:
		return cli.handleMap(cmd)
	case CmdUnits:
		return cli.handleUnits(cmd)
	case CmdPlayer:
		return cli.handlePlayer(cmd)
	case CmdHelp:
		return cli.handleHelp(cmd)
	case CmdSave:
		return cli.handleSave(cmd)
	case CmdLoad:
		return cli.handleLoad(cmd)
	case CmdRender:
		return cli.handleRender(cmd)
	case CmdEnd:
		return cli.handleEndTurn(cmd)
	case CmdQuit:
		return cli.handleQuit(cmd)
	case CmdNew:
		return cli.handleNew(cmd)
	case CmdVerbose:
		return cli.handleVerbose(cmd)
	case CmdCompact:
		return cli.handleCompact(cmd)
	case "autorender":
		return cli.handleAutoRender(cmd)
	case "predict":
		return cli.handlePredict(cmd)
	case "attackoptions":
		return cli.handleAttackOptions(cmd)
	case "moveoptions":
		return cli.handleMoveOptions(cmd)
	default:
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown command: %s", cmd.Command),
			Error:   "Use 'help' to see available commands",
		}
	}
}

// GetAvailableCommands returns list of valid commands
func (cli *WeeWarCLI) GetAvailableCommands() []string {
	commands := []string{
		CmdMove, CmdAttack, CmdStatus, CmdMap, CmdUnits, CmdPlayer,
		CmdHelp, CmdSave, CmdLoad, CmdRender, CmdEnd, CmdQuit,
		CmdNew, CmdVerbose, CmdCompact, "autorender", "predict",
		"attackoptions", "moveoptions",
	}
	
	// Add REPL-specific commands if in interactive mode
	if cli.interactive {
		commands = append(commands, "state", "refresh", "turn", "actions")
	}
	
	return commands
}

// GetCommandHelp returns help text for specific command
func (cli *WeeWarCLI) GetCommandHelp(command string) string {
	switch command {
	case CmdMove:
		return "move <from> <to> - Move unit from one position to another (e.g., 'move A1 B2')"
	case CmdAttack:
		return "attack <attacker> <target> - Attack with unit at position (e.g., 'attack A1 B2')"
	case CmdStatus:
		return "status - Show current game status and player information"
	case CmdMap:
		return "map - Display the current game map"
	case CmdUnits:
		return "units [player] - Show units for current player or specified player"
	case CmdPlayer:
		return "player [id] - Show information about current player or specified player"
	case CmdHelp:
		return "help [command] - Show help for all commands or specific command"
	case CmdSave:
		return "save <filename> - Save current game state to file"
	case CmdLoad:
		return "load <filename> - Load game state from file"
	case CmdRender:
		return "render <filename> [width] [height] - Render game state to PNG file"
	case CmdEnd:
		return "end - End current player's turn"
	case CmdQuit:
		return "quit - Exit the game"
	case CmdNew:
		return "new [players] - Start a new game with specified number of players (default: 2)"
	case CmdVerbose:
		return "verbose - Toggle verbose output mode"
	case CmdCompact:
		return "compact - Set compact display mode"
	case "autorender":
		return "autorender - Toggle automatic PNG rendering after each command"
	// REPL-specific commands
	case "state", "s":
		return "state/s - Quick game status display (REPL shortcut)"
	case "refresh", "r":
		return "refresh/r - Refresh game state display (REPL shortcut)"
	case "turn":
		return "turn - Show detailed turn information including available actions"
	case "actions":
		return "actions - Show all available actions for current player"
	case "predict":
		return "predict <attacker> <target> - Show damage prediction for combat (e.g., 'predict A1 B2')"
	case "attackoptions":
		return "attackoptions <unit> - Show all possible attack targets for a unit (e.g., 'attackoptions A1')"
	case "moveoptions":
		return "moveoptions <unit> - Show all possible movement positions for a unit (e.g., 'moveoptions A1')"
	default:
		return "Unknown command. Use 'help' to see all available commands."
	}
}

// ValidateCommand checks if command is valid in current context
func (cli *WeeWarCLI) ValidateCommand(cmd *CLICommand) (bool, string) {
	// Check if game exists for game-specific commands
	gameCommands := []string{CmdMove, CmdAttack, CmdStatus, CmdMap, CmdUnits, CmdPlayer, CmdEnd, CmdSave, CmdRender, "predict", "attackoptions", "moveoptions"}
	for _, gameCmd := range gameCommands {
		if cmd.Command == gameCmd && cli.game == nil {
			return false, "No game is currently loaded. Use 'new' to start a new game or 'load' to load a saved game."
		}
	}

	// Validate specific commands
	switch cmd.Command {
	case CmdMove:
		if len(cmd.Arguments) < 2 {
			return false, "Move command requires from and to positions (e.g., 'move A1 B2')"
		}
	case CmdAttack:
		if len(cmd.Arguments) < 2 {
			return false, "Attack command requires attacker and target positions (e.g., 'attack A1 B2')"
		}
	case CmdSave:
		if len(cmd.Arguments) < 1 {
			return false, "Save command requires filename (e.g., 'save mygame.json')"
		}
	case CmdLoad:
		if len(cmd.Arguments) < 1 {
			return false, "Load command requires filename (e.g., 'load mygame.json')"
		}
	case CmdRender:
		if len(cmd.Arguments) < 1 {
			return false, "Render command requires filename (e.g., 'render game.png')"
		}
	}

	return true, ""
}

// =============================================================================
// Command Handlers
// =============================================================================

// handleMove processes move commands
func (cli *WeeWarCLI) handleMove(cmd *CLICommand) *CLIResponse {
	fromPos := cmd.Arguments[0]
	toPos := cmd.Arguments[1]

	// Parse positions
	fromRow, fromCol, valid := ParsePositionFromString(fromPos)
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid from position: %s", fromPos),
			Error:   "Use format like A1, B2, etc.",
		}
	}

	toRow, toCol, valid := ParsePositionFromString(toPos)
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid to position: %s", toPos),
			Error:   "Use format like A1, B2, etc.",
		}
	}

	// Find unit at from position
	unit := cli.game.GetUnitAt(fromRow, fromCol)
	if unit == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit found at position %s", fromPos),
			Error:   "Cannot move non-existent unit",
		}
	}

	// Check if it's the correct player's turn
	if unit.PlayerID != cli.game.GetCurrentPlayer() {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Unit at %s belongs to player %d, but it's player %d's turn", 
				fromPos, unit.PlayerID, cli.game.GetCurrentPlayer()),
			Error:   "Can only move your own units",
		}
	}

	// Execute move
	if err := cli.game.MoveUnit(unit, toRow, toCol); err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to move unit: %v", err),
			Error:   "Move was not legal",
		}
	}

	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Unit moved from %s to %s", fromPos, toPos),
	}
}

// handleAttack processes attack commands
func (cli *WeeWarCLI) handleAttack(cmd *CLICommand) *CLIResponse {
	attackerPos := cmd.Arguments[0]
	targetPos := cmd.Arguments[1]

	// Parse positions
	attackerRow, attackerCol, valid := ParsePositionFromString(attackerPos)
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid attacker position: %s", attackerPos),
			Error:   "Use format like A1, B2, etc.",
		}
	}

	targetRow, targetCol, valid := ParsePositionFromString(targetPos)
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid target position: %s", targetPos),
			Error:   "Use format like A1, B2, etc.",
		}
	}

	// Find units
	attacker := cli.game.GetUnitAt(attackerRow, attackerCol)
	if attacker == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit found at attacker position %s", attackerPos),
			Error:   "Cannot attack with non-existent unit",
		}
	}

	target := cli.game.GetUnitAt(targetRow, targetCol)
	if target == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit found at target position %s", targetPos),
			Error:   "Cannot attack non-existent unit",
		}
	}

	// Check if it's the correct player's turn
	if attacker.PlayerID != cli.game.GetCurrentPlayer() {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Unit at %s belongs to player %d, but it's player %d's turn", 
				attackerPos, attacker.PlayerID, cli.game.GetCurrentPlayer()),
			Error:   "Can only attack with your own units",
		}
	}

	// Execute attack
	result, err := cli.game.AttackUnit(attacker, target)
	if err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to attack: %v", err),
			Error:   "Attack was not legal",
		}
	}

	message := fmt.Sprintf("Attack from %s to %s: %d damage dealt", 
		attackerPos, targetPos, result.DefenderDamage)
	if result.DefenderKilled {
		message += " (target destroyed)"
	}

	return &CLIResponse{
		Success: true,
		Message: message,
	}
}

// handleStatus shows current game status
func (cli *WeeWarCLI) handleStatus(cmd *CLICommand) *CLIResponse {
	cli.PrintGameState()
	return &CLIResponse{
		Success: true,
		Message: "Game status displayed",
	}
}

// handleMap shows the game map
func (cli *WeeWarCLI) handleMap(cmd *CLICommand) *CLIResponse {
	cli.PrintMap()
	return &CLIResponse{
		Success: true,
		Message: "Map displayed",
	}
}

// handleUnits shows unit information
func (cli *WeeWarCLI) handleUnits(cmd *CLICommand) *CLIResponse {
	cli.PrintUnits()
	return &CLIResponse{
		Success: true,
		Message: "Units displayed",
	}
}

// handlePlayer shows player information
func (cli *WeeWarCLI) handlePlayer(cmd *CLICommand) *CLIResponse {
	playerID := cli.game.GetCurrentPlayer()
	if len(cmd.Arguments) > 0 {
		if id, err := strconv.Atoi(cmd.Arguments[0]); err == nil {
			playerID = id
		}
	}
	
	cli.PrintPlayerInfo(playerID)
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Player %d information displayed", playerID),
	}
}

// handleHelp shows help information
func (cli *WeeWarCLI) handleHelp(cmd *CLICommand) *CLIResponse {
	if len(cmd.Arguments) > 0 {
		cli.PrintHelp(cmd.Arguments[0])
	} else {
		cli.PrintHelp("")
	}
	return &CLIResponse{
		Success: true,
		Message: "Help displayed",
	}
}

// handleSave saves game to file
func (cli *WeeWarCLI) handleSave(cmd *CLICommand) *CLIResponse {
	filename := cmd.Arguments[0]
	if err := cli.SaveGameToFile(filename); err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to save game: %v", err),
			Error:   "Save operation failed",
		}
	}
	
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Game saved to %s", filename),
	}
}

// handleLoad loads game from file
func (cli *WeeWarCLI) handleLoad(cmd *CLICommand) *CLIResponse {
	filename := cmd.Arguments[0]
	if err := cli.LoadGameFromFile(filename); err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to load game: %v", err),
			Error:   "Load operation failed",
		}
	}
	
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Game loaded from %s", filename),
	}
}

// handleRender renders game to PNG file
func (cli *WeeWarCLI) handleRender(cmd *CLICommand) *CLIResponse {
	filename := cmd.Arguments[0]
	width := 800
	height := 600
	
	if len(cmd.Arguments) > 1 {
		if w, err := strconv.Atoi(cmd.Arguments[1]); err == nil {
			width = w
		}
	}
	
	if len(cmd.Arguments) > 2 {
		if h, err := strconv.Atoi(cmd.Arguments[2]); err == nil {
			height = h
		}
	}
	
	if err := cli.RenderToFile(filename, width, height); err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to render game: %v", err),
			Error:   "Render operation failed",
		}
	}
	
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Game rendered to %s (%dx%d)", filename, width, height),
	}
}

// handleEndTurn ends current player's turn
func (cli *WeeWarCLI) handleEndTurn(cmd *CLICommand) *CLIResponse {
	if err := cli.game.EndTurn(); err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to end turn: %v", err),
			Error:   "End turn operation failed",
		}
	}
	
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Turn ended. Now player %d's turn (Turn %d)", 
			cli.game.GetCurrentPlayer(), cli.game.GetTurnNumber()),
	}
}

// handleQuit exits the game
func (cli *WeeWarCLI) handleQuit(cmd *CLICommand) *CLIResponse {
	return &CLIResponse{
		Success: true,
		Message: "Goodbye!",
		Data:    "quit",
	}
}

// handleNew starts a new game
func (cli *WeeWarCLI) handleNew(cmd *CLICommand) *CLIResponse {
	playerCount := 2
	if len(cmd.Arguments) > 0 {
		if pc, err := strconv.Atoi(cmd.Arguments[0]); err == nil && pc >= 2 && pc <= 6 {
			playerCount = pc
		}
	}
	
	// Create test map
	testMap := NewMap(8, 12, false)
	for row := 0; row < 8; row++ {
		for col := 0; col < 12; col++ {
			tileType := 1 // Default to grass
			if (row+col)%4 == 0 {
				tileType = 2 // Some desert
			}
			tile := NewTile(row, col, tileType)
			testMap.AddTile(tile)
		}
	}
	testMap.ConnectHexNeighbors()
	
	// Create new game
	newGame, err := NewGame(playerCount, testMap, time.Now().UnixNano())
	if err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create new game: %v", err),
			Error:   "Game creation failed",
		}
	}
	
	cli.game = newGame
	
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("New game created with %d players", playerCount),
	}
}

// handleVerbose toggles verbose mode
func (cli *WeeWarCLI) handleVerbose(cmd *CLICommand) *CLIResponse {
	cli.verbose = !cli.verbose
	status := "disabled"
	if cli.verbose {
		status = "enabled"
	}
	
	return &CLIResponse{
		Success: true,
		Message: fmt.Sprintf("Verbose mode %s", status),
	}
}

// handleCompact sets compact display mode
func (cli *WeeWarCLI) handleCompact(cmd *CLICommand) *CLIResponse {
	cli.displayMode = DisplayCompact
	return &CLIResponse{
		Success: true,
		Message: "Display mode set to compact",
	}
}

// handleAutoRender toggles auto-rendering mode
func (cli *WeeWarCLI) handleAutoRender(cmd *CLICommand) *CLIResponse {
	cli.autoRender = !cli.autoRender
	status := "disabled"
	if cli.autoRender && cli.maxRenders > 0 {
		status = "enabled"
		// Create render directory
		if err := os.MkdirAll(cli.renderDir, 0755); err != nil {
			return &CLIResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create render directory: %v", err),
				Error:   "Auto-render setup failed",
			}
		}
	} else if cli.autoRender && cli.maxRenders == 0 {
		cli.autoRender = false
		status = "disabled (maxRenders is 0)"
	}
	
	message := fmt.Sprintf("Auto-rendering %s", status)
	if cli.autoRender && cli.maxRenders > 0 {
		message += fmt.Sprintf(" (max %d files in %s)", cli.maxRenders, cli.renderDir)
	}
	
	return &CLIResponse{
		Success: true,
		Message: message,
	}
}

// handlePredict shows damage prediction for combat
func (cli *WeeWarCLI) handlePredict(cmd *CLICommand) *CLIResponse {
	if len(cmd.Arguments) < 2 {
		return &CLIResponse{
			Success: false,
			Message: "Predict command requires two positions (e.g., 'predict A1 B2')",
			Error:   "Missing arguments",
		}
	}
	
	fromPos := cmd.Arguments[0]
	toPos := cmd.Arguments[1]
	
	// Parse positions
	fromRow, fromCol, valid := ParsePositionFromString(fromPos)
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid from position: %s", fromPos),
			Error:   "Use format like A1, B2, etc.",
		}
	}
	
	toRow, toCol, valid := ParsePositionFromString(toPos)
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid to position: %s", toPos),
			Error:   "Use format like A1, B2, etc.",
		}
	}
	
	// Find attacker unit
	attacker := cli.game.GetUnitAt(fromRow, fromCol)
	if attacker == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit found at position %s", fromPos),
			Error:   "Cannot predict attack from empty position",
		}
	}
	
	// Find target unit
	target := cli.game.GetUnitAt(toRow, toCol)
	if target == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit found at position %s", toPos),
			Error:   "Cannot predict attack on empty position",
		}
	}
	
	// Check if attack is valid
	canAttack, err := cli.game.CanAttack(fromRow, fromCol, toRow, toCol)
	if err != nil || !canAttack {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Unit at %s cannot attack unit at %s", fromPos, toPos),
			Error:   "Invalid attack",
		}
	}
	
	// Create predictor and get damage prediction
	predictor := NewGamePredictor(cli.game.assetManager)
	damagePrediction, err := predictor.GetCombatPredictor().PredictDamage(cli.game, fromRow, fromCol, toRow, toCol)
	if err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to predict damage: %v", err),
			Error:   "Prediction calculation failed",
		}
	}
	
	// Format prediction output
	message := fmt.Sprintf("=== Attack Prediction ===\n")
	message += fmt.Sprintf("Attacker: %s at %s (Health: %d)\n", 
		cli.game.GetUnitTypeName(attacker.UnitType), fromPos, attacker.AvailableHealth)
	message += fmt.Sprintf("Target: %s at %s (Health: %d)\n", 
		cli.game.GetUnitTypeName(target.UnitType), toPos, target.AvailableHealth)
	message += fmt.Sprintf("\nDamage Range: %d - %d\n", 
		damagePrediction.MinDamage, damagePrediction.MaxDamage)
	message += fmt.Sprintf("Expected Damage: %.1f\n", damagePrediction.ExpectedDamage)
	
	// Show damage probabilities
	message += "\nDamage Probabilities:\n"
	for damage := damagePrediction.MinDamage; damage <= damagePrediction.MaxDamage; damage++ {
		if prob, exists := damagePrediction.Probabilities[damage]; exists && prob > 0 {
			message += fmt.Sprintf("  %d damage: %.1f%%\n", damage, prob*100)
		}
	}
	
	// Show outcome predictions
	remainingHealth := target.AvailableHealth - int(damagePrediction.ExpectedDamage)
	if remainingHealth <= 0 {
		message += fmt.Sprintf("\nPredicted Outcome: Target will likely be destroyed")
	} else {
		message += fmt.Sprintf("\nPredicted Target Health: %d", remainingHealth)
	}
	
	fmt.Print(message)
	return &CLIResponse{
		Success: true,
		Message: "Damage prediction displayed",
	}
}

// handleAttackOptions shows all possible attack targets for a unit
func (cli *WeeWarCLI) handleAttackOptions(cmd *CLICommand) *CLIResponse {
	if len(cmd.Arguments) < 1 {
		return &CLIResponse{
			Success: false,
			Message: "Attack options command requires a unit position (e.g., 'attackoptions A1')",
			Error:   "Missing unit position argument",
		}
	}

	// Parse unit position
	fromRow, fromCol, valid := cli.formatter.ParsePosition(cmd.Arguments[0])
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid position format: %s", cmd.Arguments[0]),
			Error:   "Use format like A1, B2, etc.",
		}
	}

	// Check if unit exists at position
	unit := cli.game.GetUnitAt(fromRow, fromCol)
	if unit == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit at position %s", cmd.Arguments[0]),
			Error:   "Cannot show attack options for empty position",
		}
	}

	// Check if it's the current player's unit
	if unit.PlayerID != cli.game.GetCurrentPlayer() {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Unit at %s belongs to player %d, not current player %d", 
				cmd.Arguments[0], unit.PlayerID, cli.game.GetCurrentPlayer()),
			Error:   "Cannot show attack options for opponent's unit",
		}
	}

	// Get attack options using predictor
	predictor := NewGamePredictor(cli.game.assetManager)
	attackOptions, err := predictor.GetCombatPredictor().GetAttackOptions(cli.game, fromRow, fromCol)
	if err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get attack options: %v", err),
			Error:   "Attack options calculation failed",
		}
	}

	// Format output
	unitName := cli.game.GetUnitTypeName(unit.UnitType)
	message := fmt.Sprintf("=== Attack Options for %s at %s ===\n", unitName, cmd.Arguments[0])
	
	if len(attackOptions) == 0 {
		message += "No valid attack targets available.\n"
	} else {
		message += fmt.Sprintf("Available targets (%d):\n", len(attackOptions))
		for i, pos := range attackOptions {
			posStr := cli.formatter.FormatPosition(pos.Row, pos.Col)
			target := cli.game.GetUnitAt(pos.Row, pos.Col)
			if target != nil {
				targetName := cli.game.GetUnitTypeName(target.UnitType)
				message += fmt.Sprintf("  %d. %s - %s (Player %d, Health: %d)\n", 
					i+1, posStr, targetName, target.PlayerID, target.AvailableHealth)
			} else {
				message += fmt.Sprintf("  %d. %s - No unit\n", i+1, posStr)
			}
		}
		message += "\nUse 'predict <unit> <target>' to see damage prediction.\n"
	}

	fmt.Print(message)
	return &CLIResponse{
		Success: true,
		Message: "Attack options displayed",
	}
}

// handleMoveOptions shows all possible movement positions for a unit
func (cli *WeeWarCLI) handleMoveOptions(cmd *CLICommand) *CLIResponse {
	if len(cmd.Arguments) < 1 {
		return &CLIResponse{
			Success: false,
			Message: "Move options command requires a unit position (e.g., 'moveoptions A1')",
			Error:   "Missing unit position argument",
		}
	}

	// Parse unit position
	fromRow, fromCol, valid := cli.formatter.ParsePosition(cmd.Arguments[0])
	if !valid {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid position format: %s", cmd.Arguments[0]),
			Error:   "Use format like A1, B2, etc.",
		}
	}

	// Check if unit exists at position
	unit := cli.game.GetUnitAt(fromRow, fromCol)
	if unit == nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("No unit at position %s", cmd.Arguments[0]),
			Error:   "Cannot show move options for empty position",
		}
	}

	// Check if it's the current player's unit
	if unit.PlayerID != cli.game.GetCurrentPlayer() {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Unit at %s belongs to player %d, not current player %d", 
				cmd.Arguments[0], unit.PlayerID, cli.game.GetCurrentPlayer()),
			Error:   "Cannot show move options for opponent's unit",
		}
	}

	// Get movement options using predictor
	predictor := NewGamePredictor(cli.game.assetManager)
	moveOptions, err := predictor.GetMovementPredictor().GetMovementOptions(cli.game, fromRow, fromCol)
	if err != nil {
		return &CLIResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get movement options: %v", err),
			Error:   "Movement options calculation failed",
		}
	}

	// Format output
	unitName := cli.game.GetUnitTypeName(unit.UnitType)
	message := fmt.Sprintf("=== Movement Options for %s at %s ===\n", unitName, cmd.Arguments[0])
	message += fmt.Sprintf("Movement Points: %d\n", unit.DistanceLeft)
	
	if len(moveOptions) == 0 {
		message += "No valid movement positions available.\n"
	} else {
		message += fmt.Sprintf("Available positions (%d):\n", len(moveOptions))
		for i, pos := range moveOptions {
			posStr := cli.formatter.FormatPosition(pos.Row, pos.Col)
			tile := cli.game.GetTileAt(pos.Row, pos.Col)
			if tile != nil {
				terrainData := GetTerrainData(tile.TileType)
				message += fmt.Sprintf("  %d. %s - %s (Move Cost: %d)\n", 
					i+1, posStr, terrainData.Name, terrainData.MoveCost)
			} else {
				message += fmt.Sprintf("  %d. %s - Unknown terrain\n", i+1, posStr)
			}
		}
		message += "\nUse 'move <unit> <destination>' to move the unit.\n"
	}

	fmt.Print(message)
	return &CLIResponse{
		Success: true,
		Message: "Movement options displayed",
	}
}

// =============================================================================
// Display Functions
// =============================================================================

// PrintGameState outputs current game state to console
func (cli *WeeWarCLI) PrintGameState() {
	if cli.game == nil {
		fmt.Println("No game currently loaded")
		return
	}

	fmt.Printf("=== Game Status ===\n")
	fmt.Printf("Turn: %d\n", cli.game.GetTurnNumber())
	fmt.Printf("Current Player: %d\n", cli.game.GetCurrentPlayer())
	fmt.Printf("Game Status: %s\n", cli.game.GetGameStatus())
	fmt.Printf("Map: %s\n", cli.game.GetMapName())
	
	if winner, hasWinner := cli.game.GetWinner(); hasWinner {
		fmt.Printf("Winner: Player %d\n", winner)
	}
	
	fmt.Printf("Players: %d\n", cli.game.PlayerCount)
	for i := 0; i < cli.game.PlayerCount; i++ {
		units := cli.game.GetUnitsForPlayer(i)
		fmt.Printf("  Player %d: %d units\n", i, len(units))
	}
}

// getTileEmoji returns emoji representation for tile types
func (cli *WeeWarCLI) getTileEmoji(tileType int) string {
	switch tileType {
	case 1:
		return "🌱" // Grass
	case 2:
		return "🏜️" // Desert
	case 3:
		return "🌊" // Water (Regular)
	case 4:
		return "⛰️" // Mountains
	case 5:
		return "🗿" // Rock
	case 6:
		return "🏥" // Hospital
	case 7:
		return "🌾" // Swamp
	case 8:
		return "🌲" // Forest
	case 9:
		return "🌋" // Lava
	case 10:
		return "💧" // Water (Shallow)
	case 11:
		return "🌊" // Water (Deep)
	case 12:
		return "🚀" // Missile Silo
	case 13:
		return "🌉" // Bridge (Regular)
	case 14:
		return "🌉" // Bridge (Shallow)
	case 15:
		return "🌉" // Bridge (Deep)
	case 16:
		return "⛏️" // Mines
	case 17:
		return "🏙️" // City
	case 18:
		return "🛣️" // Road
	case 19:
		return "🗿" // Water (Rocky)
	case 20:
		return "🗼" // Guard Tower
	case 21:
		return "❄️" // Snow
	case 22:
		return "🏰" // Land Base
	case 23:
		return "🏛️" // Naval Base
	case 24:
		return "✈️" // Airport Base
	default:
		return "❓" // Unknown
	}
}

// PrintMap outputs map representation to console with hex grid layout using emojis
func (cli *WeeWarCLI) PrintMap() {
	if cli.game == nil {
		fmt.Println("No game currently loaded")
		return
	}

	fmt.Printf("=== Game Map ===\n")
	rows, cols := cli.game.GetMapSize()
	fmt.Printf("Size: %dx%d\n", rows, cols)
	
	// Print column headers with hex offset consideration
	fmt.Print("       ") // Extra space for hex offset and row numbers
	for col := 0; col < cols; col++ {
		fmt.Printf("  %c   ", 'A'+col)
	}
	fmt.Println()
	
	// Print map rows with hex offset (2 lines per row)
	for row := 0; row < rows; row++ {
		// Apply hex offset based on EvenRowsOffset flag
		isEvenRow := (row % 2) == 0
		needsOffset := (cli.game.Map.EvenRowsOffset && isEvenRow) || (!cli.game.Map.EvenRowsOffset && !isEvenRow)
		
		// First line: terrain emojis
		fmt.Printf("%2d ", row+1)
		if needsOffset {
			fmt.Print("   ") // Offset by 3 spaces for hex layout
		}
		
		for col := 0; col < cols; col++ {
			tile := cli.game.GetTileAt(row, col)
			if tile == nil {
				fmt.Print("      ") // 6 spaces for empty tiles
				continue
			}
			
			// Show terrain emoji centered
			emoji := cli.getTileEmoji(tile.TileType)
			fmt.Printf("  %s  ", emoji)
		}
		fmt.Println()
		
		// Second line: unit information
		fmt.Print("   ") // Space for row number
		if needsOffset {
			fmt.Print("   ") // Offset by 3 spaces for hex layout
		}
		
		for col := 0; col < cols; col++ {
			tile := cli.game.GetTileAt(row, col)
			if tile == nil {
				fmt.Print("      ") // 6 spaces for empty tiles
				continue
			}
			
			// Show unit info centered with unit type
			if tile.Unit != nil {
				fmt.Printf("P%dU%d", tile.Unit.PlayerID, tile.Unit.UnitType)
			} else {
				fmt.Print(" -- ")
			}
		}
		fmt.Println()
		fmt.Println() // Extra line between rows for clarity
	}
	
	fmt.Println("Terrain Key:")
	fmt.Println("🌱=Grass  🏜️=Desert  🌊=Water  ⛰️=Mountains  🗿=Rock  🏥=Hospital")
	fmt.Println("🌾=Swamp  🌲=Forest  🌋=Lava  💧=Shallow  🚀=Missile  🌉=Bridge")
	fmt.Println("⛏️=Mines  🏙️=City  🛣️=Road  🗼=Tower  ❄️=Snow  🏰=Land Base")
	fmt.Println("🏛️=Naval Base  ✈️=Airport  ❓=Unknown")
	fmt.Println()
	fmt.Println("Units: P{Player}U{UnitType} (e.g., P0U1 = Player 0, Unit Type 1), -- = No unit")
	fmt.Println("Hex Layout: Offset rows based on EvenRowsOffset flag")
}

// PrintUnits outputs unit list to console
func (cli *WeeWarCLI) PrintUnits() {
	if cli.game == nil {
		fmt.Println("No game currently loaded")
		return
	}

	fmt.Printf("=== Units ===\n")
	for playerID := 0; playerID < cli.game.PlayerCount; playerID++ {
		units := cli.game.GetUnitsForPlayer(playerID)
		fmt.Printf("Player %d: %d units\n", playerID, len(units))
		
		for i, unit := range units {
			pos := FormatPositionToString(unit.Row, unit.Col)
			unitName := cli.game.GetUnitTypeName(unit.UnitType)
			fmt.Printf("  %d. %s - %s (Type:%d) Health:%d Movement:%d\n", 
				i+1, pos, unitName, unit.UnitType, unit.AvailableHealth, unit.DistanceLeft)
		}
	}
}

// PrintPlayerInfo outputs player statistics
func (cli *WeeWarCLI) PrintPlayerInfo(playerID int) {
	if cli.game == nil {
		fmt.Println("No game currently loaded")
		return
	}

	if playerID < 0 || playerID >= cli.game.PlayerCount {
		fmt.Printf("Invalid player ID: %d\n", playerID)
		return
	}

	fmt.Printf("=== Player %d ===\n", playerID)
	units := cli.game.GetUnitsForPlayer(playerID)
	fmt.Printf("Units: %d\n", len(units))
	
	if playerID == cli.game.GetCurrentPlayer() {
		fmt.Println("Status: Current player")
	} else {
		fmt.Println("Status: Waiting")
	}
	
	// Calculate total health
	totalHealth := 0
	for _, unit := range units {
		totalHealth += unit.AvailableHealth
	}
	fmt.Printf("Total Health: %d\n", totalHealth)
}

// PrintHelp outputs help information
func (cli *WeeWarCLI) PrintHelp(topic string) {
	if topic == "" {
		fmt.Printf("=== WeeWar CLI Help ===\n")
		fmt.Printf("Available commands:\n")
		for _, cmd := range cli.GetAvailableCommands() {
			fmt.Printf("  %s\n", cli.GetCommandHelp(cmd))
		}
		fmt.Printf("\nUse 'help <command>' for detailed help on a specific command.\n")
	} else {
		fmt.Printf("=== Help: %s ===\n", topic)
		fmt.Printf("%s\n", cli.GetCommandHelp(topic))
	}
}

// =============================================================================
// Display Configuration
// =============================================================================

// SetDisplayMode changes output format
func (cli *WeeWarCLI) SetDisplayMode(mode CLIDisplayMode) {
	cli.displayMode = mode
}

// GetDisplayMode returns current display mode
func (cli *WeeWarCLI) GetDisplayMode() CLIDisplayMode {
	return cli.displayMode
}

// SetVerbose enables/disables verbose output
func (cli *WeeWarCLI) SetVerbose(verbose bool) {
	cli.verbose = verbose
}

// IsVerbose returns whether verbose output is enabled
func (cli *WeeWarCLI) IsVerbose() bool {
	return cli.verbose
}

// SetAutoRender enables/disables automatic rendering after commands
func (cli *WeeWarCLI) SetAutoRender(autoRender bool) {
	cli.autoRender = autoRender
}

// SetRenderDir sets the directory for auto-rendered files
func (cli *WeeWarCLI) SetRenderDir(renderDir string) {
	cli.renderDir = renderDir
}

// IsAutoRender returns whether auto-rendering is enabled
func (cli *WeeWarCLI) IsAutoRender() bool {
	return cli.autoRender && cli.maxRenders > 0
}

// SetMaxRenders sets the maximum number of rendered files to keep
func (cli *WeeWarCLI) SetMaxRenders(maxRenders int) {
	cli.maxRenders = maxRenders
	if maxRenders == 0 {
		cli.autoRender = false
	}
}

// GetMaxRenders returns the maximum number of rendered files
func (cli *WeeWarCLI) GetMaxRenders() int {
	return cli.maxRenders
}

// =============================================================================
// Interactive Functions
// =============================================================================

// StartInteractiveMode begins interactive CLI gameplay with proper REPL
func (cli *WeeWarCLI) StartInteractiveMode() {
	cli.interactive = true
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("=== WeeWar Interactive REPL ===")
	fmt.Println("Type 'help' for available commands or 'quit' to exit")
	
	// Show initial game state if game is loaded
	if cli.game != nil {
		cli.showGameState()
	}
	
	for {
		// Show current player prompt
		prompt := cli.getREPLPrompt()
		fmt.Print(prompt)
		
		if !scanner.Scan() {
			break
		}
		
		command := scanner.Text()
		if command == "" {
			continue
		}
		
		// Execute command and handle REPL-specific logic
		response := cli.executeREPLCommand(command)
		
		// Display response
		if response.Success {
			fmt.Printf("✓ %s\n", response.Message)
		} else {
			fmt.Printf("✗ %s\n", response.Message)
			if response.Error != "" {
				fmt.Printf("  Error: %s\n", response.Error)
			}
		}
		
		// Check for quit
		if response.Data == "quit" {
			break
		}
		
		// Show updated game state after successful game actions
		if response.Success && cli.isGameAction(command) {
			cli.showREPLGameState()
			
			// Auto-render game state if enabled
			if cli.autoRender && cli.maxRenders > 0 {
				cli.autoRenderGameState(command)
			}
		}
	}
}

// ProcessTurn handles single player turn interactively
func (cli *WeeWarCLI) ProcessTurn(playerID int) {
	if cli.game == nil {
		fmt.Println("No game loaded")
		return
	}
	
	if playerID != cli.game.GetCurrentPlayer() {
		fmt.Printf("Not player %d's turn\n", playerID)
		return
	}
	
	fmt.Printf("=== Player %d's Turn ===\n", playerID)
	cli.PrintPlayerInfo(playerID)
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Turn> ")
		if !scanner.Scan() {
			break
		}
		
		command := scanner.Text()
		if command == "" {
			continue
		}
		
		response := cli.ExecuteCommand(command)
		fmt.Println(response.Message)
		
		if response.Error != "" {
			fmt.Printf("Error: %s\n", response.Error)
		}
		
		if command == "end" || response.Data == "quit" {
			break
		}
	}
}

// PromptForInput prompts user for input with validation
func (cli *WeeWarCLI) PromptForInput(prompt string, validator func(string) bool) string {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(prompt)
		if !scanner.Scan() {
			return ""
		}
		
		input := scanner.Text()
		if validator == nil || validator(input) {
			return input
		}
		
		fmt.Println("Invalid input, please try again")
	}
}

// ConfirmAction prompts for yes/no confirmation
func (cli *WeeWarCLI) ConfirmAction(message string) bool {
	input := cli.PromptForInput(fmt.Sprintf("%s (y/n): ", message), nil)
	return strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"
}

// =============================================================================
// Game Management
// =============================================================================

// SaveGameToFile saves current game state to file
func (cli *WeeWarCLI) SaveGameToFile(filename string) error {
	if cli.game == nil {
		return fmt.Errorf("no game loaded")
	}
	
	saveData, err := cli.game.SaveGame()
	if err != nil {
		return fmt.Errorf("failed to serialize game: %w", err)
	}
	
	return os.WriteFile(filename, saveData, 0644)
}

// LoadGameFromFile loads game state from file
func (cli *WeeWarCLI) LoadGameFromFile(filename string) error {
	saveData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read save file: %w", err)
	}
	
	game, err := LoadGame(saveData)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}
	
	cli.game = game
	return nil
}

// RenderToFile renders current game state to PNG file
func (cli *WeeWarCLI) RenderToFile(filename string, width, height int) error {
	if cli.game == nil {
		return fmt.Errorf("no game loaded")
	}
	
	// Create buffer
	buffer := NewBuffer(width, height)
	
	// Render game
	if err := cli.game.RenderToBuffer(buffer, 60, 52, 39); err != nil {
		return fmt.Errorf("failed to render game: %w", err)
	}
	
	// Save to file
	return buffer.Save(filename)
}

// =============================================================================
// Batch Processing
// =============================================================================

// ExecuteBatchCommands processes multiple commands from file
func (cli *WeeWarCLI) ExecuteBatchCommands(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open batch file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		fmt.Printf("Executing: %s\n", line)
		response := cli.ExecuteCommand(line)
		
		if !response.Success {
			return fmt.Errorf("batch command failed at line %d: %s", lineNum, response.Error)
		}
		
		fmt.Printf("Result: %s\n", response.Message)
	}
	
	return scanner.Err()
}

// RecordSession records commands to file for replay
func (cli *WeeWarCLI) RecordSession(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create recording file: %w", err)
	}
	
	cli.recordFile = file
	cli.recording = true
	
	// Write header
	fmt.Fprintf(file, "# WeeWar CLI Session Recording\n")
	fmt.Fprintf(file, "# Started: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "\n")
	
	return nil
}

// StopRecording stops current session recording
func (cli *WeeWarCLI) StopRecording() {
	if cli.recordFile != nil {
		fmt.Fprintf(cli.recordFile, "\n# Recording ended: %s\n", time.Now().Format(time.RFC3339))
		cli.recordFile.Close()
		cli.recordFile = nil
	}
	cli.recording = false
}

// =============================================================================
// REPL Helper Functions
// =============================================================================

// getREPLPrompt returns appropriate prompt for current game state
func (cli *WeeWarCLI) getREPLPrompt() string {
	if cli.game == nil {
		return "weewar> "
	}
	
	currentPlayer := cli.game.GetCurrentPlayer()
	turnNumber := cli.game.GetTurnNumber()
	gameStatus := cli.game.GetGameStatus()
	
	// Check if game ended
	if gameStatus == GameStatusEnded {
		if winner, hasWinner := cli.game.GetWinner(); hasWinner {
			return fmt.Sprintf("weewar[GAME ENDED - Player %d Won]> ", winner)
		}
		return "weewar[GAME ENDED]> "
	}
	
	// Show turn and player info
	return fmt.Sprintf("weewar[T%d:P%d]> ", turnNumber, currentPlayer)
}

// executeREPLCommand executes command with REPL-specific enhancements
func (cli *WeeWarCLI) executeREPLCommand(command string) *CLIResponse {
	// Check for REPL-specific commands
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "state", "s":
		// Quick state command
		return cli.handleStatus(nil)
	case "refresh", "r":
		// Refresh display
		cli.showREPLGameState()
		return &CLIResponse{Success: true, Message: "Display refreshed"}
	case "turn":
		// Show detailed turn info
		return cli.handleTurnInfo()
	case "actions":
		// Show available actions
		return cli.handleAvailableActions()
	}
	
	// Execute normal command
	return cli.ExecuteCommand(command)
}

// isGameAction checks if command is a game-affecting action
func (cli *WeeWarCLI) isGameAction(command string) bool {
	cmd := strings.ToLower(strings.TrimSpace(strings.Fields(command)[0]))
	gameActions := []string{"move", "attack", "end", "new", "load"}
	
	for _, action := range gameActions {
		if cmd == action {
			return true
		}
	}
	return false
}

// showGameState displays comprehensive game state
func (cli *WeeWarCLI) showGameState() {
	if cli.game == nil {
		fmt.Println("No game loaded. Use 'new' to create a game or 'load' to load one.")
		return
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	cli.PrintGameState()
	fmt.Println(strings.Repeat("=", 60))
}

// showREPLGameState displays condensed game state for REPL
func (cli *WeeWarCLI) showREPLGameState() {
	if cli.game == nil {
		return
	}
	
	// Show brief status
	currentPlayer := cli.game.GetCurrentPlayer()
	turnNumber := cli.game.GetTurnNumber()
	gameStatus := cli.game.GetGameStatus()
	
	fmt.Printf("\n--- Turn %d | Player %d | Status: %s ---\n", 
		turnNumber, currentPlayer, gameStatus)
	
	// Show current player's units
	units := cli.game.GetUnitsForPlayer(currentPlayer)
	fmt.Printf("Your units: %d | ", len(units))
	
	// Show opponent units
	for i := 0; i < cli.game.PlayerCount; i++ {
		if i != currentPlayer {
			opponentUnits := cli.game.GetUnitsForPlayer(i)
			fmt.Printf("Player %d: %d units | ", i, len(opponentUnits))
		}
	}
	fmt.Println()
	
	// Check for victory conditions
	if gameStatus == GameStatusEnded {
		if winner, hasWinner := cli.game.GetWinner(); hasWinner {
			fmt.Printf("🎉 GAME OVER: Player %d Wins! 🎉\n", winner)
		} else {
			fmt.Println("🎮 GAME OVER: Draw!")
		}
	}
	
	fmt.Println()
}

// handleTurnInfo shows detailed turn information
func (cli *WeeWarCLI) handleTurnInfo() *CLIResponse {
	if cli.game == nil {
		return &CLIResponse{
			Success: false,
			Message: "No game loaded",
			Error:   "Use 'new' to create a game or 'load' to load one",
		}
	}
	
	currentPlayer := cli.game.GetCurrentPlayer()
	turnNumber := cli.game.GetTurnNumber()
	gameStatus := cli.game.GetGameStatus()
	
	var message strings.Builder
	message.WriteString(fmt.Sprintf("=== Turn Information ===\n"))
	message.WriteString(fmt.Sprintf("Turn Number: %d\n", turnNumber))
	message.WriteString(fmt.Sprintf("Current Player: %d\n", currentPlayer))
	message.WriteString(fmt.Sprintf("Game Status: %s\n", gameStatus))
	
	// Show turn capabilities
	canEndTurn := cli.game.CanEndTurn()
	message.WriteString(fmt.Sprintf("Can End Turn: %v\n", canEndTurn))
	
	// Show player stats
	units := cli.game.GetUnitsForPlayer(currentPlayer)
	message.WriteString(fmt.Sprintf("Your Units: %d\n", len(units)))
	
	// Show unit movement status
	unitsWithMovement := 0
	for _, unit := range units {
		if unit.DistanceLeft > 0 {
			unitsWithMovement++
		}
	}
	message.WriteString(fmt.Sprintf("Units with Movement: %d\n", unitsWithMovement))
	
	fmt.Print(message.String())
	return &CLIResponse{
		Success: true,
		Message: "Turn information displayed",
	}
}

// handleAvailableActions shows what player can do
func (cli *WeeWarCLI) handleAvailableActions() *CLIResponse {
	if cli.game == nil {
		return &CLIResponse{
			Success: false,
			Message: "No game loaded",
			Error:   "Use 'new' to create a game or 'load' to load one",
		}
	}
	
	currentPlayer := cli.game.GetCurrentPlayer()
	units := cli.game.GetUnitsForPlayer(currentPlayer)
	
	var message strings.Builder
	message.WriteString(fmt.Sprintf("=== Available Actions (Player %d) ===\n", currentPlayer))
	
	// Show units that can move
	unitsCanMove := 0
	for _, unit := range units {
		if unit.DistanceLeft > 0 {
			unitsCanMove++
			pos := FormatPositionToString(unit.Row, unit.Col)
			message.WriteString(fmt.Sprintf("  Move unit at %s (movement: %d)\n", pos, unit.DistanceLeft))
		}
	}
	
	if unitsCanMove == 0 {
		message.WriteString("  No units can move\n")
	}
	
	// Show units that can attack
	unitsCanAttack := 0
	for _, unit := range units {
		// Check if unit can attack any enemy
		for _, enemy := range cli.game.GetAllUnits() {
			if enemy.PlayerID != currentPlayer && cli.game.CanAttackUnit(unit, enemy) {
				unitsCanAttack++
				pos := FormatPositionToString(unit.Row, unit.Col)
				enemyPos := FormatPositionToString(enemy.Row, enemy.Col)
				message.WriteString(fmt.Sprintf("  Attack with unit at %s -> enemy at %s\n", pos, enemyPos))
				break // Only show first available target per unit
			}
		}
	}
	
	if unitsCanAttack == 0 {
		message.WriteString("  No attack opportunities\n")
	}
	
	// Show turn management
	if cli.game.CanEndTurn() {
		message.WriteString("  End turn (use 'end' command)\n")
	}
	
	// Show utility actions
	message.WriteString("  View map (use 'map' command)\n")
	message.WriteString("  View units (use 'units' command)\n")
	message.WriteString("  Save game (use 'save <filename>' command)\n")
	message.WriteString("  Render game (use 'render <filename>' command)\n")
	
	fmt.Print(message.String())
	return &CLIResponse{
		Success: true,
		Message: "Available actions displayed",
	}
}

// autoRenderGameState automatically renders game state after commands
func (cli *WeeWarCLI) autoRenderGameState(command string) {
	if cli.game == nil || cli.maxRenders == 0 {
		return
	}
	
	// Increment command counter
	cli.commandCount++
	
	// Create render directory if it doesn't exist
	if err := os.MkdirAll(cli.renderDir, 0755); err != nil {
		if cli.verbose {
			fmt.Printf("Warning: Failed to create render directory: %v\n", err)
		}
		return
	}
	
	// Generate filename with sequential numbering
	turnInfo := fmt.Sprintf("T%d_P%d", cli.game.GetTurnNumber(), cli.game.GetCurrentPlayer())
	commandName := strings.Fields(command)[0] // Get first word of command
	
	filename := fmt.Sprintf("%s/game_%03d_%s_%s.png", 
		cli.renderDir, cli.commandCount, turnInfo, commandName)
	
	// Render game state
	if err := cli.RenderToFile(filename, 800, 600); err != nil {
		if cli.verbose {
			fmt.Printf("Warning: Failed to auto-render game state: %v\n", err)
		}
		return
	}
	
	if cli.verbose {
		fmt.Printf("Auto-rendered game state to: %s\n", filename)
	}
	
	// Clean up old files if we exceed maxRenders
	cli.cleanupOldRenders()
}

// cleanupOldRenders removes old render files if we exceed maxRenders
func (cli *WeeWarCLI) cleanupOldRenders() {
	if cli.maxRenders <= 0 {
		return
	}
	
	// List all PNG files in render directory
	files, err := os.ReadDir(cli.renderDir)
	if err != nil {
		if cli.verbose {
			fmt.Printf("Warning: Failed to read render directory: %v\n", err)
		}
		return
	}
	
	// Filter for PNG files matching our pattern
	var renderFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "game_") && strings.HasSuffix(file.Name(), ".png") {
			renderFiles = append(renderFiles, file.Name())
		}
	}
	
	// Sort files by name (which includes the sequential number)
	sort.Strings(renderFiles)
	
	// Remove oldest files if we exceed maxRenders
	if len(renderFiles) > cli.maxRenders {
		filesToRemove := len(renderFiles) - cli.maxRenders
		for i := 0; i < filesToRemove; i++ {
			filePath := fmt.Sprintf("%s/%s", cli.renderDir, renderFiles[i])
			if err := os.Remove(filePath); err != nil {
				if cli.verbose {
					fmt.Printf("Warning: Failed to remove old render file %s: %v\n", filePath, err)
				}
			} else if cli.verbose {
				fmt.Printf("Removed old render file: %s\n", filePath)
			}
		}
	}
}