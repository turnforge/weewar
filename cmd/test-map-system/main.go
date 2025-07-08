package main

import (
	"fmt"
	"log"

	"github.com/panyam/turnengine/games/weewar"
)

func main() {
	fmt.Println("WeeWar Map System Test")
	fmt.Println("=====================")

	// Test creating map system
	mapSystem, err := weewar.NewWeeWarMapSystem()
	if err != nil {
		log.Fatalf("Failed to create map system: %v", err)
	}

	// Display map statistics
	stats := mapSystem.GetMapStatistics()
	fmt.Printf("Map Statistics:\n")
	fmt.Printf("  Total Maps: %d\n", stats["totalMaps"])
	fmt.Printf("  Total Games Played: %d\n", stats["totalGamesPlayed"])
	
	if mostPlayed, ok := stats["mostPlayedMap"].(map[string]interface{}); ok {
		fmt.Printf("  Most Played Map: %s (%d games)\n", 
			mostPlayed["name"], mostPlayed["gamesPlayed"])
	}

	// Test getting a specific map
	mapData, err := mapSystem.GetMapByName("Small World")
	if err != nil {
		log.Fatalf("Failed to get Small World map: %v", err)
	}

	fmt.Printf("\nMap Details: %s\n", mapData.Name)
	fmt.Printf("  ID: %d\n", mapData.ID)
	fmt.Printf("  Players: %d\n", mapData.Players)
	fmt.Printf("  Size: %s\n", mapData.Size)
	fmt.Printf("  Creator: %s\n", mapData.Creator)
	fmt.Printf("  Games Played: %d\n", mapData.GamesPlayed)
	fmt.Printf("  Tiles: %d total\n", mapData.TileCount)
	
	// Show tile breakdown
	fmt.Printf("  Tile breakdown:\n")
	for tileType, count := range mapData.Tiles {
		fmt.Printf("    %s: %d\n", tileType, count)
	}
	
	// Show initial units
	fmt.Printf("  Initial units:\n")
	for unitType, count := range mapData.InitialUnits {
		fmt.Printf("    %s: %d\n", unitType, count)
	}

	// Test creating a game from map
	fmt.Printf("\nCreating game from map...\n")
	config, err := mapSystem.CreateGameConfigFromMap(mapData)
	if err != nil {
		log.Fatalf("Failed to create game config: %v", err)
	}

	fmt.Printf("Game Config:\n")
	fmt.Printf("  Board Size: %dx%d\n", config.BoardWidth, config.BoardHeight)
	fmt.Printf("  Players: %d\n", len(config.Players))
	
	// Show starting units per player
	for playerID, units := range config.StartingUnits {
		fmt.Printf("  %s units: %v\n", playerID, units)
	}

	// Test creating actual game
	fmt.Printf("\nCreating WeeWar game...\n")
	game, err := weewar.CreateWeeWarGameWithMapName("Small World")
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	fmt.Printf("Successfully created WeeWar game with Small World map!\n")
	fmt.Printf("Game State: %s\n", game.GetGameState().GameType)
	fmt.Printf("Board created successfully\n")

	fmt.Println("\nMap system test completed successfully!")
}