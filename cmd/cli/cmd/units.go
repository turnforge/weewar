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
	// Get game ID
	gameID, err := getGameID()
	if err != nil {
		return err
	}

	// Create presenter
	pc, err := createPresenter(gameID)
	if err != nil {
		return err
	}

	// Get state from panel
	if pc.GameState.State == nil {
		return fmt.Errorf("game state not initialized")
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		// JSON output
		units := []map[string]any{}
		if pc.GameState.State.WorldData != nil {
			for _, unit := range pc.GameState.State.WorldData.Units {
				if unit != nil {
					units = append(units, map[string]any{
						"player":           unit.Player,
						"shortcut":         unit.Shortcut,
						"q":                unit.Q,
						"r":                unit.R,
						"unit_type":        unit.UnitType,
						"available_health": unit.AvailableHealth,
						"distance_left":    unit.DistanceLeft,
					})
				}
			}
		}

		data := map[string]any{
			"game_id": gameID,
			"units":   units,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	text := FormatUnits(pc, pc.GameState.State)
	return formatter.PrintText(text)
}
