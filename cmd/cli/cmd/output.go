package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
)

// OutputFormatter handles formatting output in text or JSON
type OutputFormatter struct {
	JSON   bool
	Dryrun bool
}

// NewOutputFormatter creates a new formatter based on global flags
func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{
		JSON:   isJSONOutput(),
		Dryrun: isDryrun(),
	}
}

// prefix adds [DRYRUN] prefix if in dryrun mode
func (f *OutputFormatter) prefix(text string) string {
	if f.Dryrun {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = "[DRYRUN] " + line
			}
		}
		return strings.Join(lines, "\n")
	}
	return text
}

// Print outputs text or JSON based on format setting
func (f *OutputFormatter) Print(data any) error {
	if f.JSON {
		return f.PrintJSON(data)
	}
	return f.PrintText(data)
}

// PrintJSON outputs data as JSON
func (f *OutputFormatter) PrintJSON(data any) error {
	output := map[string]any{
		"data":   data,
		"dryrun": f.Dryrun,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// PrintText outputs data as human-readable text
func (f *OutputFormatter) PrintText(data any) error {
	var text string

	switch v := data.(type) {
	case string:
		text = v
	case fmt.Stringer:
		text = v.String()
	default:
		text = fmt.Sprintf("%v", v)
	}

	fmt.Println(f.prefix(text))
	return nil
}

// FormatOptions formats available options as text
func FormatOptions(pc *PresenterContext, position string) string {
	var sb strings.Builder

	// Get unit info
	if pc.TurnOptions.Unit != nil {
		unit := pc.TurnOptions.Unit
		coord := services.CoordFromInt32(unit.Q, unit.R)
		sb.WriteString(fmt.Sprintf("Unit %s at %s:\n", position, coord.String()))
		sb.WriteString(fmt.Sprintf("  Type: %d, HP: %d, Moves: %f\n\n",
			unit.UnitType, unit.AvailableHealth, unit.DistanceLeft))
	}

	// Get options
	if pc.TurnOptions.Options == nil || len(pc.TurnOptions.Options.Options) == 0 {
		sb.WriteString("No options available at this position\n")
		return sb.String()
	}

	sb.WriteString("Available options:\n")

	for i, option := range pc.TurnOptions.Options.Options {
		switch opt := option.OptionType.(type) {
		case *v1.GameOption_Move:
			moveOpt := opt.Move
			targetCoord := services.CoordFromInt32(moveOpt.Action.ToQ, moveOpt.Action.ToR)
			sb.WriteString(fmt.Sprintf("%d. move to %s (cost: %f)\n",
				i+1, targetCoord.String(), moveOpt.MovementCost))

			// Add path if available
			if moveOpt.ReconstructedPath != nil {
				pathStr := services.FormatPathCompact(moveOpt.ReconstructedPath)
				sb.WriteString(fmt.Sprintf("   Path: %s\n", pathStr))
			}

		case *v1.GameOption_Attack:
			attackOpt := opt.Attack
			targetCoord := services.CoordFromInt32(attackOpt.Action.DefenderQ, attackOpt.Action.DefenderR)
			sb.WriteString(fmt.Sprintf("%d. attack %s (damage est: %d)\n",
				i+1, targetCoord.String(), attackOpt.DamageEstimate))

		case *v1.GameOption_EndTurn:
			sb.WriteString(fmt.Sprintf("%d. end turn\n", i+1))
		}
	}

	return sb.String()
}

// FormatGameStatus formats game status as text
func FormatGameStatus(state *v1.GameState) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Turn: %d\n", state.TurnCounter))
	sb.WriteString(fmt.Sprintf("Current Player: %d\n", state.CurrentPlayer))
	sb.WriteString(fmt.Sprintf("Game Status: %s\n", state.Status))

	if state.WinningPlayer != 0 {
		sb.WriteString(fmt.Sprintf("\nGame Over! Winner: Player %d\n", state.WinningPlayer))
	}

	return sb.String()
}

// FormatUnits formats all units as text
func FormatUnits(pc *PresenterContext, state *v1.GameState) string {
	if state.WorldData == nil || len(state.WorldData.Units) == 0 {
		return "No units found\n"
	}

	rtGame, err := pc.Presenter.GamesService.GetRuntimeGame(
		pc.Presenter.GamesService.SingletonGame,
		pc.Presenter.GamesService.SingletonGameState)
	if err != nil {
		panic(err)
	}

	// Group units by player
	unitsByPlayer := make(map[int32][]*v1.Unit)
	numPlayers := int32(0)

	for _, unit := range state.WorldData.Units {
		if unit != nil {
			if err := rtGame.TopUpUnitIfNeeded(unit); err != nil {
				panic(err)
			}
			unitsByPlayer[unit.Player] = append(unitsByPlayer[unit.Player], unit)
			if unit.Player > numPlayers {
				numPlayers = unit.Player
			}
		}
	}

	var sb strings.Builder
	// Iterate starting from current player and wrap around
	for i := int32(0); i < numPlayers; i++ {
		playerID := (i + state.CurrentPlayer) % numPlayers
		if playerID == 0 {
			playerID = numPlayers
		}
		units := unitsByPlayer[playerID]
		if len(units) == 0 {
			continue
		}
		turnIndicator := ""
		if playerID == state.CurrentPlayer {
			turnIndicator = " *"
		}
		sb.WriteString(fmt.Sprintf("Player %d units%s:\n", playerID, turnIndicator))

		for _, unit := range units {
			coord := services.CoordFromInt32(unit.Q, unit.R)
			unitID := unit.Shortcut
			if unitID == "" {
				playerLetter := string(rune('A' + playerID))
				unitID = fmt.Sprintf("%s?", playerLetter)
			}
			sb.WriteString(fmt.Sprintf("  %s: Type %d at %s (HP: %d, Moves: %f)\n",
				unitID, unit.UnitType, coord.String(),
				unit.AvailableHealth, unit.DistanceLeft))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
