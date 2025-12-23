package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// tilesCmd represents the tiles command
var tilesCmd = &cobra.Command{
	Use:   "tiles",
	Short: "List all player-owned tiles in the game",
	Long: `Display all player-owned tiles grouped by player, showing their position
and terrain type.

Examples:
  ww tiles
  ww tiles --json`,
	RunE: runTiles,
}

func init() {
	rootCmd.AddCommand(tilesCmd)
}

func runTiles(cmd *cobra.Command, args []string) error {
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if gc.State == nil {
		return fmt.Errorf("game state not initialized")
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		// JSON output
		rulesEngine := gc.RTGame.GetRulesEngine()
		tiles := []map[string]any{}
		if gc.State.WorldData != nil {
			for _, tile := range gc.State.WorldData.TilesMap {
				if tile != nil {
					tileName := ""
					if terrainDef, err := rulesEngine.GetTerrainData(tile.TileType); err == nil {
						tileName = terrainDef.Name
					}
					tiles = append(tiles, map[string]any{
						"player":    tile.Player,
						"shortcut":  tile.Shortcut,
						"q":         tile.Q,
						"r":         tile.R,
						"tile_type": tile.TileType,
						"tile_name": tileName,
					})
				}
			}
		}

		data := map[string]any{
			"game_id": gc.GameID,
			"tiles":   tiles,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	text := FormatTilesWithContext(gc)
	return formatter.PrintText(text)
}
