package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/panyam/turnengine/engine/storage"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

func CreateTestWorld(name string, nq, nr int, units []*v1.Unit) *World {
	// 1. Create test world with 3 units
	world := NewWorld("test", nil)
	// Add some tiles for movement
	for q := range nq {
		for r := range nr {
			coord := AxialCoord{Q: q, R: r}
			tile := NewTile(coord, 1) // Grass terrain
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
func LoadTestWorldFromStorage(worldsStorageDir, worldId string) (*World, *v1.GameState, error) {
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
	rtGame := ProtoToRuntimeGame(dummyGame, gameState)

	// Extract the world from the runtime game
	rtWorld := rtGame.World

	return rtWorld, gameState, nil
}

// CreateTestUnit creates a test unit with given parameters
func CreateTestUnit(q, r int, player, unitType int) *v1.Unit {
	return &v1.Unit{
		Q:               int32(q),
		R:               int32(r),
		Player:          int32(player),
		UnitType:        int32(unitType),
		AvailableHealth: 100,
		DistanceLeft:    3,
	}
}

// LoadTestWorld loads a real world from the weewar data directory
// This allows tests to use actual world data created in the editor
func LoadTestWorld(worldId string) (*World, error) {
	// Default to user's dev-app-data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	// Try both possible locations for world data
	worldsDir := filepath.Join(homeDir, "dev-app-data", "weewar", "storage", "worlds")
	worldFile := filepath.Join(worldsDir, worldId, "world.json")
	worldDataFile := filepath.Join(worldsDir, worldId, "worlddata.json")

	// Read world.json
	worldBytes, err := os.ReadFile(worldFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read world file: %w", err)
	}

	var protoWorld v1.World
	if err := json.Unmarshal(worldBytes, &protoWorld); err != nil {
		return nil, fmt.Errorf("failed to unmarshal world: %w", err)
	}

	// Read worlddata.json
	worldDataBytes, err := os.ReadFile(worldDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read worlddata file: %w", err)
	}

	var protoWorldData v1.WorldData
	if err := json.Unmarshal(worldDataBytes, &protoWorldData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worlddata: %w", err)
	}

	// Create runtime world from proto data
	world := NewWorld(protoWorld.Name, &protoWorldData)

	return world, nil
}
