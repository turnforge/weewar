package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// endturnCmd represents the endturn command
var endturnCmd = &cobra.Command{
	Use:   "endturn",
	Short: "End the current player's turn",
	Long: `End the current player's turn and advance to the next player.
All units for the new player will be reset with full movement points.

Examples:
  ww endturn
  ww endturn --dryrun    Preview turn transition without saving`,
	RunE: runEndTurn,
}

func init() {
	rootCmd.AddCommand(endturnCmd)
}

func runEndTurn(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	// Get current state before ending turn
	previousPlayer := gc.State.CurrentPlayer
	previousTurn := gc.State.TurnCounter

	if isVerbose() {
		fmt.Printf("[VERBOSE] Ending turn for player %d (turn %d)\n", previousPlayer, previousTurn)
	}

	// Execute end turn directly via ProcessMoves
	resp, err := gc.Service.ProcessMoves(ctx, &v1.ProcessMovesRequest{
		GameId: gc.GameID,
		DryRun: isDryrun(),
		Moves: []*v1.GameMove{{
			Player: gc.State.CurrentPlayer,
			MoveType: &v1.GameMove_EndTurn{
				EndTurn: &v1.EndTurnAction{},
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("end turn failed: %w", err)
	}

	// Extract new player from changes
	var newPlayer int32 = previousPlayer
	var newTurn int32 = previousTurn
	if len(resp.Moves) > 0 {
		for _, change := range resp.Moves[0].Changes {
			if pc, ok := change.ChangeType.(*v1.WorldChange_PlayerChanged); ok {
				newPlayer = pc.PlayerChanged.NewPlayer
				newTurn = pc.PlayerChanged.NewTurn
				break
			}
		}
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id":         gc.GameID,
			"action":          "endturn",
			"previous_player": previousPlayer,
			"previous_turn":   previousTurn,
			"current_player":  newPlayer,
			"current_turn":    newTurn,
			"dryrun":          isDryrun(),
			"success":         true,
			"changes":         formatChangesForJSON(resp.Moves),
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	if isDryrun() {
		sb.WriteString("End Turn (dryrun): Would succeed\n")
	} else {
		sb.WriteString("End Turn: Success\n")
	}
	sb.WriteString(fmt.Sprintf("  Turn ended for player %d\n", previousPlayer))
	sb.WriteString(fmt.Sprintf("  Now player %d's turn (turn %d)\n", newPlayer, newTurn))

	return formatter.PrintText(sb.String())
}
