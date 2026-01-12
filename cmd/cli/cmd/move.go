package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// moveCmd represents the move command
var moveCmd = &cobra.Command{
	Use:   "move <from> <to>",
	Short: "Move a unit",
	Long: `Move a unit from one position to another.
Positions can be unit IDs (like A1) or coordinates (like 3,4).
The <to> position can also be a direction or sequence of directions: L, R, TL, TR, BL, BR.
Multiple directions can be chained with commas to move multiple steps: TL,TL,TR.

Examples:
  ww move A1 5,6           Move unit A1 to position 5,6
  ww move A1 R             Move unit A1 to the right
  ww move A1 TL,TL,TR      Move unit A1: top-left, then top-left, then top-right
  ww move 3,4 TL           Move unit at 3,4 to top-left
  ww move A1 R --dryrun    Preview move without saving`,
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE:         runMove,
}

func init() {
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	fromLabel := args[0]
	toLabel := args[1]

	ctx := context.Background()
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Moving from %s to %s\n", fromLabel, toLabel)
	}

	// Execute move directly via ProcessMoves - server parses labels
	resp, err := gc.Service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gc.GameID,
		DryRun: isDryrun(),
		Moves: []*v1.GameMove{{
			Player: gc.State.CurrentPlayer,
			MoveType: &v1.GameMove_MoveUnit{
				MoveUnit: &v1.MoveUnitAction{
					From: &v1.Position{Label: fromLabel},
					To:   &v1.Position{Label: toLabel},
				},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("move failed: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gc.GameID,
			"action":  "move",
			"from":    fromLabel,
			"to":      toLabel,
			"dryrun":  isDryrun(),
			"success": true,
			"changes": formatChangesForJSON(resp.Moves),
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	if isDryrun() {
		sb.WriteString("Move (dryrun): Would succeed\n")
	} else {
		sb.WriteString("Move: Success\n")
	}
	sb.WriteString(fmt.Sprintf("  Moved from %s to %s\n", fromLabel, toLabel))

	// Show changes from response
	if len(resp.Moves) > 0 && len(resp.Moves[0].Changes) > 0 {
		sb.WriteString("  Changes:\n")
		for _, change := range resp.Moves[0].Changes {
			sb.WriteString(fmt.Sprintf("    - %s\n", formatChange(change)))
		}
	}

	return formatter.PrintText(sb.String())
}

// formatChangesForJSON converts WorldChanges to a JSON-friendly format
func formatChangesForJSON(moves []*v1.GameMove) []map[string]any {
	var changes []map[string]any
	for _, move := range moves {
		for _, change := range move.Changes {
			changes = append(changes, map[string]any{
				"type":        fmt.Sprintf("%T", change.ChangeType),
				"description": formatChange(change),
			})
		}
	}
	return changes
}

// formatChange formats a WorldChange for display
func formatChange(change *v1.WorldChange) string {
	switch c := change.ChangeType.(type) {
	case *v1.WorldChange_UnitMoved:
		prev := c.UnitMoved.PreviousUnit
		upd := c.UnitMoved.UpdatedUnit
		return fmt.Sprintf("Unit %s moved from (%d,%d) to (%d,%d)", prev.Shortcut, prev.Q, prev.R, upd.Q, upd.R)
	case *v1.WorldChange_UnitDamaged:
		u := c.UnitDamaged.UpdatedUnit
		return fmt.Sprintf("Unit %s damaged (health: %d)", u.Shortcut, u.AvailableHealth)
	case *v1.WorldChange_UnitKilled:
		u := c.UnitKilled.PreviousUnit
		return fmt.Sprintf("Unit %s killed", u.Shortcut)
	case *v1.WorldChange_UnitBuilt:
		u := c.UnitBuilt.Unit
		return fmt.Sprintf("Unit %s built at (%d,%d)", u.Shortcut, u.Q, u.R)
	case *v1.WorldChange_PlayerChanged:
		return fmt.Sprintf("Turn changed to player %d", c.PlayerChanged.NewPlayer)
	case *v1.WorldChange_CoinsChanged:
		return fmt.Sprintf("Player %d coins: %d -> %d", c.CoinsChanged.PlayerId, c.CoinsChanged.PreviousCoins, c.CoinsChanged.NewCoins)
	case *v1.WorldChange_CaptureStarted:
		return fmt.Sprintf("Capture started at (%d,%d)", c.CaptureStarted.TileQ, c.CaptureStarted.TileR)
	case *v1.WorldChange_TileCaptured:
		return fmt.Sprintf("Tile captured at (%d,%d) by player %d", c.TileCaptured.TileQ, c.TileCaptured.TileR, c.TileCaptured.NewOwner)
	case *v1.WorldChange_UnitHealed:
		u := c.UnitHealed.UpdatedUnit
		return fmt.Sprintf("Unit %s healed (+%d health, now %d)", u.Shortcut, c.UnitHealed.HealAmount, u.AvailableHealth)
	default:
		return fmt.Sprintf("%T", change.ChangeType)
	}
}
