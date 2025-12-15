package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
)

// optionsCmd represents the options command
var optionsCmd = &cobra.Command{
	Use:   "options <position>",
	Short: "Show available options for a unit or tile",
	Long: `Display available actions for a unit or tile at the specified position.

For units: Shows movement, attack, and other unit actions.
For tiles: Shows build options (if the tile is a base/harbor owned by current player).

Position formats:
  A1, B2      Unit shortcut (PlayerLetter + UnitNumber)
  3,4         Q,R coordinates
  r4,5        Row,Col coordinates
  t:A1        Force tile lookup (use for build options)
  t:3,4       Force tile lookup by coordinates

Examples:
  ww options A1        Show options for unit A1
  ww options 3,4       Show options for position 3,4
  ww options t:A1      Show build options for tile A1
  ww options A1 --json Show options as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runOptions,
}

func init() {
	rootCmd.AddCommand(optionsCmd)
}

func runOptions(cmd *cobra.Command, args []string) error {
	position := args[0]

	ctx := context.Background()
	pc, _, _, _, rtGame, err := GetGame()
	if err != nil {
		return err
	}

	// Parse position (could be coordinate or unit ID)
	target, err := lib.ParsePositionOrUnit(rtGame, position)
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	coord := target.GetCoordinate()

	// Simulate click on base-map layer to show options
	_, err = pc.Presenter.SceneClicked(ctx, &v1.SceneClickedRequest{
		GameId: gameID,
		Q:      int32(coord.Q),
		R:      int32(coord.R),
		Layer:  "base-map",
	})
	if err != nil {
		return fmt.Errorf("failed to get options: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		// JSON output
		options := []map[string]any{}
		if pc.TurnOptions.Options != nil {
			for _, option := range pc.TurnOptions.Options.Options {
				switch opt := option.OptionType.(type) {
				case *v1.GameOption_Move:
					options = append(options, map[string]any{
						"type":          "move",
						"q":             opt.Move.ToQ,
						"r":             opt.Move.ToR,
						"movement_cost": opt.Move.MovementCost,
					})
				case *v1.GameOption_Attack:
					options = append(options, map[string]any{
						"type":            "attack",
						"q":               opt.Attack.DefenderQ,
						"r":               opt.Attack.DefenderR,
						"damage_estimate": opt.Attack.DamageEstimate,
					})
				case *v1.GameOption_Build:
					buildOpt := opt.Build
					unitName := fmt.Sprintf("type %d", buildOpt.UnitType)

					// Try to get actual unit name
					if unitDef, err := rtGame.GetRulesEngine().GetUnitData(buildOpt.UnitType); err == nil {
						unitName = unitDef.Name
					}

					options = append(options, map[string]any{
						"type":      "build",
						"unit_type": buildOpt.UnitType,
						"unit_name": unitName,
						"cost":      buildOpt.Cost,
					})
				case *v1.GameOption_EndTurn:
					options = append(options, map[string]any{
						"type": "endturn",
					})
				}
			}
		}

		data := map[string]any{
			"game_id":  gameID,
			"position": position,
			"q":        coord.Q,
			"r":        coord.R,
			"options":  options,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	text := FormatOptions(pc, position)
	return formatter.PrintText(text)
}
