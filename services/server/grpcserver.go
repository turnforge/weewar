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
	"os"

	oagrpc "github.com/panyam/oneauth/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Address          string
	RegisterCallback func(server *grpc.Server) error
	// PublicMethods is a list of gRPC method paths that don't require authentication.
	// Format: "/package.Service/Method" e.g. "/lilbattle.v1.WorldsService/ListWorlds"
	PublicMethods []string
}

func (s *Server) Start(ctx context.Context, srvErr chan error, srvChan chan bool) error {
	// Configure auth interceptor
	// Use DISABLE_API_AUTH=true to skip authentication (for local development)
	var authConfig *oagrpc.InterceptorConfig
	if os.Getenv("DISABLE_API_AUTH") == "true" {
		// Optional auth: extract user if present, but don't reject unauthenticated requests
		authConfig = oagrpc.OptionalAuthConfig()
	} else if len(s.PublicMethods) > 0 {
		// Use provided public methods list
		authConfig = oagrpc.NewPublicMethodsConfig(s.PublicMethods...)
	} else {
		// Default: require authentication on all API calls
		authConfig = oagrpc.DefaultInterceptorConfig()
	}

	// Enable switch auth for testing if configured
	if os.Getenv("ENABLE_SWITCH_AUTH") == "true" && authConfig.Config != nil {
		authConfig.Config.EnableSwitchAuth = true
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			oagrpc.UnaryAuthInterceptor(authConfig),
		),
		grpc.ChainStreamInterceptor(
			oagrpc.StreamAuthInterceptor(authConfig),
		),
	)

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
	s.RegisterCallback(server)

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
