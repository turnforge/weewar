package weewar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

// CreateTestUnit creates a test unit with given parameters
func CreateTestUnit(q, r int, player, unitType int) *v1.Unit {
	return &v1.Unit{
		Q:               int32(q),
		R:               int32(r),
		Player:          int32(player),
		UnitType:        int32(unitType),
		AvailableHealth: 100,
		DistanceLeft:    3,
		TurnCounter:     1,
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
