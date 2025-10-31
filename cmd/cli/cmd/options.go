package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
)

// optionsCmd represents the options command
var optionsCmd = &cobra.Command{
	Use:   "options <unit>",
	Short: "Show available options for a unit",
	Long: `Display available actions for a unit at the specified position.
The position can be a unit ID (like A1, B2) or coordinates (like 3,4).

Examples:
  ww options A1        Show options for unit A1
  ww options 3,4       Show options for position 3,4
  ww options A1 --json Show options as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runOptions,
}

func init() {
	rootCmd.AddCommand(optionsCmd)
}

func runOptions(cmd *cobra.Command, args []string) error {
	position := args[0]

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

	ctx := context.Background()

	// Get runtime game for parsing position
	rtGame, err := pc.Presenter.GamesService.GetRuntimeGame(
		pc.Presenter.GamesService.SingletonGame,
		pc.Presenter.GamesService.SingletonGameState)
	if err != nil {
		return fmt.Errorf("failed to get runtime game: %w", err)
	}

	// Parse position (could be coordinate or unit ID)
	target, err := services.ParsePositionOrUnit(rtGame, position)
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
					rtGame, err := pc.Presenter.GamesService.GetRuntimeGame(
						pc.Presenter.GamesService.SingletonGame,
						pc.Presenter.GamesService.SingletonGameState)
					if err == nil && rtGame.GetRulesEngine() != nil {
						if unitDef, err := rtGame.GetRulesEngine().GetUnitData(buildOpt.UnitType); err == nil {
							unitName = unitDef.Name
						}
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
