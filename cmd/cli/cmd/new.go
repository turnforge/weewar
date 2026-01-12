package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services/connectclient"
)

var (
	gameName          string
	gameDescription   string
	startingCoins     int32
	gameIncome        int32
	landbaseIncome    int32
	navalbaseIncome   int32
	airportbaseIncome int32
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <world_id>",
	Short: "Create a new game from a world",
	Long: `Create a new game from an existing world.
Players are auto-detected from the world based on who owns units or tiles.
Requires LILBATTLE_SERVER to be set.

Examples:
  ww new 01bdc3ce                              Create game from world
  ww new 01bdc3ce --name "My Game"             Create game with custom name
  ww new 01bdc3ce --starting-coins 200         Start with 200 coins per player
  ww new 01bdc3ce --landbase-income 100        Set landbase income to 100`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&gameName, "name", "", "name for the new game")
	newCmd.Flags().StringVar(&gameDescription, "desc", "", "description for the new game")
	newCmd.Flags().Int32Var(&startingCoins, "starting-coins", 0, "starting coins per player")
	newCmd.Flags().Int32Var(&gameIncome, "game-income", 300, "base income per turn")
	newCmd.Flags().Int32Var(&landbaseIncome, "landbase-income", 150, "income per landbase")
	newCmd.Flags().Int32Var(&navalbaseIncome, "navalbase-income", 150, "income per navalbase")
	newCmd.Flags().Int32Var(&airportbaseIncome, "airportbase-income", 150, "income per airport")
}

func runNew(cmd *cobra.Command, args []string) error {
	worldID := args[0]
	ctx := context.Background()

	serverURL := getServerURL()
	if serverURL == "" {
		return fmt.Errorf("LILBATTLE_SERVER is required for creating games (e.g., http://localhost:9080)")
	}

	// Create Connect clients
	gamesClient := connectclient.NewConnectGamesClient(serverURL)
	worldsClient := connectclient.NewConnectWorldsClient(serverURL)

	if isVerbose() {
		fmt.Printf("[VERBOSE] Using server: %s\n", serverURL)
		fmt.Printf("[VERBOSE] Loading world: %s\n", worldID)
	}

	// Load world to auto-detect players
	worldResp, err := worldsClient.GetWorld(ctx, &v1.GetWorldRequest{Id: worldID})
	if err != nil {
		return fmt.Errorf("failed to load world %s: %w", worldID, err)
	}

	if worldResp.WorldData == nil {
		return fmt.Errorf("world %s has no data", worldID)
	}

	// Auto-detect players from world data
	players := detectPlayersFromWorld(worldResp.WorldData)
	if len(players) == 0 {
		return fmt.Errorf("no players found in world (no units or tiles with player ownership)")
	}

	if isVerbose() {
		fmt.Printf("[VERBOSE] Detected %d player(s): ", len(players))
		for i, p := range players {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("Player %d", p.PlayerId)
		}
		fmt.Println()
	}

	// Build game name from world if not provided
	name := gameName
	if name == "" {
		if worldResp.World != nil && worldResp.World.Name != "" {
			name = worldResp.World.Name + " Game"
		} else {
			name = "Game from " + worldID
		}
	}

	// Create game request
	game := &v1.Game{
		WorldId:     worldID,
		Name:        name,
		Description: gameDescription,
		Config: &v1.GameConfiguration{
			Players: players,
			IncomeConfigs: &v1.IncomeConfig{
				StartingCoins:     startingCoins,
				GameIncome:        gameIncome,
				LandbaseIncome:    landbaseIncome,
				NavalbaseIncome:   navalbaseIncome,
				AirportbaseIncome: airportbaseIncome,
			},
		},
	}

	// Create the game
	resp, err := gamesClient.CreateGame(ctx, &v1.CreateGameRequest{Game: game})
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	// Format output
	formatter := NewOutputFormatter()

	if formatter.JSON {
		data := map[string]any{
			"game_id":     resp.Game.Id,
			"name":        resp.Game.Name,
			"world_id":    resp.Game.WorldId,
			"players":     len(players),
			"description": resp.Game.Description,
		}
		return formatter.PrintJSON(data)
	}

	// Text output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Created game: %s\n", resp.Game.Id))
	sb.WriteString(fmt.Sprintf("  Name: %s\n", resp.Game.Name))
	sb.WriteString(fmt.Sprintf("  World: %s\n", resp.Game.WorldId))
	sb.WriteString(fmt.Sprintf("  Players: %d\n", len(players)))
	if resp.Game.Description != "" {
		sb.WriteString(fmt.Sprintf("  Description: %s\n", resp.Game.Description))
	}
	sb.WriteString(fmt.Sprintf("\nTo play: export LILBATTLE_GAME_ID=%s\n", resp.Game.Id))

	return formatter.PrintText(sb.String())
}

// detectPlayersFromWorld scans world data and returns GamePlayer entries
// for each player that owns at least one unit or tile
func detectPlayersFromWorld(worldData *v1.WorldData) []*v1.GamePlayer {
	// Ensure maps are initialized
	lib.MigrateWorldData(worldData)

	playerSet := make(map[int32]bool)

	// Check tiles
	for _, tile := range worldData.TilesMap {
		if tile != nil && tile.Player > 0 {
			playerSet[tile.Player] = true
		}
	}

	// Check units
	for _, unit := range worldData.UnitsMap {
		if unit != nil && unit.Player > 0 {
			playerSet[unit.Player] = true
		}
	}

	// Build sorted player list
	var players []*v1.GamePlayer
	for playerID := int32(1); playerID <= 8; playerID++ {
		if playerSet[playerID] {
			players = append(players, &v1.GamePlayer{
				PlayerId:   playerID,
				PlayerType: "human",
				IsActive:   true,
			})
		}
	}

	return players
}
