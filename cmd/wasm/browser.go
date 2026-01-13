//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"fmt"
	"log"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	wasmv1 "github.com/turnforge/lilbattle/gen/wasm/go/lilbattle/v1/services"
	"github.com/turnforge/lilbattle/services"
)

// Browser specific panel and other "client" implementations

type BrowserTurnOptionsPanel struct {
	services.BaseTurnOptionsPanel
	GameViewerPage *wasmv1.GameViewerPageClient
}

func dispatch(label string, action func(), disableGo ...bool) {
	if len(disableGo) > 0 && disableGo[0] {
		log.Println("Dispatching Sync: ", label)
		action()
	} else {
		log.Println("Dispatching With Go: ", label)
		go action()
	}
}

func (b *BrowserTurnOptionsPanel) SetCurrentUnit(ctx context.Context, unit *v1.Unit, options *v1.GetOptionsAtResponse) {
	b.BaseTurnOptionsPanel.SetCurrentUnit(ctx, unit, options)
	content := renderPanelTemplate(ctx, "TurnOptionsPanel.templar.html", map[string]any{
		"Options": b.Options.Options,
		"Unit":    unit,
		"Theme":   b.Theme,
	})
	dispatch("SetTurnOptionsContent", func() {
		b.GameViewerPage.SetTurnOptionsContent(ctx, &v1.SetContentRequest{
			InnerHtml: content,
		})
	}, true)
}

type BrowserUnitStatsPanel struct {
	services.BaseUnitPanel
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserUnitStatsPanel) SetCurrentUnit(ctx context.Context, unit *v1.Unit) {
	content := renderPanelTemplate(ctx, "UnitStatsPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": b.RulesEngine,
		"Theme":      b.Theme, // Pass theme to template
	})
	dispatch("SetUnitStatsContent", func() {
		b.GameViewerPage.SetUnitStatsContent(ctx, &v1.SetContentRequest{
			InnerHtml: content,
		})
	})
}

type BrowserDamageDistributionPanel struct {
	services.BaseUnitPanel
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserDamageDistributionPanel) SetCurrentUnit(ctx context.Context, unit *v1.Unit) {
	b.BaseUnitPanel.SetCurrentUnit(ctx, unit)
	fmt.Println("Before DDP Set")
	content := renderPanelTemplate(ctx, "DamageDistributionPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": b.RulesEngine,
		"Theme":      b.Theme, // Pass theme to template
	})
	dispatch("SetDamageDistributionContent", func() {
		b.GameViewerPage.SetDamageDistributionContent(ctx, &v1.SetContentRequest{
			InnerHtml: content,
		})
	})
	fmt.Println("After DDP Set")
}

type BrowserTerrainStatsPanel struct {
	services.BaseTilePanel
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserTerrainStatsPanel) SetCurrentTile(ctx context.Context, tile *v1.Tile) {
	b.BaseTilePanel.SetCurrentTile(ctx, tile)
	fmt.Println("Before TSP Set")
	content := renderPanelTemplate(ctx, "TerrainStatsPanel.templar.html", map[string]any{
		"Tile":       tile,
		"RulesTable": b.RulesEngine,
		"Theme":      b.Theme, // Pass theme to template
	})
	dispatch("SetTerrainStatsContent", func() {
		b.GameViewerPage.SetTerrainStatsContent(ctx, &v1.SetContentRequest{
			InnerHtml: content,
		})
	})
	fmt.Println("After TSP Set")
}

type BrowserCompactSummaryCardPanel struct {
	services.PanelBase
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserCompactSummaryCardPanel) SetCurrentData(ctx context.Context, tile *v1.Tile, unit *v1.Unit) {
	content := renderPanelTemplate(ctx, "CompactSummaryCard.templar.html", map[string]any{
		"Tile":  tile,
		"Unit":  unit,
		"Theme": b.Theme,
	})
	dispatch("SetCompactSummaryCard", func() {
		b.GameViewerPage.SetCompactSummaryCard(ctx, &v1.SetContentRequest{
			InnerHtml: content,
		})
	})
}

type BrowserBuildOptionsModal struct {
	services.BaseBuildOptionsModal
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserBuildOptionsModal) Show(ctx context.Context, tile *v1.Tile, buildOptions []*v1.BuildUnitAction, playerCoins int32) {
	b.BaseBuildOptionsModal.Show(ctx, tile, buildOptions, playerCoins)
	fmt.Printf("[BrowserBuildOptionsModal] Show called with %d options, tile at (%d,%d), coins=%d\n",
		len(buildOptions), tile.Q, tile.R, playerCoins)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[BrowserBuildOptionsModal] PANIC during template rendering: %v\n", r)
		}
	}()

	content := renderPanelTemplate(ctx, "BuildOptionsModal.templar.html", map[string]any{
		"BuildOptions": buildOptions,
		"Tile":         tile,
		"PlayerCoins":  playerCoins,
		"Theme":        b.Theme,
		"RulesTable":   b.RulesEngine,
	})
	fmt.Printf("[BrowserBuildOptionsModal] Template rendered successfully, content length=%d\n", len(content))
	dispatch("ShowBuildOptions", func() {
		b.GameViewerPage.ShowBuildOptions(ctx, &v1.ShowBuildOptionsRequest{
			InnerHtml: content,
			Hide:      false,
			Q:         tile.Q,
			R:         tile.R,
		})
	})
}

