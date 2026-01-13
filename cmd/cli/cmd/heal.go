package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// healCmd represents the heal command
var healCmd = &cobra.Command{
	Use:   "heal <unit>",
	Short: "Heal a unit on a friendly or neutral terrain",
	Long: `Heal a unit that is on friendly or neutral terrain.
The unit must not have moved or attacked this turn.
Healing amount depends on the terrain type.
Air units can only heal on Airport Bases.
Units cannot heal on enemy-owned tiles.

Positions can be unit IDs (like A1) or coordinates (like 3,4).

Examples:
  ww heal A1              Heal unit A1
  ww heal 3,4             Heal unit at coordinates 3,4
  ww heal A1 --dryrun     Preview heal without saving`,
	Args: cobra.ExactArgs(1),
	RunE: runHeal,
}

func init() {
	rootCmd.AddCommand(healCmd)
}

func runHeal(cmd *cobra.Command, args []string) error {
	unitLabel := args[0]

	ctx := context.Background()
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attempting heal at %s\n", unitLabel)
	}

	// Execute heal directly via ProcessMoves - server parses labels
	resp, err := gc.Service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gc.GameID,
		DryRun: isDryrun(),
		Moves: []*v1.GameMove{{
			Player: gc.State.CurrentPlayer,
			MoveType: &v1.GameMove_HealUnit{
				HealUnit: &v1.HealUnitAction{
					Pos: &v1.Position{Label: unitLabel},
				},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("heal failed: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gc.GameID,
			"action":  "heal",
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
		sb.WriteString("Heal (dryrun): Would succeed\n")
	} else {
		sb.WriteString("Heal: Success\n")
	}

	// Show changes from response
	if len(resp.Moves) > 0 && len(resp.Moves[0].Changes) > 0 {
		for _, change := range resp.Moves[0].Changes {
			sb.WriteString(fmt.Sprintf("  %s\n", formatChange(change)))
		}
	}

	return formatter.PrintText(sb.String())
}
