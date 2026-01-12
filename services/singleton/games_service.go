package singleton

import (
	"context"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
	pj "google.golang.org/protobuf/encoding/protojson"
)

type SingletonGamesService struct {
	services.BaseGamesService
	SingletonGame            *v1.Game
	SingletonGameState       *v1.GameState
	SingletonGameMoveHistory *v1.GameMoveHistory

	RuntimeGame *lib.Game
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewSingletonGamesService() *SingletonGamesService {
	w := &SingletonGamesService{
		BaseGamesService: services.BaseGamesService{
			// WorldsService: SingletonWorldsService
		},
		SingletonGame:            &v1.Game{},
		SingletonGameState:       &v1.GameState{},
		SingletonGameMoveHistory: &v1.GameMoveHistory{},
	}
	w.Self = w
	return w
}

func (w *SingletonGamesService) WorldData() *v1.WorldData {
	return w.SingletonGameState.WorldData
}

func (w *SingletonGamesService) SaveMoveGroup(ctx context.Context, gameId string, state *v1.GameState, group *v1.GameMoveGroup) error {
	// NOOP - Will have to do local first coordination etc
	return nil
}

func (w *SingletonGamesService) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (out *lib.Game, err error) {
	if w.RuntimeGame == nil {
		w.RuntimeGame = lib.ProtoToRuntimeGame(w.SingletonGame, w.SingletonGameState)
	}
	return w.RuntimeGame, err
}

func (w *SingletonGamesService) SaveGame(game *v1.Game, state *v1.GameState, history *v1.GameMoveHistory) error {
	// Update singleton instances with new data
	w.SingletonGame = game
	w.SingletonGameState = state
	w.SingletonGameMoveHistory = history
	return nil
}

func (w *SingletonGamesService) Load(
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

func (w *SingletonGamesService) GetGame(ctx context.Context, req *v1.GetGameRequest) (*v1.GetGameResponse, error) {
	return &v1.GetGameResponse{
		Game:    w.SingletonGame,
		State:   w.SingletonGameState,
		History: w.SingletonGameMoveHistory,
	}, nil
}

func (w *SingletonGamesService) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error) {
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

// ListMoves returns moves from game history, optionally filtered by group range
func (w *SingletonGamesService) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (*v1.ListMovesResponse, error) {
	if w.SingletonGameMoveHistory == nil {
		return &v1.ListMovesResponse{}, nil
	}

	var groups []*v1.GameMoveGroup
	for _, group := range w.SingletonGameMoveHistory.Groups {
		// Filter by group range
		if req.FromGroup > 0 && group.GroupNumber < req.FromGroup {
			continue
		}
		if req.ToGroup > 0 && group.GroupNumber > req.ToGroup {
			break
		}
		groups = append(groups, group)
	}

	// Check if there are earlier moves
	hasMore := req.FromGroup > 0 && len(w.SingletonGameMoveHistory.Groups) > 0 &&
		w.SingletonGameMoveHistory.Groups[0].GroupNumber < req.FromGroup

	return &v1.ListMovesResponse{
		MoveGroups: groups,
		HasMore:    hasMore,
	}, nil
}
