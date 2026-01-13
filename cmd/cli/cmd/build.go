package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
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
	tileLabel := args[0]
	unitTypeArg := args[1]

	ctx := context.Background()
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	// Parse unit type (can be numeric ID or name)
	unitType, err := parseUnitType(gc, unitTypeArg)
	if err != nil {
		return fmt.Errorf("invalid unit type: %w", err)
	}

	// Get unit information for confirmation
	unitData, err := gc.RTGame.GetRulesEngine().GetUnitData(unitType)
	if err != nil || unitData == nil {
		return fmt.Errorf("unit type %d not found", unitType)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Building %s (type %d) at %s\n", unitData.Name, unitType, tileLabel)
	}

	// Confirmation prompt (unless in dryrun or disabled with --confirm=false)
	if !isDryrun() && shouldConfirm() {
		fmt.Printf("Build %s for %d coins? (y/n): ", unitData.Name, unitData.Coins)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return fmt.Errorf("build cancelled")
		}
	}

	// Execute build directly via ProcessMoves - server parses labels
	resp, err := gc.Service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gc.GameID,
		DryRun: isDryrun(),
		Moves: []*v1.GameMove{{
			Player: gc.State.CurrentPlayer,
			MoveType: &v1.GameMove_BuildUnit{
				BuildUnit: &v1.BuildUnitAction{
					Pos:      &v1.Position{Label: tileLabel},
					UnitType: unitType,
				},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id":   gc.GameID,
			"action":    "build",
			"tile":      tileLabel,
			"unit_type": unitType,
			"unit_name": unitData.Name,
			"dryrun":    isDryrun(),
			"success":   true,
			"changes":   formatChangesForJSON(resp.Moves),
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	if isDryrun() {
		sb.WriteString("Build (dryrun): Would succeed\n")
	} else {
		sb.WriteString("Build: Success\n")
	}
	sb.WriteString(fmt.Sprintf("  Built %s at %s\n", unitData.Name, tileLabel))

	// Show changes from response
	if len(resp.Moves) > 0 && len(resp.Moves[0].Changes) > 0 {
		sb.WriteString("  Changes:\n")
		for _, change := range resp.Moves[0].Changes {
			sb.WriteString(fmt.Sprintf("    - %s\n", formatChange(change)))
		}
	}

	return formatter.PrintText(sb.String())
}

// parseUnitType parses a unit type argument which can be:
// - A numeric unit type ID (e.g., "5")
// - A unit name (e.g., "trooper", "tank")
func parseUnitType(gc *GameContext, unitTypeArg string) (int32, error) {
	// Try parsing as numeric ID first
	if id, err := strconv.ParseInt(unitTypeArg, 10, 32); err == nil {
		return int32(id), nil
	}

	// Otherwise, search by name
	rulesEngine := gc.RTGame.GetRulesEngine()
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
