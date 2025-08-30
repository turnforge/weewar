package services

import (
	"fmt"
	
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

// ProtoToRuntimeGame converts protobuf game/state to runtime game
// This is WeeWar-specific and doesn't belong in TurnEngine
func ProtoToRuntimeGame(game *v1.Game, gameState *v1.GameState) (*weewar.Game, error) {
	// Create the runtime game from the protobuf data
	world := weewar.NewWorld(game.Name)
	
	// Convert protobuf tiles to runtime tiles
	if gameState.WorldData != nil {
		for _, protoTile := range gameState.WorldData.Tiles {
			coord := weewar.AxialCoord{Q: int(protoTile.Q), R: int(protoTile.R)}
			world.SetTileType(coord, int(protoTile.TileType))
		}
		
		// Convert protobuf units to runtime units
		for _, protoUnit := range gameState.WorldData.Units {
			coord := weewar.AxialCoord{Q: int(protoUnit.Q), R: int(protoUnit.R)}
			fmt.Printf("ProtoToRuntimeGame: Converting unit at (%d, %d), saved DistanceLeft=%d, AvailableHealth=%d\n",
				coord.Q, coord.R, protoUnit.DistanceLeft, protoUnit.AvailableHealth)
			unit := &v1.Unit{
				UnitType:        protoUnit.UnitType,
				Q:               int32(coord.Q),
				R:               int32(coord.R),
				Player:          protoUnit.Player,
				AvailableHealth: protoUnit.AvailableHealth,
				DistanceLeft:    0, // Will be initialized by initializeStartingUnits()
				TurnCounter:     protoUnit.TurnCounter,
			}
			world.AddUnit(unit)
		}
	}
	
	// Create the runtime game with loaded default rules engine
	rulesEngine := weewar.DefaultRulesEngine()            // Use loaded default rules engine
	out, err := weewar.NewGame(world, rulesEngine, 12345) // Default seed
	if err != nil {
		return nil, err
	}
	
	// Debug: Check unit movement points after NewGame initialization
	if out.World != nil {
		for playerId := 1; playerId <= int(out.World.PlayerCount()); playerId++ {
			units := out.World.GetPlayerUnits(playerId)
			fmt.Printf("ProtoToRuntimeGame: After NewGame - Player %d has %d units\n", playerId, len(units))
			for _, unit := range units {
				fmt.Printf("ProtoToRuntimeGame: Player %d unit at (%d, %d) - DistanceLeft=%d, AvailableHealth=%d\n",
					playerId, unit.Q, unit.R, unit.DistanceLeft, unit.AvailableHealth)
			}
		}
	}
	
	// Set game state from protobuf data
	if out != nil && gameState != nil {
		// Set current player and turn counter from GameState
		out.CurrentPlayer = gameState.CurrentPlayer
		out.TurnCounter = gameState.TurnCounter
	}
	
	return out, nil
}