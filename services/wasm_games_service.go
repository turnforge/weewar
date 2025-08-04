package services

import (
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	pj "google.golang.org/protobuf/encoding/protojson"
)

type WasmGamesServiceImpl struct {
	BaseGamesServiceImpl
	SingletonGame            *v1.Game
	SingletonGameState       *v1.GameState
	SingletonGameMoveHistory *v1.GameMoveHistory

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
	}
	w.Self = w
	return w
}

func (w *WasmGamesServiceImpl) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (out *weewar.Game, err error) {
	fmt.Printf("GetRuntimeGame: Called - RuntimeGame cached: %t\n", w.RuntimeGame != nil)
	if w.RuntimeGame == nil {
		fmt.Printf("GetRuntimeGame: Creating new runtime game via ProtoToRuntimeGame\n")
		w.RuntimeGame, err = ProtoToRuntimeGame(w.SingletonGame, w.SingletonGameState)
	} else {
		fmt.Printf("GetRuntimeGame: Using cached runtime game\n")
	}
	return w.RuntimeGame, err
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

func (w *WasmGamesServiceImpl) GetGame(ctx context.Context, req *v1.GetGameRequest) (*v1.GetGameResponse, error) {
	return &v1.GetGameResponse{
		Game:    w.SingletonGame,
		State:   w.SingletonGameState,
		History: w.SingletonGameMoveHistory,
	}, nil
}

func (w *WasmGamesServiceImpl) GetGameState(ctx context.Context, req *v1.GetGameStateRequest) (*v1.GetGameStateResponse, error) {
	return &v1.GetGameStateResponse{
		State: w.SingletonGameState,
	}, nil
}

func (w *WasmGamesServiceImpl) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error) {
	fmt.Printf("UpdateGame: Called - this will invalidate RuntimeGame cache!\n")
	fmt.Printf("UpdateGame: NewGame provided: %t\n", req.NewGame != nil)
	fmt.Printf("UpdateGame: NewState provided: %t\n", req.NewState != nil)
	fmt.Printf("UpdateGame: NewHistory provided: %t\n", req.NewHistory != nil)

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
	// The runtime game is the source of truth and should persist across moves
	fmt.Printf("UpdateGame: Keeping RuntimeGame alive (not invalidating cache)\n")
	// w.RuntimeGame = nil  // COMMENTED OUT - keep runtime game alive

	return &v1.UpdateGameResponse{
		Game: w.SingletonGame,
	}, nil
}
