package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/proto"
)

// newRandomId generates a new unique random ID of specified length (default 8 chars)
// It is upto the caller to check for collissions
func newRandomId(numChars ...int) (string, error) {
	const maxRetries = 10

	// Default to 8 characters if not specified
	length := 8
	if len(numChars) > 0 && numChars[0] > 0 {
		length = numChars[0]
	}

	// Calculate number of bytes needed (2 hex chars per byte)
	numBytes := (length + 1) / 2

	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func newProtoInstance[T proto.Message]() (out T) {
	var zero T
	tType := reflect.TypeOf(zero)

	// If T is a pointer type, create new instance
	if tType.Kind() == reflect.Ptr {
		out = reflect.New(tType.Elem()).Interface().(T)
	} else {
		panic("only pointer types supported")
	}
	return
}

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
			unit := &v1.Unit{
				UnitType:        protoUnit.UnitType,
				Q:               int32(coord.Q),
				R:               int32(coord.R),
				Player:          protoUnit.Player,
				AvailableHealth: protoUnit.AvailableHealth,
				DistanceLeft:    protoUnit.DistanceLeft,
				TurnCounter:     protoUnit.TurnCounter,
			}
			world.AddUnit(unit)
		}
	}

	// Create the runtime game with loaded default rules engine
	rulesEngine := weewar.DefaultRulesEngine()           // Use loaded default rules engine
	out, err := weewar.NewGame(world, rulesEngine, 12345) // Default seed
	if err != nil {
		return nil, err
	}

	// Set game state from protobuf data
	if out != nil && gameState != nil {
		// Set current player and turn counter from GameState
		out.CurrentPlayer = gameState.CurrentPlayer
		out.TurnCounter = gameState.TurnCounter
	}

	return out, nil
}
