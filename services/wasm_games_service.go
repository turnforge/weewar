package services

import (
	"context"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	pj "google.golang.org/protobuf/encoding/protojson"
)

type WasmGamesServiceImpl struct {
	BaseGamesServiceImpl
	SingletonGame            *v1.Game
	SingletonGameState       *v1.GameState
	SingletonGameMoveHistory *v1.GameMoveHistory
	SingletonWorld           *v1.World
	SingletonWorldData       *v1.WorldData

	RuntimeGame *weewar.Game
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewWasmGamesServiceImpl() *WasmGamesServiceImpl {
	w := &WasmGamesServiceImpl{
		BaseGamesServiceImpl: BaseGamesServiceImpl{
			// WorldsService: SingletonWorldsService
		},
		SingletonGame:            &v1.Game{},
		SingletonGameState:       &v1.GameState{},
		SingletonGameMoveHistory: &v1.GameMoveHistory{},
		SingletonWorld:           &v1.World{},
		SingletonWorldData:       &v1.WorldData{},
	}
	w.Self = w
	return w
}

func (w *WasmGamesServiceImpl) GetRuntimeGame(gameId string) (*weewar.Game, error) {
	if w.RuntimeGame == nil {
		// Create the runtime game from the protobuf data
		world := weewar.NewWorld(w.SingletonWorld.Name)

		// Convert protobuf tiles to runtime tiles
		if w.SingletonGameState.WorldData != nil {
			for _, protoTile := range w.SingletonGameState.WorldData.Tiles {
				coord := weewar.AxialCoord{Q: int(protoTile.Q), R: int(protoTile.R)}
				world.SetTileType(coord, int(protoTile.TileType))
			}

			// Convert protobuf units to runtime units
			for _, protoUnit := range w.SingletonGameState.WorldData.Units {
				coord := weewar.AxialCoord{Q: int(protoUnit.Q), R: int(protoUnit.R)}
				unit := &weewar.Unit{
					UnitType:        int(protoUnit.UnitType),
					Coord:           coord,
					Player:          int(protoUnit.Player),
					AvailableHealth: int(protoUnit.AvailableHealth),
					DistanceLeft:    int(protoUnit.DistanceLeft),
					TurnCounter:     int(protoUnit.TurnCounter),
				}
				world.AddUnit(unit)
			}
		}

		// Create the runtime game with default rules engine
		// TODO: Load a proper rules engine or make it configurable
		rulesEngine := &weewar.RulesEngine{}                   // Default rules engine
		game, err := weewar.NewGame(world, rulesEngine, 12345) // Default seed
		if err != nil {
			return nil, err
		}

		// Set game state from protobuf data
		if w.SingletonGame != nil {
			// Map game metadata fields as needed
			// TODO: Map additional game fields from protobuf if needed
		}

		w.RuntimeGame = game
	}

	return w.RuntimeGame, nil
}

func (w *WasmGamesServiceImpl) SaveGame(game *v1.Game, state *v1.GameState, history *v1.GameMoveHistory) error {
	// Update singleton instances with new data
	w.SingletonGame = game
	w.SingletonGameState = state
	w.SingletonGameMoveHistory = history
	return nil
}

func (w *WasmGamesServiceImpl) Load(
	gameBytes []byte,
	gameStateBytes []byte,
	gameMoveHistoryBytes []byte,
	worldBytes []byte,
	worldDataBytes []byte,
) {
	// Now load data from the bytes
	if err := pj.Unmarshal(gameBytes, w.SingletonGame); err != nil {
		panic(err)
	}
	if err := pj.Unmarshal(gameStateBytes, w.SingletonGameState); err != nil {
		panic(err)
	}
	if err := pj.Unmarshal(gameMoveHistoryBytes, w.SingletonGameMoveHistory); err != nil {
		panic(err)
	}
	if err := pj.Unmarshal(worldBytes, w.SingletonWorld); err != nil {
		panic(err)
	}
	if err := pj.Unmarshal(worldDataBytes, w.SingletonWorldData); err != nil {
		panic(err)
	}
}

// WASM-specific implementations that operate on singleton data

func (w *WasmGamesServiceImpl) GetGame(ctx context.Context, req *v1.GetGameRequest) (*v1.GetGameResponse, error) {
	return &v1.GetGameResponse{
		Game:    w.SingletonGame,
		State:   w.SingletonGameState,
		History: w.SingletonGameMoveHistory,
	}, nil
}

func (w *WasmGamesServiceImpl) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error) {
	// Update singleton instances with new data
	if req.NewGame != nil {
		w.SingletonGame = req.NewGame
	}
	if req.NewState != nil {
		w.SingletonGameState = req.NewState
	}
	if req.NewHistory != nil {
		w.SingletonGameMoveHistory = req.NewHistory
	}

	// Invalidate runtime game cache so it gets recreated with new data
	w.RuntimeGame = nil

	return &v1.UpdateGameResponse{
		Game: w.SingletonGame,
	}, nil
}
