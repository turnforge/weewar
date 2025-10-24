package main

import (
	"github.com/panyam/turnengine/games/weewar/services"
)

func newPresenter() *services.SingletonGameViewPresenterImpl {
	// Create WASM singleton services (data will be loaded via Load() calls from JS)
	// wasmWorldsService := services.NewSingletonWorldsServiceImpl()
	wasmGamesService := services.NewSingletonGamesServiceImpl()
	wasmGameViewPresenter := services.NewSingletonGameViewPresenterImpl()
	wasmGameViewPresenter.GamesService = wasmGamesService

	// Wire service implementations to generated WASM exports
	wasmGameViewPresenter.GameState = &services.BaseGameState{}
	wasmGameViewPresenter.DamageDistributionPanel = &services.BaseUnitPanel{}
	wasmGameViewPresenter.DamageDistributionPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.DamageDistributionPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.UnitStatsPanel = &services.BaseUnitPanel{}
	wasmGameViewPresenter.UnitStatsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.UnitStatsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.TerrainStatsPanel = &services.BaseTilePanel{}
	wasmGameViewPresenter.TerrainStatsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.TerrainStatsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.GameScene = &services.BaseGameScene{}
	wasmGameViewPresenter.GameScene.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.GameScene.SetRulesEngine(wasmGameViewPresenter.RulesEngine)

	wasmGameViewPresenter.TurnOptionsPanel = &services.BaseTurnOptionsPanel{}
	wasmGameViewPresenter.TurnOptionsPanel.SetTheme(wasmGameViewPresenter.Theme)
	wasmGameViewPresenter.TurnOptionsPanel.SetRulesEngine(wasmGameViewPresenter.RulesEngine)
	return wasmGameViewPresenter
}
