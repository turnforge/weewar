package services

import (
	"context"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	pj "google.golang.org/protobuf/encoding/protojson"
)

type SingletonGamesServiceImpl struct {
	BaseGamesServiceImpl
	SingletonGame            *v1.Game
	SingletonGameState       *v1.GameState
	SingletonGameMoveHistory *v1.GameMoveHistory

	RuntimeGame *Game
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewSingletonGamesServiceImpl() *SingletonGamesServiceImpl {
	w := &SingletonGamesServiceImpl{
		BaseGamesServiceImpl: BaseGamesServiceImpl{
			// WorldsService: SingletonWorldsService
		},
		SingletonGame:            &v1.Game{},
		SingletonGameState:       &v1.GameState{},
		SingletonGameMoveHistory: &v1.GameMoveHistory{},
	}
	w.Self = w
	return w
}

func (w *SingletonGamesServiceImpl) WorldData() *v1.WorldData {
	return w.SingletonGameState.WorldData
}

func (w *SingletonGamesServiceImpl) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (out *Game, err error) {
	if w.RuntimeGame == nil {
		w.RuntimeGame = ProtoToRuntimeGame(w.SingletonGame, w.SingletonGameState)
	}
	return w.RuntimeGame, err
}

func (w *SingletonGamesServiceImpl) SaveGame(game *v1.Game, state *v1.GameState, history *v1.GameMoveHistory) error {
	// Update singleton instances with new data
	w.SingletonGame = game
	w.SingletonGameState = state
	w.SingletonGameMoveHistory = history
	return nil
}

func (w *SingletonGamesServiceImpl) Load(
	gameBytes []byte,
	gameStateBytes []byte,
	gameMoveHistoryBytes []byte,
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
}

// WASM-specific implementations that operate on singleton data

func (w *SingletonGamesServiceImpl) GetGame(ctx context.Context, req *v1.GetGameRequest) (*v1.GetGameResponse, error) {
	return &v1.GetGameResponse{
		Game:    w.SingletonGame,
		State:   w.SingletonGameState,
		History: w.SingletonGameMoveHistory,
	}, nil
}

func (w *SingletonGamesServiceImpl) GetGameState(ctx context.Context, req *v1.GetGameStateRequest) (*v1.GetGameStateResponse, error) {
	return &v1.GetGameStateResponse{
		State: w.SingletonGameState,
	}, nil
}

func (w *SingletonGamesServiceImpl) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error) {
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

	// Don't invalidate runtime game cache for WASM singleton - keep it alive
	// w.RuntimeGame = nil  // COMMENTED OUT - keep runtime game alive

	return &v1.UpdateGameResponse{
		Game: w.SingletonGame,
	}, nil
}
