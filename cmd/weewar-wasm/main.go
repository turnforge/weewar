//go:build js && wasm
// +build js,wasm

// WeeWar WASM Module - Service Injection Architecture
// This module provides a thin dependency injection layer that wires existing
// service implementations into the generated WASM exports.

package main

import (
	"context"
	"fmt"
	"syscall/js"

	// Generated WASM exports

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1/models"
	weewar_v1_services "github.com/panyam/turnengine/games/weewar/gen/wasm/go/weewar/v1/services"
	"github.com/panyam/turnengine/games/weewar/services/singleton"

	// Service implementations
	"github.com/panyam/turnengine/games/weewar/services"
)

type SingletonInitializerService struct {
	GamesService      *singleton.SingletonGamesService
	GameViewPresenter *services.GameViewPresenter
}

func (s *SingletonInitializerService) InitializeSingleton(ctx context.Context, req *v1.InitializeSingletonRequest) (resp *v1.InitializeSingletonResponse, err error) {
	s.GamesService.Load([]byte(req.GameData), []byte(req.GameState), []byte(req.MoveHistory))
	r1, err := s.GameViewPresenter.InitializeGame(ctx, &v1.InitializeGameRequest{GameId: req.GameId})
	return &v1.InitializeSingletonResponse{Response: r1}, err
}

func main() {
	fmt.Println("WeeWar WASM module loading...")

	// Create WASM singleton services (data will be loaded via Load() calls from JS)
	wasmWorldsService := singleton.NewSingletonWorldsService()
	wasmGamesService := singleton.NewSingletonGamesService()
	wasmGameViewPresenter := services.NewGameViewPresenter()
	wasmGameViewPresenter.GamesService = wasmGamesService
	wasmInitializer := &SingletonInitializerService{
		GamesService:      wasmGamesService,
		GameViewPresenter: wasmGameViewPresenter,
	}

	// Wire service implementations to generated WASM exports
	exports := &weewar_v1_services.Weewar_v1ServicesExports{
		GamesService:      wasmGamesService,
		GameViewPresenter: wasmGameViewPresenter,
		// UsersService:                services.NewUsersService(),
		WorldsService:               wasmWorldsService,
		GameViewerPage:              weewar_v1_services.NewGameViewerPageClient(),
		SingletonInitializerService: wasmInitializer,
	}
	wasmGameViewPresenter.GameState = &BrowserGameState{
		GameViewerPage: exports.GameViewerPage,
	}

	wasmGameViewPresenter.DamageDistributionPanel = &BrowserDamageDistributionPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.DamageDistributionPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.DamageDistributionPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.UnitStatsPanel = &BrowserUnitStatsPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.UnitStatsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.UnitStatsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.TerrainStatsPanel = &BrowserTerrainStatsPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.TerrainStatsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.TerrainStatsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.GameScene = &BrowserGameScene{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.GameScene.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.GameScene.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.TurnOptionsPanel = &BrowserTurnOptionsPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.TurnOptionsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.TurnOptionsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.BuildOptionsModal = &BrowserBuildOptionsModal{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.BuildOptionsModal.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.BuildOptionsModal.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.CompactSummaryCardPanel = &BrowserCompactSummaryCardPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.CompactSummaryCardPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.CompactSummaryCardPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	// Wire GameViewerPage client for mobile-specific RPC calls
	wasmGameViewPresenter.GameViewerPage = exports.GameViewerPage

	// Register the JavaScript API using generated exports
	// wasm.SetGlobalMarshaller(wasm.NewVTProtoMarshallerWithFallback())
	exports.RegisterAPI()

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

		// Call the Load method on SingletonGamesService
		wasmGamesService.Load(gameBytes, gameStateBytes, gameMoveHistoryBytes)

		fmt.Printf("WASM singleton data loaded: game=%d bytes, state=%d bytes, history=%d bytes\n",
			len(gameBytes), len(gameStateBytes), len(gameMoveHistoryBytes))

		return map[string]any{
			"success": true,
			"message": "Game data loaded successfully into WASM singletons",
		}
	}))

	fmt.Println("WeeWar WASM module loaded successfully")

	// Keep the WASM module running
	select {}
}
