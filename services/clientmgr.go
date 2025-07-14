package services

import (
	"context"
	"errors"
	"log"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const APP_ID = "weewar"

var ErrNoSuchEntity = errors.New("entity not found")

type ClientMgr struct {
	svcAddr             string
	gamesSvcClient      protos.GamesServiceClient
	gameSetupsSvcClient protos.GameSetupsServiceClient
	mapsSvcClient       protos.MapsServiceClient
	authSvc             *AuthService
	// We may need an auth svc at some point
}

func NewClientMgr(svc_addr string) *ClientMgr {
	log.Println("Client Mgr Svc Addr: ", svc_addr)
	if svc_addr == "" {
		panic("Service Address is nil")
	}
	return &ClientMgr{svcAddr: svc_addr}
}

func (c *ClientMgr) Address() string {
	return c.svcAddr
}

func (c *ClientMgr) ClientContext(ctx context.Context, loggedInUserId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return metadata.AppendToOutgoingContext(context.Background(), "LoggedInUserId", loggedInUserId)
}

func (c *ClientMgr) GetAuthService() *AuthService {
	if c.authSvc == nil {
		c.authSvc = &AuthService{clients: c}
	}
	return c.authSvc
}

// We will have one client per service here
func (c *ClientMgr) GetMapsSvcClient() (out protos.MapsServiceClient, err error) {
	if c.mapsSvcClient == nil {
		log.Println("Addr: ", c.svcAddr)
		mapsSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("cannot connect with server %v", err)
			return nil, err
		}

		c.mapsSvcClient = protos.NewMapsServiceClient(mapsSvcConn)
	}
	return c.mapsSvcClient, nil
}

// We will have one client per service here
func (c *ClientMgr) GetGameSetupsSvcClient() (out protos.GameSetupsServiceClient, err error) {
	if c.gameSetupsSvcClient == nil {
		log.Println("Addr: ", c.svcAddr)
		gameSetupsSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("cannot connect with server %v", err)
			return nil, err
		}

		c.gameSetupsSvcClient = protos.NewGameSetupsServiceClient(gameSetupsSvcConn)
	}
	return c.gameSetupsSvcClient, nil
}

// We will have one client per service here
func (c *ClientMgr) GetGamesSvcClient() (out protos.GamesServiceClient, err error) {
	if c.gamesSvcClient == nil {
		log.Println("Addr: ", c.svcAddr)
		gamesSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("cannot connect with server %v", err)
			return nil, err
		}

		c.gamesSvcClient = protos.NewGamesServiceClient(gamesSvcConn)
	}
	return c.gamesSvcClient, nil
}
