package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// captureCmd represents the capture command
var captureCmd = &cobra.Command{
	Use:   "capture <unit>",
	Short: "Start capturing a building with a unit",
	Long: `Start capturing a building (base, harbor, etc.) with your unit.
The unit must be on a capturable tile that it doesn't already own.
Only certain unit types (like soldiers and hovercrafts) can capture buildings.

The capture completes at the start of your next turn if the unit survives.

Positions can be unit IDs (like A1) or coordinates (like 3,4).

Examples:
  ww capture A1              Start capturing with unit A1
  ww capture 3,4             Start capturing at coordinates 3,4
  ww capture A1 --dryrun     Preview capture without saving`,
	Args: cobra.ExactArgs(1),
	RunE: runCapture,
}

func init() {
	rootCmd.AddCommand(captureCmd)
}

func runCapture(cmd *cobra.Command, args []string) error {
	unitPos := args[0]

	ctx := context.Background()
	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		return err
	}

	// Parse unit position
	target, err := lib.ParsePositionOrUnit(rtGame, unitPos)
	if err != nil {
		return fmt.Errorf("invalid unit position: %w", err)
	}
	coord := target.GetCoordinate()

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attempting capture at %s\n", coord.String())
	}

	// Two-click pattern: Click unit to select, then click same position to capture
	// Click 1: Select unit on base-map layer
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(coord.Q),
		R:      int32(coord.R),
		Layer:  "base-map",
	})
	if err != nil {
		return fmt.Errorf("failed to select unit: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Unit selected at %s\n", coord.String())
	}

	// Click 2: Click same position on capture-highlight layer to execute capture
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(coord.Q),
		R:      int32(coord.R),
		Layer:  "capture-highlight",
	})
	if err != nil {
		return fmt.Errorf("failed to execute capture: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Capture started at %s\n", coord.String())
	}

	// Save state unless in dryrun mode
	if err := savePresenterState(pc, isDryrun()); err != nil {
		return err
	}

	// Get display info
	unit := rtGame.World.UnitAt(coord)
	unitShortcut := "?"
	if unit != nil {
		unitShortcut = unit.Shortcut
	}

	tile := rtGame.World.TileAt(coord)
	terrainName := "unknown"
	if tile != nil {
		if terrainData, err := rtGame.GetRulesEngine().GetTerrainData(tile.TileType); err == nil && terrainData != nil {
			terrainName = terrainData.Name
		}
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gameID,
			"action":  "capture",
			"unit": map[string]any{
				"shortcut": unitShortcut,
				"q":        coord.Q,
				"r":        coord.R,
			},
			"tile": map[string]any{
				"q":            coord.Q,
				"r":            coord.R,
				"terrain_name": terrainName,
			},
			"success": true,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	sb.WriteString("Capture: Started\n")
	sb.WriteString(fmt.Sprintf("  Unit %s is capturing %s at %s\n",
		unitShortcut, terrainName, coord.String()))
	sb.WriteString("  Capture will complete at the start of your next turn if the unit survives.\n")
	sb.WriteString(fmt.Sprintf("\nCurrent player: %d, Turn: %d\n",
		pc.GameState.State.CurrentPlayer, pc.GameState.State.TurnCounter))

	return formatter.PrintText(sb.String())
}
