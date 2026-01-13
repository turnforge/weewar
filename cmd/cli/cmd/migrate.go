package cmd

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/services/connectclient"

	"github.com/spf13/cobra"
)

var (
	sourceToken string
	destToken   string
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate <source-url> <dest-url>",
	Short: "Migrate a world from one server to another",
	Long: `Migrate a world from one LilBattle server to another.

The URLs should be full API URLs for worlds, e.g.:
  http://localhost:6060/api/v1/worlds/Desert

The world ID is extracted from the URL. You can specify a different
destination world ID by using a different ID in the destination URL.

Authentication:
  Uses stored credentials from 'ww login'. You can also provide tokens
  directly via --source-token and --dest-token flags.

Examples:
  # Migrate Desert from local dev to another local server
  ww migrate http://localhost:6060/api/v1/worlds/Desert \
             http://localhost:8080/api/v1/worlds/Desert

  # Migrate and rename the world
  ww migrate http://localhost:6060/api/v1/worlds/Desert \
             http://localhost:8080/api/v1/worlds/DesertCopy

  # Migrate with explicit tokens
  ww migrate http://localhost:6060/api/v1/worlds/Desert \
             https://prod.example.com/api/v1/worlds/Desert \
             --dest-token $PROD_TOKEN`,
	Args: cobra.ExactArgs(2),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVar(&sourceToken, "source-token", "", "Auth token for source server (overrides stored credentials)")
	migrateCmd.Flags().StringVar(&destToken, "dest-token", "", "Auth token for destination server (overrides stored credentials)")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	sourceURL := args[0]
	destURL := args[1]

	ctx := context.Background()
	formatter := NewOutputFormatter()

	// Extract server bases and world IDs
	sourceBase, err := extractServerBase(sourceURL)
	if err != nil {
		return fmt.Errorf("invalid source URL: %w", err)
	}

	destBase, err := extractServerBase(destURL)
	if err != nil {
		return fmt.Errorf("invalid destination URL: %w", err)
	}

	sourceWorldID, err := extractWorldID(sourceURL)
	if err != nil {
		return fmt.Errorf("invalid source URL: %w", err)
	}

	destWorldID, err := extractWorldID(destURL)
	if err != nil {
		return fmt.Errorf("invalid destination URL: %w", err)
	}

	// Get tokens for both servers
	srcToken := sourceToken
	if srcToken == "" {
		srcToken = GetTokenForServer(sourceBase)
	}

	dstToken := destToken
	if dstToken == "" {
		dstToken = GetTokenForServer(destBase)
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Source: %s (world: %s, auth: %v)\n", sourceBase, sourceWorldID, srcToken != "")
		fmt.Printf("[VERBOSE] Dest: %s (world: %s, auth: %v)\n", destBase, destWorldID, dstToken != "")
	}

	// Create clients
	sourceClient := connectclient.NewConnectWorldsClientWithAuth(sourceBase, srcToken)
	destClient := connectclient.NewConnectWorldsClientWithAuth(destBase, dstToken)

	// Fetch world from source
	if !formatter.JSON {
		fmt.Printf("Fetching world '%s' from %s...\n", sourceWorldID, sourceBase)
	}

	getResp, err := sourceClient.GetWorld(ctx, &v1.GetWorldRequest{
		Id: sourceWorldID,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch source world: %w", err)
	}

	if getResp.World == nil {
		return fmt.Errorf("source world not found: %s", sourceWorldID)
	}

	// Show what we're migrating
	world := getResp.World
	worldData := getResp.WorldData

	tileCount := 0
	unitCount := 0
	if worldData != nil {
		tileCount = len(worldData.TilesMap)
		unitCount = len(worldData.UnitsMap)
	}

	if !formatter.JSON {
		fmt.Printf("World: %s\n", world.Name)
		fmt.Printf("  Description: %s\n", world.Description)
		fmt.Printf("  Tiles: %d, Units: %d\n", tileCount, unitCount)
		fmt.Printf("Migrating to %s as '%s'...\n", destBase, destWorldID)
	}

	// Prepare the world for destination (update ID)
	world.Id = destWorldID

	// Try to create the world first
	createReq := &v1.CreateWorldRequest{
		World:     world,
		WorldData: worldData,
	}

	_, err = destClient.CreateWorld(ctx, createReq)
	if err != nil {
		// Check if it's an "already exists" error
		errStr := err.Error()
		if containsIgnoreCase(errStr, "already exists") || containsIgnoreCase(errStr, "AlreadyExists") {
			if !formatter.JSON {
				fmt.Println("World already exists, updating instead...")
			}

			// Update instead
			updateReq := &v1.UpdateWorldRequest{
				World:     world,
				WorldData: worldData,
			}

			_, err = destClient.UpdateWorld(ctx, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update destination world: %w", err)
			}

			if formatter.JSON {
				return formatter.PrintJSON(map[string]any{
					"source_server": sourceBase,
					"source_world":  sourceWorldID,
					"dest_server":   destBase,
					"dest_world":    destWorldID,
					"action":        "updated",
					"tiles":         tileCount,
					"units":         unitCount,
				})
			}

			fmt.Println("World updated successfully!")
		} else {
			return fmt.Errorf("failed to create destination world: %w", err)
		}
	} else {
		if formatter.JSON {
			return formatter.PrintJSON(map[string]any{
				"source_server": sourceBase,
				"source_world":  sourceWorldID,
				"dest_server":   destBase,
				"dest_world":    destWorldID,
				"action":        "created",
				"tiles":         tileCount,
				"units":         unitCount,
			})
		}

		fmt.Println("World created successfully!")
	}

	if !formatter.JSON {
		fmt.Println("Migration complete!")
	}

	return nil
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
