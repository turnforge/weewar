package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
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
	fromPos := args[0]
	toPos := args[1]

	ctx := context.Background()
	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		return err
	}

	// Parse from position
	fromTarget, err := services.ParsePositionOrUnit(rtGame, fromPos)
	if err != nil {
		return fmt.Errorf("invalid from position: %w", err)
	}
	fromCoord := fromTarget.GetCoordinate()

	// Parse to position with context (supports directions)
	toTarget, err := services.ParsePositionOrUnitWithContext(rtGame, toPos, &fromCoord)
	if err != nil {
		return fmt.Errorf("invalid to position: %w", err)
	}
	toCoord := toTarget.GetCoordinate()

	if isVerbose() {
		fmt.Printf("[VERBOSE] Moving from %s to %s\n", fromCoord.String(), toCoord.String())
	}

	// Two-click pattern: Click unit to select, then click destination to move
	// Click 1: Select unit on base-map layer
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(fromCoord.Q),
		R:      int32(fromCoord.R),
		Layer:  "base-map",
	})
	if err != nil {
		return fmt.Errorf("failed to select unit: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Unit selected at %s\n", fromCoord.String())
	}

	// Click 2: Click destination on movement-highlight layer to execute move
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(toCoord.Q),
		R:      int32(toCoord.R),
		Layer:  "movement-highlight",
	})
	if err != nil {
		return fmt.Errorf("failed to execute move: %w", err)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Move executed to %s\n", toCoord.String())
	}

	// Save state unless in dryrun mode
	if err := savePresenterState(pc, isDryrun()); err != nil {
		return err
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gameID,
			"action":  "move",
			"from": map[string]int{
				"q": fromCoord.Q,
				"r": fromCoord.R,
			},
			"to": map[string]int{
				"q": toCoord.Q,
				"r": toCoord.R,
			},
			"success": true,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	sb.WriteString("Move: Success\n")
	sb.WriteString(fmt.Sprintf("  Moved unit from %s to %s\n", fromCoord.String(), toCoord.String()))
	sb.WriteString(fmt.Sprintf("\nCurrent player: %d, Turn: %d\n",
		pc.GameState.State.CurrentPlayer, pc.GameState.State.TurnCounter))

	return formatter.PrintText(sb.String())
}
