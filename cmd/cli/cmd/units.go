package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// unitsCmd represents the units command
var unitsCmd = &cobra.Command{
	Use:   "units",
	Short: "List all units in the game",
	Long: `Display all units grouped by player, showing their position, health,
and remaining movement points.

Examples:
  ww units
  ww units --json`,
	RunE: runUnits,
}

func init() {
	rootCmd.AddCommand(unitsCmd)
}

func runUnits(cmd *cobra.Command, args []string) error {
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if gc.State == nil {
		return fmt.Errorf("game state not initialized")
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		// JSON output
		rulesEngine := gc.RTGame.GetRulesEngine()
		units := []map[string]any{}
		if gc.State.WorldData != nil {
			for _, unit := range gc.State.WorldData.UnitsMap {
				if unit != nil {
					unitName := ""
					if unitDef, err := rulesEngine.GetUnitData(unit.UnitType); err == nil {
						unitName = unitDef.Name
					}
					units = append(units, map[string]any{
						"player":           unit.Player,
						"shortcut":         unit.Shortcut,
						"q":                unit.Q,
						"r":                unit.R,
						"unit_type":        unit.UnitType,
						"unit_name":        unitName,
						"available_health": unit.AvailableHealth,
						"distance_left":    unit.DistanceLeft,
					})
				}
			}
		}

		data := map[string]any{
			"game_id": gc.GameID,
			"units":   units,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	text := FormatUnitsWithContext(gc)
	return formatter.PrintText(text)
}
