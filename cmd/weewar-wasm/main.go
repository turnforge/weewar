//go:build js && wasm
// +build js,wasm

// WeeWar WASM Module - Service Injection Architecture
// This module provides a thin dependency injection layer that wires existing
// service implementations into the generated WASM exports.

package main

import (
	"fmt"
	"syscall/js"

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

	// Add our custom loadGameData function to the existing weewar object created by RegisterAPI()
	weewarObj := js.Global().Get("weewar")
	if !weewarObj.Truthy() {
		fmt.Println("Warning: weewar object not found after RegisterAPI(), creating it")
		weewarObj = js.ValueOf(map[string]any{})
		js.Global().Set("weewar", weewarObj)
	}

	fmt.Println("Adding loadGameData function to existing weewar object")
	weewarObj.Set("loadGameData", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 3 {
			return map[string]any{
				"success": false,
				"error":   "loadGameData requires 3 arguments: gameBytes, gameStateBytes, gameMoveHistoryBytes",
			}
		}

		// Convert JavaScript Uint8Array arguments to Go byte slices
		gameBytes := make([]byte, args[0].Get("length").Int())
		js.CopyBytesToGo(gameBytes, args[0])

		gameStateBytes := make([]byte, args[1].Get("length").Int())
		js.CopyBytesToGo(gameStateBytes, args[1])

		gameMoveHistoryBytes := make([]byte, args[2].Get("length").Int())
		js.CopyBytesToGo(gameMoveHistoryBytes, args[2])

		// Call the Load method on WasmGamesServiceImpl
		wasmGamesService.Load(gameBytes, gameStateBytes, gameMoveHistoryBytes)

		fmt.Printf("WASM singleton data loaded: game=%d bytes, state=%d bytes, history=%d bytes\n",
			len(gameBytes), len(gameStateBytes), len(gameMoveHistoryBytes))

		return map[string]any{
			"success": true,
			"message": "Game data loaded successfully into WASM singletons",
		}
	}))

	fmt.Println("WeeWar WASM module loaded successfully - singleton service architecture")

	// Keep the WASM module running
	select {}
}
