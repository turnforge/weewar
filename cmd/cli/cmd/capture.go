package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
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
	unitLabel := args[0]

	ctx := context.Background()
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attempting capture at %s\n", unitLabel)
	}

	// Execute capture directly via ProcessMoves - server parses labels
	resp, err := gc.Service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gc.GameID,
		DryRun: isDryrun(),
		Moves: []*v1.GameMove{{
			Player: gc.State.CurrentPlayer,
			MoveType: &v1.GameMove_CaptureBuilding{
				CaptureBuilding: &v1.CaptureBuildingAction{
					Pos: &v1.Position{Label: unitLabel},
				},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("capture failed: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gc.GameID,
			"action":  "capture",
			"unit":    unitLabel,
			"dryrun":  isDryrun(),
			"success": true,
			"changes": formatChangesForJSON(resp.Moves),
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	if isDryrun() {
		sb.WriteString("Capture (dryrun): Would succeed\n")
	} else {
		sb.WriteString("Capture: Started\n")
	}
	sb.WriteString(fmt.Sprintf("  Unit at %s is capturing\n", unitLabel))
	sb.WriteString("  Capture will complete at the start of your next turn if the unit survives.\n")

	// Show changes from response
	if len(resp.Moves) > 0 && len(resp.Moves[0].Changes) > 0 {
		sb.WriteString("  Changes:\n")
		for _, change := range resp.Moves[0].Changes {
			sb.WriteString(fmt.Sprintf("    - %s\n", formatChange(change)))
		}
	}

	return formatter.PrintText(sb.String())
}
