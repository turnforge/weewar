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
	Use:   "migrate <source> <dest>",
	Short: "Migrate a world from one server to another",
	Long: `Migrate a world from one LilBattle server to another.

Source and destination can be specified as either:
  - Profile shorthand: profile:worldId (e.g., fsbe:01bdc3ce, prod:arube)
  - Full URL: http://localhost:8080/api/v1/worlds/Desert

Authentication:
  Uses stored credentials from profiles. You can also provide tokens
  directly via --source-token and --dest-token flags.

Examples:
  # Migrate using profile shorthand (recommended)
  ww migrate fsbe:01bdc3ce prod:arube

  # Migrate and rename the world
  ww migrate fsbe:Desert prod:DesertCopy

  # Migrate using full URLs
  ww migrate http://localhost:8080/api/v1/worlds/Desert \
             https://prod.example.com/api/v1/worlds/Desert

  # Migrate with explicit tokens
  ww migrate fsbe:Desert prod:Desert --dest-token $PROD_TOKEN`,
	Args: cobra.ExactArgs(2),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVar(&sourceToken, "source-token", "", "Auth token for source server (overrides stored credentials)")
	migrateCmd.Flags().StringVar(&destToken, "dest-token", "", "Auth token for destination server (overrides stored credentials)")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	sourceSpec := args[0]
	destSpec := args[1]

	ctx := context.Background()
	formatter := NewOutputFormatter()

	// Parse source and destination specs (profile:worldId or full URL)
	source, err := parseWorldSpec(sourceSpec)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	dest, err := parseWorldSpec(destSpec)
	if err != nil {
		return fmt.Errorf("invalid destination: %w", err)
	}

	// Override tokens if provided via flags
	srcToken := sourceToken
	if srcToken == "" {
		srcToken = source.Token
	}

	dstToken := destToken
	if dstToken == "" {
		dstToken = dest.Token
	}

	if isVerbose() {
		srcProfile := ""
		if source.ProfileName != "" {
			srcProfile = fmt.Sprintf(" [profile: %s]", source.ProfileName)
		}
		destProfile := ""
		if dest.ProfileName != "" {
			destProfile = fmt.Sprintf(" [profile: %s]", dest.ProfileName)
		}
		fmt.Printf("[VERBOSE] Source: %s (world: %s, auth: %v, token_len: %d)%s\n", source.Host, source.WorldID, srcToken != "", len(srcToken), srcProfile)
		fmt.Printf("[VERBOSE] Dest: %s (world: %s, auth: %v, token_len: %d)%s\n", dest.Host, dest.WorldID, dstToken != "", len(dstToken), destProfile)
		fmt.Printf("[VERBOSE] Source API endpoint: %s\n", source.APIEndpoint())
		fmt.Printf("[VERBOSE] Dest API endpoint: %s\n", dest.APIEndpoint())
	}

	// Create clients
	sourceClient := connectclient.NewConnectWorldsClientWithAuth(source.APIEndpoint(), srcToken)
	destClient := connectclient.NewConnectWorldsClientWithAuth(dest.APIEndpoint(), dstToken)

	// Fetch world from source
	if !formatter.JSON {
		fmt.Printf("Fetching world '%s' from %s...\n", source.WorldID, source.Host)
	}

	getResp, err := sourceClient.GetWorld(ctx, &v1.GetWorldRequest{
		Id: source.WorldID,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch source world: %w", err)
	}

	if getResp.World == nil {
		return fmt.Errorf("source world not found: %s", source.WorldID)
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
		fmt.Printf("Migrating to %s as '%s'...\n", dest.Host, dest.WorldID)
	}

	// Prepare the world for destination (update ID)
	world.Id = dest.WorldID

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
					"source_server": source.Host,
					"source_world":  source.WorldID,
					"dest_server":   dest.Host,
					"dest_world":    dest.WorldID,
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
				"source_server": source.Host,
				"source_world":  source.WorldID,
				"dest_server":   dest.Host,
				"dest_world":    dest.WorldID,
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
