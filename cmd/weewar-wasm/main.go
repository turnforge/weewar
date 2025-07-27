//go:build js && wasm
// +build js,wasm

// WeeWar WASM Module - Service Injection Architecture
// This module provides a thin dependency injection layer that wires existing
// service implementations into the generated WASM exports.

package main

import (
	"fmt"

	// Generated WASM exports
	weewar_v1_services "github.com/panyam/turnengine/games/weewar/gen/wasm"
	
	// Service implementations
	"github.com/panyam/turnengine/games/weewar/services"
)

func main() {
	fmt.Println("WeeWar WASM module loading...")

	// Create WASM singleton services (data will be loaded via Load() calls from JS)
	wasmGamesService := services.NewWasmGamesServiceImpl()
	wasmWorldsService := services.NewWasmWorldsServiceImpl()

	// Wire service implementations to generated WASM exports
	exports := &weewar_v1_services.Weewar_v1_servicesServicesExports{
		GamesService:  wasmGamesService,
		UsersService:  services.NewUsersService(), 
		WorldsService: wasmWorldsService,
	}
	
	// Register the JavaScript API using generated exports
	exports.RegisterAPI()
	
	fmt.Println("WeeWar WASM module loaded successfully - singleton service architecture")
	
	// Keep the WASM module running
	select {}
}
