package lib

import (
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// Proto Unit helper methods
func UnitGetCoord(u *v1.Unit) AxialCoord {
	return AxialCoord{Q: int(u.Q), R: int(u.R)}
}

func UnitSetCoord(u *v1.Unit, coord AxialCoord) {
	u.Q = int32(coord.Q)
	u.R = int32(coord.R)
}

// Proto Tile helper methods
func TileGetCoord(t *v1.Tile) AxialCoord {
	return AxialCoord{Q: int(t.Q), R: int(t.R)}
}

func TileSetCoord(t *v1.Tile, coord AxialCoord) {
	t.Q = int32(coord.Q)
	t.R = int32(coord.R)
}

// Proto factory functions
func NewUnit(unitType, player int, coord AxialCoord) *v1.Unit {
	return &v1.Unit{
		Q:        int32(coord.Q),
		R:        int32(coord.R),
		Player:   int32(player),
		UnitType: int32(unitType),
	}
}

func NewTile(coord AxialCoord, tileType int) *v1.Tile {
	return &v1.Tile{
		Q:        int32(coord.Q),
		R:        int32(coord.R),
		TileType: int32(tileType),
		Player:   0, // Default to neutral
	}
}

// Helper functions to convert between int and int32 for proto fields
func ProtoInt32(val int) int32 {
	return int32(val)
}

func ProtoInt(val int32) int {
	return int(val)
}

// ProtoToRuntimeGame converts protobuf game/state to runtime game
// This is LilBattle-specific and doesn't belong in TurnEngine
func ProtoToRuntimeGame(game *v1.Game, gameState *v1.GameState) *Game {
	// Create the runtime game from the protobuf data
	world := NewWorld(game.Name, gameState.WorldData)

	// Create the runtime game with loaded default rules engine
	rulesEngine := DefaultRulesEngine() // Use loaded default rules engine

	// Use NewGameFromState instead of NewGame to preserve unit stats
	return NewGame(game, gameState, world, rulesEngine, 12345) // Default seed
}

// RuntimeGameToProto returns the proto state from a runtime game
// Since runtime game already embeds the proto objects, we just return the pointer
func RuntimeGameToProto(rtGame *Game) *v1.GameState {
	if rtGame == nil || rtGame.GameState == nil {
		return nil
	}

	// Just return the embedded proto state - it's already being updated by the runtime game
	return rtGame.GameState
}
