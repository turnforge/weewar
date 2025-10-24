package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
	weewar "github.com/panyam/turnengine/games/weewar/services"
)

// CLI is a headless command processor for WeeWar games
type CLI struct {
	gameID   string
	service  *services.FSGamesServiceImpl
	readline *readline.Instance // For reading user input with history
}

// NewCLI creates a new CLI instance
func NewCLI(gameID string) (*CLI, error) {
	service := services.NewFSGamesService()

	// Verify game exists by trying to load it
	ctx := context.Background()
	_, err := service.GetGame(ctx, &v1.GetGameRequest{Id: gameID})
	if err != nil {
		return nil, fmt.Errorf("failed to load game %s: %w", gameID, err)
	}

	// Set up history file path
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".weewar_cli_history")

	// Configure readline with auto-complete for commands
	completer := readline.NewPrefixCompleter(
		readline.PcItem("options"),
		readline.PcItem("optionsd"),
		readline.PcItem("move"),
		readline.PcItem("attack"),
		readline.PcItem("end"),
		readline.PcItem("status"),
		readline.PcItem("units"),
		readline.PcItem("player"),
		readline.PcItem("help"),
		readline.PcItem("quit"),
		readline.PcItem("exit"),
	)

	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("weewar[%s]> ", gameID),
		HistoryFile:     historyFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	return &CLI{
		gameID:   gameID,
		service:  service,
		readline: rl,
	}, nil
}

// filterInput allows special characters in readline
func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// Close cleans up the CLI resources
func (cli *CLI) Close() error {
	if cli.readline != nil {
		return cli.readline.Close()
	}
	return nil
}

// ExecuteCommand processes a command and returns the result
func (cli *CLI) ExecuteCommand(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return "Empty command"
	}

	parts := strings.Fields(command)
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "options":
		return cli.handleOptions(args)
	case "optionsd":
		return cli.handleOptionsDetailed(args)
	case "move":
		return cli.handleMove(args)
	case "attack":
		return cli.handleAttack(args)
	case "end":
		return cli.handleEndTurn()
	case "status":
		return cli.handleStatus()
	case "units":
		return cli.handleUnits()
	case "player":
		return cli.handlePlayer(args)
	case "help":
		return cli.handleHelp()
	case "quit", "exit":
		return "quit"
	default:
		return fmt.Sprintf("Unknown command: %s. Type 'help' for available commands.", cmd)
	}
}

// handleOptions shows available options at a position as a menu
func (cli *CLI) handleOptions(args []string) string {
	if len(args) != 1 {
		return "Usage: options <position>\nExample: options 3,4 or options A1"
	}
	return cli.showOptions(args[0], false)
}

// handleOptionsDetailed shows available options with detailed path information
func (cli *CLI) handleOptionsDetailed(args []string) string {
	if len(args) != 1 {
		return "Usage: optionsd <position>\nExample: optionsd 3,4 or optionsd A1"
	}
	return cli.showOptions(args[0], true)
}

