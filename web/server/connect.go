//go:build !wasm
// +build !wasm

package server

import (
	"context"

	"connectrpc.com/connect"
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1s "github.com/turnforge/weewar/gen/go/weewar/v1/services"
)

// ConnectIndexerServiceAdapter adapts the gRPC IndexerService to Connect's interface
type ConnectIndexerServiceAdapter struct {
	client v1s.IndexerServiceClient
}

func NewConnectIndexerServiceAdapter(client v1s.IndexerServiceClient) *ConnectIndexerServiceAdapter {
	return &ConnectIndexerServiceAdapter{client: client}
}

func (a *ConnectIndexerServiceAdapter) EnsureIndexState(ctx context.Context, req *connect.Request[v1.EnsureIndexStateRequest]) (*connect.Response[v1.EnsureIndexStateResponse], error) {
	resp, err := a.client.EnsureIndexState(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// ConnectGamesServiceAdapter adapts the gRPC GamesService to Connect's interface
type ConnectGamesServiceAdapter struct {
	client v1s.GamesServiceClient
}

func NewConnectGamesServiceAdapter(client v1s.GamesServiceClient) *ConnectGamesServiceAdapter {
	return &ConnectGamesServiceAdapter{client: client}
}

func (a *ConnectGamesServiceAdapter) CreateGame(ctx context.Context, req *connect.Request[v1.CreateGameRequest]) (*connect.Response[v1.CreateGameResponse], error) {
	resp, err := a.client.CreateGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) ListGames(ctx context.Context, req *connect.Request[v1.ListGamesRequest]) (*connect.Response[v1.ListGamesResponse], error) {
	resp, err := a.client.ListGames(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) GetGame(ctx context.Context, req *connect.Request[v1.GetGameRequest]) (*connect.Response[v1.GetGameResponse], error) {
	resp, err := a.client.GetGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) GetGames(ctx context.Context, req *connect.Request[v1.GetGamesRequest]) (*connect.Response[v1.GetGamesResponse], error) {
	resp, err := a.client.GetGames(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) DeleteGame(ctx context.Context, req *connect.Request[v1.DeleteGameRequest]) (*connect.Response[v1.DeleteGameResponse], error) {
	resp, err := a.client.DeleteGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) UpdateGame(ctx context.Context, req *connect.Request[v1.UpdateGameRequest]) (*connect.Response[v1.UpdateGameResponse], error) {
	resp, err := a.client.UpdateGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) ProcessMoves(ctx context.Context, req *connect.Request[v1.ProcessMovesRequest]) (*connect.Response[v1.ProcessMovesResponse], error) {
	resp, err := a.client.ProcessMoves(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) GetOptionsAt(ctx context.Context, req *connect.Request[v1.GetOptionsAtRequest]) (*connect.Response[v1.GetOptionsAtResponse], error) {
	resp, err := a.client.GetOptionsAt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) ListMoves(ctx context.Context, req *connect.Request[v1.ListMovesRequest]) (*connect.Response[v1.ListMovesResponse], error) {
	resp, err := a.client.ListMoves(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) SimulateAttack(ctx context.Context, req *connect.Request[v1.SimulateAttackRequest]) (*connect.Response[v1.SimulateAttackResponse], error) {
	resp, err := a.client.SimulateAttack(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) SimulateFix(ctx context.Context, req *connect.Request[v1.SimulateFixRequest]) (*connect.Response[v1.SimulateFixResponse], error) {
	resp, err := a.client.SimulateFix(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) GetGameState(ctx context.Context, req *connect.Request[v1.GetGameStateRequest]) (*connect.Response[v1.GetGameStateResponse], error) {
	resp, err := a.client.GetGameState(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

/** If you had a streamer than you can use this to act as a bridge between websocket and grpc streams
func (a *ConnectGameServiceAdapter) StreamSomeThing(ctx context.Context, req *connect.Request[v1.StreamSomeThingRequest], stream *connect.ServerStream[v1.StreamSomeThingResponse]) error {
	// Create a custom stream implementation that bridges to Connect
	bridgeStream := &ConnectStreamBridge[v1.StreamSomeThingResponse]{
		connectStream: stream,
		ctx:           ctx,
	}

	// Call your existing gRPC streaming method
	return a.client.StreamSomeThing(req.Msg, bridgeStream)
}
*/

// ConnectWorldsServiceAdapter adapts the gRPC WorldsService to Connect's interface
type ConnectWorldsServiceAdapter struct {
	client v1s.WorldsServiceClient
}

func NewConnectWorldsServiceAdapter(client v1s.WorldsServiceClient) *ConnectWorldsServiceAdapter {
	return &ConnectWorldsServiceAdapter{client: client}
}

func (a *ConnectWorldsServiceAdapter) CreateWorld(ctx context.Context, req *connect.Request[v1.CreateWorldRequest]) (*connect.Response[v1.CreateWorldResponse], error) {
	resp, err := a.client.CreateWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) ListWorlds(ctx context.Context, req *connect.Request[v1.ListWorldsRequest]) (*connect.Response[v1.ListWorldsResponse], error) {
	resp, err := a.client.ListWorlds(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) GetWorld(ctx context.Context, req *connect.Request[v1.GetWorldRequest]) (*connect.Response[v1.GetWorldResponse], error) {
	resp, err := a.client.GetWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) GetWorlds(ctx context.Context, req *connect.Request[v1.GetWorldsRequest]) (*connect.Response[v1.GetWorldsResponse], error) {
	resp, err := a.client.GetWorlds(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) DeleteWorld(ctx context.Context, req *connect.Request[v1.DeleteWorldRequest]) (*connect.Response[v1.DeleteWorldResponse], error) {
	resp, err := a.client.DeleteWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) UpdateWorld(ctx context.Context, req *connect.Request[v1.UpdateWorldRequest]) (*connect.Response[v1.UpdateWorldResponse], error) {
	resp, err := a.client.UpdateWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// ConnectGameSyncServiceAdapter adapts the gRPC GameSyncService to Connect's interface
// This enables multiplayer sync via HTTP/Connect for frontend clients
type ConnectGameSyncServiceAdapter struct {
	client v1s.GameSyncServiceClient
}

func NewConnectGameSyncServiceAdapter(client v1s.GameSyncServiceClient) *ConnectGameSyncServiceAdapter {
	return &ConnectGameSyncServiceAdapter{client: client}
}

func (a *ConnectGameSyncServiceAdapter) Subscribe(ctx context.Context, req *connect.Request[v1.SubscribeRequest], stream *connect.ServerStream[v1.GameUpdate]) error {
	// Call the gRPC streaming method
	grpcStream, err := a.client.Subscribe(ctx, req.Msg)
	if err != nil {
		return err
	}

	// Forward messages from gRPC stream to Connect stream
	for {
		update, err := grpcStream.Recv()
		if err != nil {
			// Stream closed or error
			return err
		}
		if err := stream.Send(update); err != nil {
			return err
		}
	}
}

func (a *ConnectGameSyncServiceAdapter) Broadcast(ctx context.Context, req *connect.Request[v1.BroadcastRequest]) (*connect.Response[v1.BroadcastResponse], error) {
	resp, err := a.client.Broadcast(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}
