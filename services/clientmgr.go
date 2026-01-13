//go:build !wasm
// +build !wasm

// This file is excluded from WASM builds.
// It contains gRPC client code that requires net/http packages
// which are not supported by TinyGo's WASM target.

package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	goalservices "github.com/panyam/goapplib/services"
	v1s "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const APP_ID = "lilbattle"

var ErrNoSuchEntity = errors.New("entity not found")

type ClientMgr struct {
	svcAddr            string
	indexerSvcClient   v1s.IndexerServiceClient
	worldsSvcClient    v1s.WorldsServiceClient
	gamesSvcClient     v1s.GamesServiceClient
	filestoreSvcClient v1s.FileStoreServiceClient
	gameSyncSvcClient  v1s.GameSyncServiceClient
	authSvc *goalservices.AuthService
}

func NewClientMgr(svc_addr string) *ClientMgr {
	if svc_addr == "" {
		panic("Service Address is nil")
	}
	log.Println("Client Mgr Svc Addr: ", svc_addr)
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

func (c *ClientMgr) GetAuthService() *goalservices.AuthService {
	if c.authSvc == nil {
		c.authSvc = &goalservices.AuthService{
			// clients: c
		}
	}
	return c.authSvc
}

func (c *ClientMgr) GetIndexerSvcClient() (out v1s.IndexerServiceClient) {
	if c.indexerSvcClient == nil {
		indexerSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
			// return nil, err
		}

		c.indexerSvcClient = v1s.NewIndexerServiceClient(indexerSvcConn)
	}
	return c.indexerSvcClient
}

func (c *ClientMgr) GetGamesSvcClient() (out v1s.GamesServiceClient) {
	if c.gamesSvcClient == nil {
		gamesSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
			// return nil, err
		}

		c.gamesSvcClient = v1s.NewGamesServiceClient(gamesSvcConn)
	}
	return c.gamesSvcClient

}

func (c *ClientMgr) GetWorldsSvcClient() (out v1s.WorldsServiceClient) {
	if c.worldsSvcClient == nil {
		worldsSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
			// return nil
		}

		c.worldsSvcClient = v1s.NewWorldsServiceClient(worldsSvcConn)
	}
	return c.worldsSvcClient
}

func (c *ClientMgr) GetFileStoreSvcClient() (out v1s.FileStoreServiceClient) {
	log.Println("C = ", c)
	if c.filestoreSvcClient == nil {
		filestoreSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
			// return nil, err
		}

		c.filestoreSvcClient = v1s.NewFileStoreServiceClient(filestoreSvcConn)
	}
	return c.filestoreSvcClient
}

func (c *ClientMgr) GetGameSyncSvcClient() (out v1s.GameSyncServiceClient) {
	if c.gameSyncSvcClient == nil {
		gameSyncSvcConn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
		}

		c.gameSyncSvcClient = v1s.NewGameSyncServiceClient(gameSyncSvcConn)
	}
	return c.gameSyncSvcClient
}
