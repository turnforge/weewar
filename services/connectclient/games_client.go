package connectclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/gen/go/lilbattle/v1/services/lilbattlev1connect"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
)

// ConnectGamesClient wraps a Connect client and implements the lib.GamesService interface
type ConnectGamesClient struct {
	services.BaseGamesService
	client lilbattlev1connect.GamesServiceClient
}

// NewConnectGamesClient creates a new Connect client for the GamesService
func NewConnectGamesClient(serverURL string) *ConnectGamesClient {
	return NewConnectGamesClientWithAuth(serverURL, "")
}

// NewConnectGamesClientWithAuth creates a new Connect client with authentication
func NewConnectGamesClientWithAuth(serverURL, token string) *ConnectGamesClient {
	httpClient := http.DefaultClient
	if token != "" {
		httpClient = &http.Client{
			Transport: &authTransport{
				base:  http.DefaultTransport,
				token: token,
			},
		}
	}
	client := lilbattlev1connect.NewGamesServiceClient(
		httpClient,
		serverURL,
	)
	gc := &ConnectGamesClient{
		client: client,
	}
	gc.Self = gc
	return gc
}

func (c *ConnectGamesClient) SaveMoveGroup(ctx context.Context, gameId string, state *v1.GameState, group *v1.GameMoveGroup) error {
	// SaveMoveGroup is internal-only and not exposed as an RPC.
	// The server's ProcessMoves handles validation + save atomically.
	// This client's ProcessMoves delegates to the server, so SaveMoveGroup should never be called.
	panic("SaveMoveGroup should not be called on ConnectGamesClient - use ProcessMoves RPC instead")
}

// CreateGame creates a new game via Connect
func (c *ConnectGamesClient) CreateGame(ctx context.Context, req *v1.CreateGameRequest) (*v1.CreateGameResponse, error) {
	resp, err := c.client.CreateGame(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// GetGames batch gets multiple games by ID via Connect
func (c *ConnectGamesClient) GetGames(ctx context.Context, req *v1.GetGamesRequest) (*v1.GetGamesResponse, error) {
	resp, err := c.client.GetGames(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// ListGames returns all available games via Connect
func (c *ConnectGamesClient) ListGames(ctx context.Context, req *v1.ListGamesRequest) (*v1.ListGamesResponse, error) {
	resp, err := c.client.ListGames(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// GetGame returns a specific game with metadata via Connect
func (c *ConnectGamesClient) GetGame(ctx context.Context, req *v1.GetGameRequest) (*v1.GetGameResponse, error) {
	resp, err := c.client.GetGame(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// DeleteGame deletes a particular game via Connect
func (c *ConnectGamesClient) DeleteGame(ctx context.Context, req *v1.DeleteGameRequest) (*v1.DeleteGameResponse, error) {
	resp, err := c.client.DeleteGame(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// UpdateGame updates a game via Connect
func (c *ConnectGamesClient) UpdateGame(ctx context.Context, req *v1.UpdateGameRequest) (*v1.UpdateGameResponse, error) {
	resp, err := c.client.UpdateGame(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// GetGameState gets the latest game state via Connect
func (c *ConnectGamesClient) GetGameState(ctx context.Context, req *v1.GetGameStateRequest) (*v1.GetGameStateResponse, error) {
	resp, err := c.client.GetGameState(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// ListMoves lists moves for a game via Connect
func (c *ConnectGamesClient) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (*v1.ListMovesResponse, error) {
	resp, err := c.client.ListMoves(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// ProcessMoves processes moves via Connect (delegates to server)
func (c *ConnectGamesClient) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (*v1.ProcessMovesResponse, error) {
	resp, err := c.client.ProcessMoves(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// GetOptionsAt gets options at a position via Connect (delegates to server)
func (c *ConnectGamesClient) GetOptionsAt(ctx context.Context, req *v1.GetOptionsAtRequest) (*v1.GetOptionsAtResponse, error) {
	resp, err := c.client.GetOptionsAt(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// SimulateAttack simulates combat via Connect
func (c *ConnectGamesClient) SimulateAttack(ctx context.Context, req *v1.SimulateAttackRequest) (*v1.SimulateAttackResponse, error) {
	resp, err := c.client.SimulateAttack(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// JoinGame joins an open player slot in a game via Connect
func (c *ConnectGamesClient) JoinGame(ctx context.Context, req *v1.JoinGameRequest) (*v1.JoinGameResponse, error) {
	resp, err := c.client.JoinGame(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// GetRuntimeGame converts proto game data to runtime game
// This is a local operation that doesn't require the server
func (c *ConnectGamesClient) GetRuntimeGame(game *v1.Game, gameState *v1.GameState) (*lib.Game, error) {
	return lib.ProtoToRuntimeGame(game, gameState), nil
}
