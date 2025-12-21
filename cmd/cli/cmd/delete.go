package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/services/connectclient"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <game_id>",
	Short: "Delete a game",
	Long: `Delete an existing game by its ID.
Requires WEEWAR_SERVER to be set.

Examples:
  ww delete abc123                    Delete game abc123
  ww delete abc123 --confirm=false    Delete without confirmation prompt`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	gameID := args[0]
	ctx := context.Background()

	serverURL := getServerURL()
	if serverURL == "" {
		return fmt.Errorf("WEEWAR_SERVER is required for deleting games (e.g., http://localhost:9080)")
	}

	// Confirmation prompt (unless disabled with --confirm=false)
	if shouldConfirm() {
		fmt.Printf("Are you sure you want to delete game %s? (y/n): ", gameID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return fmt.Errorf("delete cancelled")
		}
	}

	// Create Connect client
	gamesClient := connectclient.NewConnectGamesClient(serverURL)

	if isVerbose() {
		fmt.Printf("[VERBOSE] Using server: %s\n", serverURL)
		fmt.Printf("[VERBOSE] Deleting game: %s\n", gameID)
	}

	// Delete the game
	_, err := gamesClient.DeleteGame(ctx, &v1.DeleteGameRequest{Id: gameID})
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id": gameID,
			"deleted": true,
		}
		return formatter.PrintJSON(data)
	}

	return formatter.PrintText(fmt.Sprintf("Deleted game: %s\n", gameID))
}
