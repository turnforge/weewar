//go:build !wasm
// +build !wasm

// This file is excluded from WASM builds.
// It contains gRPC server setup code that requires net/http packages
// which are not supported by TinyGo's WASM target.

package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	v1s "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1/services"
	"github.com/panyam/turnengine/games/weewar/services/fsbe"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Address string
}

func (s *Server) Start(ctx context.Context, srvErr chan error, srvChan chan bool) error {
	// Use provided CanvasService or create a new one

	server := grpc.NewServer(
	// grpc.UnaryInterceptor(EnsureAccessToken), // Add interceptors if needed
	)

	// Create GamesService
	gamesService := fsbe.NewFSGamesService("")

	// Create coordination storage
	/*
		coordStorageDir := fsbe.DevDataPath("storage/coordination")
		coordStorage, err := coordination.NewFileCoordinationStorage(coordStorageDir)
		if err != nil {
			return fmt.Errorf("failed to create coordination storage: %w", err)
		}

		// Create coordinator service with games service as callback
			coordConfig := coordination.Config{
				RequiredValidators: 1, // Start with 1 for testing
				ValidationTimeout:  5 * time.Minute,
			}
			coordService := coordination.NewService(coordStorage, coordConfig, gamesService)
	*/

	// Register services
	v1s.RegisterGamesServiceServer(server, gamesService)
	v1s.RegisterWorldsServiceServer(server, fsbe.NewFSWorldsService(""))

	// turnengine.RegisterCoordinatorServiceServer(server, coordService)

	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		slog.Error("error in listening on port", "port", s.Address, "err", err)
		// Consider returning the error instead of fatal/panic in Start method
		return fmt.Errorf("failed to listen on %s: %w", s.Address, err)
	}

	// the gRPC server
	slog.Info("Starting grpc endpoint on: ", "addr", s.Address)
	reflection.Register(server)

	// Run server in a goroutine to allow graceful shutdown
	go func() {
		if err := server.Serve(l); err != nil && err != grpc.ErrServerStopped {
			log.Printf("grpc server failed to serve: %v", err)
			srvErr <- err // Send error to the main app routine
		}
	}()

	// Handle shutdown signal
	go func() {
		<-srvChan // Wait for shutdown signal from main app
		slog.Info("Shutting down gRPC server...")
		server.GracefulStop()
		slog.Info("gRPC server stopped.")
	}()

	return nil // Indicate successful start
}