// showOptions is the shared implementation for showing options
func (cli *CLI) showOptions(position string, detailed bool) string {
	ctx := context.Background()

	// Get runtime game for parsing
	rtGame, err := cli.service.GetRuntimeGameByID(ctx, cli.gameID)
	if err != nil {
		return fmt.Sprintf("Failed to get game: %v", err)
	}

	// Parse position (could be coordinate or unit ID)
	target, err := weewar.ParsePositionOrUnit(rtGame, position)
	if err != nil {
		return fmt.Sprintf("Invalid position: %v", err)
	}

	coord := target.GetCoordinate()

	// Get options at this position
	resp, err := cli.service.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: cli.gameID,
		Q:      int32(coord.Q),
		R:      int32(coord.R),
	})
	if err != nil {
		return fmt.Sprintf("Failed to get options: %v", err)
	}

	// Build menu of options
	var menuItems []string
	var actions []*v1.GameMove
	menuIndex := 1

	// Process sorted options
	for _, option := range resp.Options {
		switch opt := option.OptionType.(type) {
		case *v1.GameOption_Move:
			moveOpt := opt.Move
			targetCoord := weewar.CoordFromInt32(moveOpt.Q, moveOpt.R)

			// Build the menu item with path if available
			menuItem := fmt.Sprintf("%d. move to %s (cost: %d)",
				menuIndex, targetCoord.String(), moveOpt.MovementCost)

			// Add path visualization if AllPaths is available
			if resp.AllPaths != nil {
				path := moveOpt.ReconstructedPath // weewar.ReconstructPath(resp.AllPaths, moveOpt.Q, moveOpt.R)
				if path != nil {
					if detailed {
						pathStr := weewar.FormatPathDetailed(path, "   ")
						menuItem += "\n" + pathStr
					} else {
						pathStr := weewar.FormatPathCompact(path)
						menuItem += fmt.Sprintf("\n   Path: %s", pathStr)
					}
				}
			}

			menuItems = append(menuItems, menuItem)

			// Create the move action using the provided action
			move := &v1.GameMove{
				MoveType: &v1.GameMove_MoveUnit{
					MoveUnit: moveOpt.Action,
				},
			}
			actions = append(actions, move)
			menuIndex++

		case *v1.GameOption_Attack:
			attackOpt := opt.Attack
			targetCoord := weewar.CoordFromInt32(attackOpt.Q, attackOpt.R)
			menuItems = append(menuItems, fmt.Sprintf("%d. attack %s (type %d, damage est: %d)",
				menuIndex, targetCoord.String(), attackOpt.TargetUnitType, attackOpt.DamageEstimate))

			// Create the attack action using the provided action
			move := &v1.GameMove{
				MoveType: &v1.GameMove_AttackUnit{
					AttackUnit: attackOpt.Action,
				},
			}
			actions = append(actions, move)
			menuIndex++

		case *v1.GameOption_EndTurn:
			menuItems = append(menuItems, fmt.Sprintf("%d. end turn", menuIndex))
			move := &v1.GameMove{
				MoveType: &v1.GameMove_EndTurn{
					EndTurn: &v1.EndTurnAction{},
				},
			}
			actions = append(actions, move)
			menuIndex++

		case *v1.GameOption_Build:
			// TODO: Handle build options when needed

		case *v1.GameOption_Capture:
			// TODO: Handle capture options when needed
		}
	}

	// Show menu
	if len(menuItems) == 0 {
		return "No options available at this position"
	}

	// Show what's at this position
	if target.IsUnit && target.Unit != nil {
		unitID := target.Unit.Shortcut
		if unitID == "" {
			unitID = target.Raw
		}
		fmt.Printf("\nUnit %s at %s:\n", unitID, coord.String())
		fmt.Printf("  Type: %d, HP: %d, Moves: %d\n",
			target.Unit.UnitType, target.Unit.AvailableHealth, target.Unit.DistanceLeft)
	} else {
		fmt.Printf("\nPosition %s:\n", coord.String())
	}

	if detailed {
		fmt.Println("\nAvailable options (detailed):")
	} else {
		fmt.Println("\nAvailable options:")
	}
	for _, item := range menuItems {
		fmt.Println(item)
	}
	// Temporarily change prompt for selection
	oldPrompt := cli.readline.Config.Prompt
	defer cli.readline.SetPrompt(oldPrompt)
	cli.readline.SetPrompt(fmt.Sprintf("Select option (1-%d) or Enter to cancel: ", len(menuItems)))

	// Read user selection
	selection, err := cli.readline.Readline()
	if err != nil {
		if err == readline.ErrInterrupt || err == io.EOF {
			return "Cancelled"
		}
		return "Failed to read input"
	}
	selection = strings.TrimSpace(selection)
	if selection == "" {
		return "Cancelled"
	}

	// Parse selection
	index, err := strconv.Atoi(selection)
	if err != nil || index < 1 || index > len(actions) {
		return "Invalid selection"
	}

	// Execute the selected action
	selectedAction := actions[index-1]
	return cli.processMoves([]*v1.GameMove{selectedAction})
}

