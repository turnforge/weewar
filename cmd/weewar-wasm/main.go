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

	// Wire service implementations to generated WASM exports
	exports := &weewar_v1_services.Weewar_v1_servicesServicesExports{
		GamesService:  services.NewGamesService(),
		UsersService:  services.NewUsersService(), 
		WorldsService: services.NewWorldsService(),
	}
	
	// Register the JavaScript API using generated exports
	exports.RegisterAPI()
	
	fmt.Println("WeeWar WASM module loaded successfully - service-based architecture")
	
	// Keep the WASM module running
	select {}
}
