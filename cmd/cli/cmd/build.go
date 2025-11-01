package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build <tile> <unit_type>",
	Short: "Build a unit at a tile",
	Long: `Build a new unit at the specified tile (typically a city).
The tile position can be a tile ID (like t:A1), unit shortcut (like A1), or coordinates (like 3,4).
The unit_type can be a unit type ID number or unit name.

Examples:
  ww build t:A1 trooper         Build a trooper at tile A1
  ww build 3,4 5                Build unit type 5 at coordinates 3,4
  ww build A1 tank              Build a tank at tile with same position as unit A1
  ww build t:A1 tank --dryrun   Preview build without saving`,
	Args: cobra.ExactArgs(2),
	RunE: runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	tilePos := args[0]
	unitTypeArg := args[1]

	// Get game ID
	ctx := context.Background()
	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		return err
	}

	// Parse tile position
	target, err := services.ParsePositionOrUnit(rtGame, tilePos)
	if err != nil {
		return fmt.Errorf("invalid tile position: %w", err)
	}
	coord := target.GetCoordinate()

	// Parse unit type (can be numeric ID or name)
	unitType, err := parseUnitType(rtGame, unitTypeArg)
	if err != nil {
		return fmt.Errorf("invalid unit type: %w", err)
	}

	// Get unit information for confirmation
	unitData, err := rtGame.GetRulesEngine().GetUnitData(unitType)
	if err != nil || unitData == nil {
		return fmt.Errorf("unit type %d not found", unitType)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Building %s (type %d) at %s\n", unitData.Name, unitType, coord.String())
	}

	// Confirmation prompt (unless in dryrun or disabled with --confirm=false)
	if !isDryrun() && shouldConfirm() {
		fmt.Printf("You will build a %s costing %d coins. Are you sure? (y/n): ", unitData.Name, unitData.Coins)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return fmt.Errorf("build cancelled")
		}
	}

	// Call BuildOptionClicked through the presenter
	_, err = pc.Presenter.BuildOptionClicked(ctx, &v1.BuildOptionClickedRequest{
		GameId:   gameID,
		Q:        int32(coord.Q),
		R:        int32(coord.R),
		UnitType: unitType,
	})
	if err != nil {
		return fmt.Errorf("failed to build unit: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Build executed successfully\n")
	}

	// Save state unless in dryrun mode
	if err := savePresenterState(pc, isDryrun()); err != nil {
		return err
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id":   gameID,
			"action":    "build",
			"unit_type": unitType,
			"unit_name": unitData.Name,
			"tile": map[string]int{
				"q": coord.Q,
				"r": coord.R,
			},
			"success": true,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder

	sb.WriteString("Build: Success\n")
	sb.WriteString(fmt.Sprintf("  Built %s at %s\n", unitData.Name, coord.String()))

	// Show the newly created unit
	newUnit := rtGame.World.UnitAt(coord)
	if newUnit != nil {
		sb.WriteString(fmt.Sprintf("  New unit: %s (Player %d, Health: %d)\n",
			newUnit.Shortcut, newUnit.Player, newUnit.AvailableHealth))
	}

	sb.WriteString(fmt.Sprintf("\nCurrent player: %d, Turn: %d\n",
		pc.GameState.State.CurrentPlayer, pc.GameState.State.TurnCounter))

	return formatter.PrintText(sb.String())
}

// parseUnitType parses a unit type argument which can be:
// - A numeric unit type ID (e.g., "5")
// - A unit name (e.g., "trooper", "tank")
func parseUnitType(rtGame *services.Game, unitTypeArg string) (int32, error) {
	// Try parsing as numeric ID first
	if id, err := strconv.ParseInt(unitTypeArg, 10, 32); err == nil {
		return int32(id), nil
	}

	// Otherwise, search by name
	rulesEngine := rtGame.GetRulesEngine()
	if rulesEngine == nil {
		return 0, fmt.Errorf("rules engine not available")
	}

	// Search through all unit definitions
	unitTypeArg = strings.ToLower(strings.TrimSpace(unitTypeArg))
	for unitID, unitDef := range rulesEngine.RulesEngine.Units {
		if strings.ToLower(unitDef.Name) == unitTypeArg {
			return unitID, nil
		}
	}

	return 0, fmt.Errorf("unit type not found: %s", unitTypeArg)
}