func (cli *CLI) ParseFromAndToCoords(arg0, arg1 string) (fromCoord services.AxialCoord, toCoord services.AxialCoord, err error) {
	// Get runtime game for parsing
	ctx := context.Background()
	rtGame, err := cli.service.GetRuntimeGameByID(ctx, cli.gameID)
	if err != nil {
		err = fmt.Errorf("Failed to get game: %v", err)
		return
	}

	// Parse from position
	fromTarget, err := weewar.ParsePositionOrUnit(rtGame, arg0)
	if err != nil {
		err = fmt.Errorf("Invalid from position: %v", err)
		return
	}

	fromCoord = fromTarget.GetCoordinate()

	// Parse to position with context (supports directions like L, R, TL, etc.)
	toTarget, err := weewar.ParsePositionOrUnitWithContext(rtGame, arg1, &fromCoord)
	if err != nil {
		err = fmt.Errorf("Invalid to position: %v", err)
		return
	}

	toCoord = toTarget.GetCoordinate()
	return
}

// handleMove processes move command
func (cli *CLI) handleMove(args []string) string {
	if len(args) != 2 {
		return "Usage: move <from> <to>\nExample: move A1 5,6 or move A1 R or move 3,4 L"
	}

	fromCoord, toCoord, err := cli.ParseFromAndToCoords(args[0], args[1])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	// Create move action
	move := &v1.GameMove{
		MoveType: &v1.GameMove_MoveUnit{
			MoveUnit: &v1.MoveUnitAction{
				FromQ: int32(fromCoord.Q),
				FromR: int32(fromCoord.R),
				ToQ:   int32(toCoord.Q),
				ToR:   int32(toCoord.R),
			},
		},
	}

	return cli.processMoves([]*v1.GameMove{move})
}

// handleAttack processes attack command
func (cli *CLI) handleAttack(args []string) string {
	if len(args) != 2 {
		return "Usage: attack <attacker> <target>\nExample: attack A1 B2 or attack A1 R or attack 3,4 TL"
	}

	attackerCoord, targetCoord, err := cli.ParseFromAndToCoords(args[0], args[1])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	// Create attack action
	move := &v1.GameMove{
		MoveType: &v1.GameMove_AttackUnit{
			AttackUnit: &v1.AttackUnitAction{
				AttackerQ: int32(attackerCoord.Q),
				AttackerR: int32(attackerCoord.R),
				DefenderQ: int32(targetCoord.Q),
				DefenderR: int32(targetCoord.R),
			},
		},
	}

	return cli.processMoves([]*v1.GameMove{move})
}

// handleEndTurn processes end turn command
func (cli *CLI) handleEndTurn() string {
	// Create end turn action
	move := &v1.GameMove{
		MoveType: &v1.GameMove_EndTurn{
			EndTurn: &v1.EndTurnAction{},
		},
	}

	return cli.processMoves([]*v1.GameMove{move})
}

