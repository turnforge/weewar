package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current game status",
	Long: `Display the current game state including turn number, current player,
and game status.

Examples:
  ww status
  ww status --json`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	gc, err := GetGameContext()
	if err != nil {
		return err
	}

	if gc.State == nil {
		return fmt.Errorf("game state not initialized")
	}
	if gc.Game == nil {
		return fmt.Errorf("game metadata not initialized")
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		// Build player info for JSON output
		players := []map[string]any{}

		// Count units per player
		unitCounts := make(map[int32]int)
		if gc.State.WorldData != nil {
			for _, unit := range gc.State.WorldData.UnitsMap {
				if unit != nil {
					unitCounts[unit.Player]++
				}
			}
		}

		// Count tiles per player
		tileCounts := make(map[int32]int)
		if gc.State.WorldData != nil {
			for _, tile := range gc.State.WorldData.TilesMap {
				if tile != nil && tile.Player > 0 {
					tileCounts[tile.Player]++
				}
			}
		}

		if gc.Game.Config != nil {
			for _, player := range gc.Game.Config.Players {
				// Get coins from GameState.PlayerStates
				coins := int32(0)
				if playerState := gc.State.PlayerStates[player.PlayerId]; playerState != nil {
					coins = playerState.Coins
				}
				players = append(players, map[string]any{
					"player_id":   player.PlayerId,
					"player_type": player.PlayerType,
					"name":        player.Name,
					"coins":       coins,
					"units":       unitCounts[player.PlayerId],
					"tiles":       tileCounts[player.PlayerId],
					"team_id":     player.TeamId,
					"is_active":   player.IsActive,
				})
			}
		}

		// JSON output
		data := map[string]any{
			"game_id":        gc.GameID,
			"game_name":      gc.Game.Name,
			"description":    gc.Game.Description,
			"turn":           gc.State.TurnCounter,
			"current_player": gc.State.CurrentPlayer,
			"status":         gc.State.Status.String(),
			"winning_player": gc.State.WinningPlayer,
			"players":        players,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	text := FormatGameStatus(gc.Game, gc.State)
	return formatter.PrintText(text)
}
