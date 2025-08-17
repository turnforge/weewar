package services

import (
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
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