// processMoves sends moves to the service via ProcessMoves RPC
func (cli *CLI) processMoves(moves []*v1.GameMove) string {
	ctx := context.Background()

	// Call ProcessMoves
	resp, err := cli.service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: cli.gameID,
		Moves:  moves,
	})
	if err != nil {
		return fmt.Sprintf("Failed to process moves: %v", err)
	}

	// Build result message
	var result strings.Builder

	for i, moveResult := range resp.MoveResults {
		result.WriteString(fmt.Sprintf("Move %d: ", i+1))

		// Check if there are changes (success) or not (failure)
		if len(moveResult.Changes) > 0 {
			result.WriteString("Success\n")

			// Add specific details based on action type
			if len(moves) > i {
				switch action := moves[i].MoveType.(type) {
				case *v1.GameMove_MoveUnit:
					fromCoord := weewar.CoordFromInt32(action.MoveUnit.FromQ, action.MoveUnit.FromR)
					toCoord := weewar.CoordFromInt32(action.MoveUnit.ToQ, action.MoveUnit.ToR)
					result.WriteString(fmt.Sprintf("  Moved unit from %s to %s\n",
						fromCoord.String(), toCoord.String()))

				case *v1.GameMove_AttackUnit:
					attackerCoord := weewar.CoordFromInt32(action.AttackUnit.AttackerQ, action.AttackUnit.AttackerR)
					targetCoord := weewar.CoordFromInt32(action.AttackUnit.DefenderQ, action.AttackUnit.DefenderR)
					result.WriteString(fmt.Sprintf("  Attacked from %s to %s\n",
						attackerCoord.String(), targetCoord.String()))

					// Look for damage in the changes
					for _, change := range moveResult.Changes {
						switch c := change.ChangeType.(type) {
						case *v1.WorldChange_UnitDamaged:
							prevHealth := c.UnitDamaged.PreviousUnit.AvailableHealth
							newHealth := c.UnitDamaged.UpdatedUnit.AvailableHealth
							damage := prevHealth - newHealth
							result.WriteString(fmt.Sprintf("  Unit took %d damage (HP: %d -> %d)\n",
								damage, prevHealth, newHealth))
						case *v1.WorldChange_UnitKilled:
							result.WriteString("  Unit destroyed!\n")
						}
					}

				case *v1.GameMove_EndTurn:
					result.WriteString("  Turn ended\n")
					// Look for player change
					for _, change := range moveResult.Changes {
						if pc, ok := change.ChangeType.(*v1.WorldChange_PlayerChanged); ok {
							result.WriteString(fmt.Sprintf("  Now player %d's turn\n",
								pc.PlayerChanged.NewPlayer))
						}
					}
				}
			}
		} else {
			result.WriteString("No changes (possibly invalid move)\n")
		}

		if moveResult.IsPermanent {
			result.WriteString("  (This action is permanent and cannot be undone)\n")
		}
	}

	// Get updated game state to show current status
	gameResp, err := cli.service.GetGame(ctx, &v1.GetGameRequest{Id: cli.gameID})
	if err == nil && gameResp.State != nil {
		result.WriteString(fmt.Sprintf("\nCurrent player: %d, Turn: %d\n",
			gameResp.State.CurrentPlayer, gameResp.State.TurnCounter))

		if gameResp.State.WinningPlayer != 0 {
			result.WriteString(fmt.Sprintf("Game Over! WinningPlayer: Player %d\n", gameResp.State.WinningPlayer))
		}
	}

	return result.String()
}

// handleStatus shows current game status
func (cli *CLI) handleStatus() string {
	ctx := context.Background()

	// Get game state
	resp, err := cli.service.GetGame(ctx, &v1.GetGameRequest{Id: cli.gameID})
	if err != nil {
		return fmt.Sprintf("Failed to get game: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Game ID: %s\n", cli.gameID))
	result.WriteString(fmt.Sprintf("Turn: %d\n", resp.State.TurnCounter))
	result.WriteString(fmt.Sprintf("Current Player: %d\n", resp.State.CurrentPlayer))
	result.WriteString(fmt.Sprintf("Game Status: %s\n", resp.State.Status))

	// Count units per player
	unitCounts := make(map[int32]int)
	if resp.State.WorldData != nil && resp.State.WorldData.Units != nil {
		for _, unit := range resp.State.WorldData.Units {
			if unit != nil {
				unitCounts[unit.Player]++
			}
		}
	}

	for playerID, count := range unitCounts {
		result.WriteString(fmt.Sprintf("Player %c: %d units\n", 'A'+playerID, count))
	}

	if resp.State.WinningPlayer != 0 {
		result.WriteString(fmt.Sprintf("\nGame Over! WinningPlayer: Player %d\n", resp.State.WinningPlayer))
	}

	return result.String()
}

