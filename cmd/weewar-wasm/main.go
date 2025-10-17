//go:build js && wasm
// +build js,wasm

// WeeWar WASM Module - Service Injection Architecture
// This module provides a thin dependency injection layer that wires existing
// service implementations into the generated WASM exports.

package main

import (
	"fmt"

	// Generated WASM exports
	weewar_v1_services "github.com/panyam/turnengine/games/weewar/gen/wasm/go/weewar/v1"

	// Service implementations
	"github.com/panyam/turnengine/games/weewar/services"
)

func main() {
	fmt.Println("WeeWar WASM module loading...")

	// Create WASM singleton services (data will be loaded via Load() calls from JS)
	wasmWorldsService := services.NewSingletonWorldsServiceImpl()
	wasmGamesService := services.NewSingletonGamesServiceImpl()
	wasmGameViewPresenter := services.NewSingletonGameViewPresenterImpl()
	wasmGameViewPresenter.GamesService = wasmGamesService

	// Wire service implementations to generated WASM exports
	exports := &weewar_v1_services.Weewar_v1ServicesExports{
		GamesService:      wasmGamesService,
		GameViewPresenter: wasmGameViewPresenter,
		UsersService:      services.NewUsersService(),
		WorldsService:     wasmWorldsService,
		GameViewerPage:    weewar_v1_services.NewGameViewerPageClient(),
	}
	wasmGameViewPresenter.GameViewerPage = exports.GameViewerPage

	// Register the JavaScript API using generated exports
	exports.RegisterAPI()

	fmt.Println("WeeWar WASM module loaded successfully")

	// Keep the WASM module running
	select {}
}
