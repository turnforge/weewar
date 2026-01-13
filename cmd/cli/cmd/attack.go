package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// attackCmd represents the attack command
var attackCmd = &cobra.Command{
	Use:   "attack <attacker> <target>",
	Short: "Attack a unit",
	Long: `Attack a target unit with your unit.
Positions can be unit IDs (like A1) or coordinates (like 3,4).
The <target> position can also be a direction or sequence of directions: L, R, TL, TR, BL, BR.
Multiple directions can be chained with commas to target distant units: TL,TL,TR.

Examples:
  ww attack A1 B2              Attack unit B2 with unit A1
  ww attack A1 TR              Attack top-right neighbor with A1
  ww attack A1 TL,TL           Attack unit 2 steps top-left from A1
  ww attack 3,4 5,6            Attack position 5,6 with unit at 3,4
  ww attack A1 B2 --dryrun     Preview attack outcome without saving`,
	Args: cobra.ExactArgs(2),
	RunE: runAttack,
}

func init() {
	rootCmd.AddCommand(attackCmd)
}

func runAttack(cmd *cobra.Command, args []string) error {
	attackerLabel := args[0]
	targetLabel := args[1]

	ctx := context.Background()
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Attacking from %s to %s\n", attackerLabel, targetLabel)
	}

	// Execute attack directly via ProcessMoves - server parses labels
	resp, err := gc.Service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gc.GameID,
		DryRun: isDryrun(),
		Moves: []*v1.GameMove{{
			Player: gc.State.CurrentPlayer,
			MoveType: &v1.GameMove_AttackUnit{
				AttackUnit: &v1.AttackUnitAction{
					Attacker: &v1.Position{Label: attackerLabel},
					Defender: &v1.Position{Label: targetLabel},
				},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("attack failed: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id":  gc.GameID,
			"action":   "attack",
			"attacker": attackerLabel,
			"target":   targetLabel,
			"dryrun":   isDryrun(),
			"success":  true,
			"changes":  formatChangesForJSON(resp.Moves),
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	if isDryrun() {
		sb.WriteString("Attack (dryrun): Would succeed\n")
	} else {
		sb.WriteString("Attack: Success\n")
	}
	sb.WriteString(fmt.Sprintf("  Attacked from %s to %s\n", attackerLabel, targetLabel))

	// Show changes from response (damage dealt, units killed, etc.)
	if len(resp.Moves) > 0 && len(resp.Moves[0].Changes) > 0 {
		sb.WriteString("  Results:\n")
		for _, change := range resp.Moves[0].Changes {
			sb.WriteString(fmt.Sprintf("    - %s\n", formatChange(change)))
		}
	}

	return formatter.PrintText(sb.String())
}
