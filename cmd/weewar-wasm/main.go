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
	wasmGameViewPresenter.DamageDistributionPanel = &services.BrowserDamageDistributionPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.DamageDistributionPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.DamageDistributionPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.UnitStatsPanel = &services.BrowserUnitStatsPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.UnitStatsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.UnitStatsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.TerrainStatsPanel = &services.BrowserTerrainStatsPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.TerrainStatsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.TerrainStatsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.GameScene = &services.BrowserGameScene{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.GameScene.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.GameScene.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.TurnOptionsPanel = &services.BrowserTurnOptionsPanel{
		GameViewerPage: exports.GameViewerPage,
	}
	wasmGameViewPresenter.TurnOptionsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.TurnOptionsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	// Register the JavaScript API using generated exports
	exports.RegisterAPI()

	fmt.Println("WeeWar WASM module loaded successfully")

	// Keep the WASM module running
	select {}
}
