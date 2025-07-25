package server

import (
	"context"

	"connectrpc.com/connect"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
)

// ConnectGamesServiceAdapter adapts the gRPC GamesService to Connect's interface
type ConnectGamesServiceAdapter struct {
	svc *services.GamesServiceImpl
}

func NewConnectGamesServiceAdapter(svc *services.GamesServiceImpl) *ConnectGamesServiceAdapter {
	return &ConnectGamesServiceAdapter{svc: svc}
}

func (a *ConnectGamesServiceAdapter) CreateGame(ctx context.Context, req *connect.Request[v1.CreateGameRequest]) (*connect.Response[v1.CreateGameResponse], error) {
	resp, err := a.svc.CreateGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) ListGames(ctx context.Context, req *connect.Request[v1.ListGamesRequest]) (*connect.Response[v1.ListGamesResponse], error) {
	resp, err := a.svc.ListGames(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) GetGame(ctx context.Context, req *connect.Request[v1.GetGameRequest]) (*connect.Response[v1.GetGameResponse], error) {
	resp, err := a.svc.GetGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) GetGames(ctx context.Context, req *connect.Request[v1.GetGamesRequest]) (*connect.Response[v1.GetGamesResponse], error) {
	resp, err := a.svc.GetGames(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) DeleteGame(ctx context.Context, req *connect.Request[v1.DeleteGameRequest]) (*connect.Response[v1.DeleteGameResponse], error) {
	resp, err := a.svc.DeleteGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) UpdateGame(ctx context.Context, req *connect.Request[v1.UpdateGameRequest]) (*connect.Response[v1.UpdateGameResponse], error) {
	resp, err := a.svc.UpdateGame(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectGamesServiceAdapter) ProcessMoves(ctx context.Context, req *connect.Request[v1.ProcessMovesRequest]) (*connect.Response[v1.ProcessMovesResponse], error) {
	resp, err := a.svc.ProcessMoves(ctx, req.Msg)
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
	return a.svc.StreamSomeThing(req.Msg, bridgeStream)
}
*/

// ConnectWorldsServiceAdapter adapts the gRPC WorldsService to Connect's interface
type ConnectWorldsServiceAdapter struct {
	svc *services.WorldsServiceImpl
}

func NewConnectWorldsServiceAdapter(svc *services.WorldsServiceImpl) *ConnectWorldsServiceAdapter {
	return &ConnectWorldsServiceAdapter{svc: svc}
}

func (a *ConnectWorldsServiceAdapter) CreateWorld(ctx context.Context, req *connect.Request[v1.CreateWorldRequest]) (*connect.Response[v1.CreateWorldResponse], error) {
	resp, err := a.svc.CreateWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) ListWorlds(ctx context.Context, req *connect.Request[v1.ListWorldsRequest]) (*connect.Response[v1.ListWorldsResponse], error) {
	resp, err := a.svc.ListWorlds(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) GetWorld(ctx context.Context, req *connect.Request[v1.GetWorldRequest]) (*connect.Response[v1.GetWorldResponse], error) {
	resp, err := a.svc.GetWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) GetWorlds(ctx context.Context, req *connect.Request[v1.GetWorldsRequest]) (*connect.Response[v1.GetWorldsResponse], error) {
	resp, err := a.svc.GetWorlds(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) DeleteWorld(ctx context.Context, req *connect.Request[v1.DeleteWorldRequest]) (*connect.Response[v1.DeleteWorldResponse], error) {
	resp, err := a.svc.DeleteWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectWorldsServiceAdapter) UpdateWorld(ctx context.Context, req *connect.Request[v1.UpdateWorldRequest]) (*connect.Response[v1.UpdateWorldResponse], error) {
	resp, err := a.svc.UpdateWorld(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

/** If you had a streamer than you can use this to act as a bridge between websocket and grpc streams
func (a *ConnectWorldServiceAdapter) StreamSomeThing(ctx context.Context, req *connect.Request[v1.StreamSomeThingRequest], stream *connect.ServerStream[v1.StreamSomeThingResponse]) error {
	// Create a custom stream implementation that bridges to Connect
	bridgeStream := &ConnectStreamBridge[v1.StreamSomeThingResponse]{
		connectStream: stream,
		ctx:           ctx,
	}

	// Call your existing gRPC streaming method
	return a.svc.StreamSomeThing(req.Msg, bridgeStream)
}
*/
