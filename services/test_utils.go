package services

import (
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"github.com/panyam/turnengine/engine/storage"
)

func CreateTestWorld(name string, nq, nr int, units []*v1.Unit) *weewar.World {
	// 1. Create test world with 3 units
	world := weewar.NewWorld("test")
	// Add some tiles for movement
	for q := range nq {
		for r := range nr {
			coord := weewar.AxialCoord{Q: q, R: r}
			tile := weewar.NewTile(coord, 1) // Grass terrain
			world.AddTile(tile)
		}
	}

	for _, unit := range units {
		world.AddUnit(unit)
	}
	return world
}

// LoadTestWorldFromStorage loads world data from storage directory using FSWorldsServiceImpl
// This allows using real worlds created in the world editor UI
func LoadTestWorldFromStorage(worldsStorageDir, worldId string) (*weewar.World, *v1.GameState, error) {
	// Create FSWorldsService to load real world data
	worldsService := &FSWorldsServiceImpl{
		storage: storage.NewFileStorage(worldsStorageDir),
	}
	
	// Load the world using GetWorld RPC (same as production code)
	worldResp, err := worldsService.GetWorld(context.Background(), &v1.GetWorldRequest{
		Id: worldId,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load world %s from %s: %w", worldId, worldsStorageDir, err)
	}
	
	// Create basic game state using the loaded world data
	gameState := &v1.GameState{
		CurrentPlayer: 1, // Default to player 1
		TurnCounter:   1, // Default to turn 1
		WorldData:     worldResp.WorldData,
	}
	
	// Create a dummy game from the world data (ProtoToRuntimeGame expects *v1.Game)
	dummyGame := &v1.Game{
		Id:          "test-game-" + worldId,
		Name:        worldResp.World.Name,
		Description: worldResp.World.Description,
		WorldId:     worldId,
	}
	
	// Convert protobuf world data to runtime game 
	rtGame, err := ProtoToRuntimeGame(dummyGame, gameState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert world to runtime game: %w", err)
	}
	
	// Extract the world from the runtime game
	rtWorld := rtGame.World
	
	return rtWorld, gameState, nil
}