func (b *BrowserBuildOptionsModal) Hide(ctx context.Context) {
	b.BaseBuildOptionsModal.Hide(ctx)
	dispatch("ShowBuildOptions", func() { b.GameViewerPage.ShowBuildOptions(ctx, &v1.ShowBuildOptionsRequest{Hide: true}) })
}

type BrowserGameScene struct {
	services.BaseGameScene
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserGameScene) ClearPaths(ctx context.Context) {
	b.BaseGameScene.ClearPaths(ctx)
	dispatch("ClearPaths", func() {
		b.GameViewerPage.ClearPaths(ctx, &v1.ClearPathsRequest{})
	})
}

func (b *BrowserGameScene) ClearHighlights(ctx context.Context, req *v1.ClearHighlightsRequest) {
	b.BaseGameScene.ClearHighlights(ctx, req)
	if req == nil {
		req = &v1.ClearHighlightsRequest{} // Clear all if no request provided
	}
	dispatch("ClearHighlights", func() {
		b.GameViewerPage.ClearHighlights(ctx, req)
	})
}

func (b *BrowserGameScene) ShowPath(ctx context.Context, p *v1.ShowPathRequest) {
	b.BaseGameScene.ShowPath(ctx, p)
	dispatch("ShowPath", func() {
		b.GameViewerPage.ShowPath(ctx, p)
	})
}

func (b *BrowserGameScene) ShowHighlights(ctx context.Context, h *v1.ShowHighlightsRequest) {
	b.BaseGameScene.ShowHighlights(ctx, h)
	dispatch("ShowHighlights", func() {
		b.GameViewerPage.ShowHighlights(ctx, h)
	})
}

// Animation methods - forward to browser
func (b *BrowserGameScene) MoveUnit(ctx context.Context, req *v1.MoveUnitRequest) (*v1.MoveUnitResponse, error) {
	b.BaseGameScene.MoveUnit(ctx, req)
	dispatch("MoveUnit", func() {
		b.GameViewerPage.MoveUnit(ctx, req)
	})
	return &v1.MoveUnitResponse{}, nil
}

func (b *BrowserGameScene) SetUnitAt(ctx context.Context, req *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error) {
	b.BaseGameScene.SetUnitAt(ctx, req)
	dispatch("SetUnitAt", func() {
		b.GameViewerPage.SetUnitAt(ctx, req)
	})
	return &v1.SetUnitAtResponse{}, nil
}

func (b *BrowserGameScene) RemoveUnitAt(ctx context.Context, req *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error) {
	b.BaseGameScene.RemoveUnitAt(ctx, req)
	dispatch("RemoveUnitAt", func() {
		b.GameViewerPage.RemoveUnitAt(ctx, req)
	})
	return &v1.RemoveUnitAtResponse{}, nil
}

// BrowserGameState is a non-UI implementation of GameState interface
// Used for CLI and testing - stores game state without rendering
type BrowserGameState struct {
	services.BaseGameState
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserGameState) SetGameState(ctx context.Context, req *v1.SetGameStateRequest) (*v1.SetGameStateResponse, error) {
	b.BaseGameState.SetGameState(ctx, req)
	dispatch("SetGameState", func() {
		b.GameViewerPage.SetGameState(ctx, req)
	})
	return nil, nil
}

func (b *BrowserGameState) SetUnitAt(ctx context.Context, req *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error) {
	b.BaseGameState.SetUnitAt(ctx, req)
	dispatch("SetUnitAt", func() {
		b.GameViewerPage.SetUnitAt(ctx, req)
	})
	return nil, nil
}

func (b *BrowserGameState) RemoveUnitAt(ctx context.Context, req *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error) {
	b.BaseGameState.RemoveUnitAt(ctx, req)
	dispatch("RemoveUnitAt", func() {
		b.GameViewerPage.RemoveUnitAt(ctx, req)
	})
	return nil, nil
}

func (b *BrowserGameState) UpdateGameStatus(ctx context.Context, req *v1.UpdateGameStatusRequest) (*v1.UpdateGameStatusResponse, error) {
	b.BaseGameState.UpdateGameStatus(ctx, req)
	dispatch("UpdateGameStatus", func() {
		b.GameViewerPage.UpdateGameStatus(ctx, req)
	})
	return nil, nil
}

type BrowserGameStatePanel struct {
	services.BaseGameStatePanel
	GameViewerPage *wasmv1.GameViewerPageClient
}

func (b *BrowserGameStatePanel) Update(ctx context.Context, game *v1.Game, state *v1.GameState) {
	b.BaseGameStatePanel.Update(ctx, game, state)
	// Pass the panel itself to the template so it can access computed fields
	content := renderPanelTemplate(ctx, "GameStatePanel.templar.html", b)
	dispatch("SetGameStatePanelContent", func() {
		b.GameViewerPage.SetGameStatePanelContent(ctx, &v1.SetContentRequest{
			InnerHtml: content,
		})
	})
}
