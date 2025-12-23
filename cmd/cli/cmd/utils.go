package cmd

import (
	"context"
	"fmt"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	"github.com/turnforge/weewar/lib"
	"github.com/turnforge/weewar/services"
	"github.com/turnforge/weewar/services/connectclient"
	"github.com/turnforge/weewar/services/fsbe"
)

// GameContext holds the game service and loaded data for CLI operations
type GameContext struct {
	Service  services.GamesService
	Game     *v1.Game
	State    *v1.GameState
	History  *v1.GameMoveHistory
	RTGame   *lib.Game
	GameID   string
	IsRemote bool
}

// GetGameContext loads game and creates context for CLI commands
func GetGameContext() (*GameContext, error) {
	id, err := getGameID()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	serverURL := getServerURL()

	// Create the appropriate GamesService
	var svc services.GamesService
	var isRemote bool
	if serverURL != "" {
		svc = connectclient.NewConnectGamesClient(serverURL)
		isRemote = true
		if isVerbose() {
			fmt.Printf("[VERBOSE] Connecting to server: %s\n", serverURL)
		}
	} else {
		svc = fsbe.NewFSGamesService("", nil)
		isRemote = false
		if isVerbose() {
			fmt.Println("[VERBOSE] Using local file storage")
		}
	}

	// Load game
	resp, err := svc.GetGame(ctx, &v1.GetGameRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("failed to load game %s: %w", id, err)
	}

	// Create runtime game for position parsing and rules access
	rtGame, err := svc.GetRuntimeGame(resp.Game, resp.State)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime game: %w", err)
	}

	return &GameContext{
		Service:  svc,
		Game:     resp.Game,
		State:    resp.State,
		History:  resp.History,
		RTGame:   rtGame,
		GameID:   id,
		IsRemote: isRemote,
	}, nil
}