// handleUnits shows all units
func (cli *CLI) handleUnits() string {
	ctx := context.Background()

	// Get runtime game to get units with shortcuts
	rtGame, err := cli.service.GetRuntimeGameByID(ctx, cli.gameID)
	if err != nil {
		return fmt.Sprintf("Failed to get game: %v", err)
	}

	if rtGame.World.NumUnits() == 0 {
		return "No units found"
	}

	// Group units by player from runtime world
	unitsByPlayer := make(map[int32][]*v1.Unit)
	for _, unit := range rtGame.World.UnitsByCoord() {
		if unit != nil {
			unitsByPlayer[unit.Player] = append(unitsByPlayer[unit.Player], unit)
		}
	}

	var result strings.Builder
	for playerID, units := range unitsByPlayer {
		playerLetter := string(rune('A' + playerID))
		result.WriteString(fmt.Sprintf("Player %s units:\n", playerLetter))

		for _, unit := range units {
			coord := weewar.CoordFromInt32(unit.Q, unit.R)
			// Use the actual shortcut from the unit
			unitID := unit.Shortcut
			if unitID == "" {
				// Fallback if no shortcut (shouldn't happen)
				unitID = fmt.Sprintf("%s?", playerLetter)
			}
			result.WriteString(fmt.Sprintf("  %s: Type %d at %s (HP: %d, Moves: %d)\n",
				unitID, unit.UnitType, coord.String(),
				unit.AvailableHealth, unit.DistanceLeft))
		}
	}

	return result.String()
}

// handlePlayer shows player information
func (cli *CLI) handlePlayer(args []string) string {
	ctx := context.Background()

	// Get game state
	resp, err := cli.service.GetGame(ctx, &v1.GetGameRequest{Id: cli.gameID})
	if err != nil {
		return fmt.Sprintf("Failed to get game: %v", err)
	}

	playerID := resp.State.CurrentPlayer
	if len(args) > 0 {
		// Parse player ID if provided
		if len(args[0]) == 1 {
			playerLetter := strings.ToUpper(args[0])[0]
			if playerLetter >= 'A' && playerLetter <= 'Z' {
				playerID = int32(playerLetter - 'A')
			}
		}
	}

	// Count units for this player
	unitCount := 0
	if resp.State.WorldData != nil && resp.State.WorldData.Units != nil {
		for _, unit := range resp.State.WorldData.Units {
			if unit != nil && unit.Player == playerID {
				unitCount++
			}
		}
	}

	playerLetter := string(rune('A' + playerID))

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Player %s:\n", playerLetter))
	result.WriteString(fmt.Sprintf("  Units: %d\n", unitCount))

	if playerID == resp.State.CurrentPlayer {
		result.WriteString("  Status: Active (current turn)\n")
	} else {
		result.WriteString("  Status: Waiting\n")
	}

	return result.String()
}

// handleHelp shows available commands
func (cli *CLI) handleHelp() string {
	return `Available commands:
  options <pos>        - Show available actions at position (interactive menu)
  optionsd <pos>       - Show available actions with detailed path info
  move <from> <to>     - Move unit (e.g. "move A1 5,6" or "move A1 R")
  attack <att> <tgt>   - Attack target (e.g. "attack A1 B2" or "attack A1 TL")
  end                  - End current player's turn
  status               - Show game status
  units                - Show all units
  player [ID]          - Show player information (e.g. "player A")
  help                 - Show this help
  quit                 - Exit game

Position formats:
  - Unit ID: A1, B12, C2 (Player letter + unit number)
  - Q,R coordinate: 3,4 or -1,2
  - Row/col coordinate: r4,5 (prefix with 'r')
  - Direction: L, R, TL, TR, BL, BR (relative to from position)

Direction codes (for move/attack targets):
  L  = Left            R  = Right
  TL = Top-Left        TR = Top-Right
  BL = Bottom-Left     BR = Bottom-Right

Examples:
  options A1           # Show menu of available actions for unit A1
  optionsd A1          # Show detailed menu with path breakdowns
  options 3,4          # Show menu of available actions at position 3,4
  move A1 5,6          # Move unit A1 to position 5,6
  move A1 R            # Move unit A1 to the right
  move A1 TL           # Move unit A1 to top-left neighbor
  attack A1 B2         # Attack unit B2 with unit A1
  attack A1 TR         # Attack top-right neighbor with unit A1
  attack 3,4 5,6       # Attack position 5,6 with unit at 3,4
  end                  # End current turn`
}
