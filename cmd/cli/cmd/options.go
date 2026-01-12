package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
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
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Getting options for %s\n", position)
	}

	// Call GetOptionsAt directly with label
	opts, err := gc.Service.GetOptionsAt(ctx, &v1.GetOptionsAtRequest{
		GameId: gc.GameID,
		Pos:    &v1.Position{Label: position},
	})
	if err != nil {
		return fmt.Errorf("failed to get options: %w", err)
	}

	// Try to get unit at position for display purposes
	var unit *v1.Unit
	target, err := lib.ParsePositionOrUnit(gc.RTGame, position)
	if err == nil {
		coord := target.GetCoordinate()
		unit = gc.RTGame.World.UnitAt(coord)
		if unit != nil {
			gc.RTGame.TopUpUnitIfNeeded(unit)
		}
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		// JSON output
		options := []map[string]any{}
		for _, option := range opts.Options {
			switch opt := option.OptionType.(type) {
			case *v1.GameOption_Move:
				options = append(options, map[string]any{
					"type":          "move",
					"q":             opt.Move.To.Q,
					"r":             opt.Move.To.R,
					"movement_cost": opt.Move.MovementCost,
				})
			case *v1.GameOption_Attack:
				options = append(options, map[string]any{
					"type":            "attack",
					"q":               opt.Attack.Defender.Q,
					"r":               opt.Attack.Defender.R,
					"damage_estimate": opt.Attack.DamageEstimate,
				})
			case *v1.GameOption_Build:
				buildOpt := opt.Build
				unitName := fmt.Sprintf("type %d", buildOpt.UnitType)

				// Try to get actual unit name
				if unitDef, err := gc.RTGame.GetRulesEngine().GetUnitData(buildOpt.UnitType); err == nil {
					unitName = unitDef.Name
				}

				options = append(options, map[string]any{
					"type":      "build",
					"unit_type": buildOpt.UnitType,
					"unit_name": unitName,
					"cost":      buildOpt.Cost,
				})
			case *v1.GameOption_Capture:
				captureOpt := opt.Capture
				terrainName := fmt.Sprintf("type %d", captureOpt.TileType)

				// Try to get actual terrain name
				if terrainDef, err := gc.RTGame.GetRulesEngine().GetTerrainData(captureOpt.TileType); err == nil {
					terrainName = terrainDef.Name
				}

				options = append(options, map[string]any{
					"type":         "capture",
					"q":            captureOpt.Pos.Q,
					"r":            captureOpt.Pos.R,
					"tile_type":    captureOpt.TileType,
					"terrain_name": terrainName,
				})
			case *v1.GameOption_EndTurn:
				options = append(options, map[string]any{
					"type": "endturn",
				})
			}
		}

		data := map[string]any{
			"game_id":  gc.GameID,
			"position": position,
			"options":  options,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	text := FormatOptionsResponse(gc, position, opts, unit)
	return formatter.PrintText(text)
}
