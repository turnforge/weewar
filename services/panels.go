package services

import (
	"bytes"
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/web/assets/themes"
	tmpls "github.com/panyam/turnengine/games/weewar/web/templates"
)

// Data-Only panel implementations

type PanelBase struct {
	Theme       themes.Theme
	RulesEngine *v1.RulesEngine
}

func (p *PanelBase) SetTheme(t themes.Theme) {
	p.Theme = t
}

func (p *PanelBase) SetRulesEngine(r *v1.RulesEngine) {
	p.RulesEngine = r
}

// BaseGameState is a non-UI implementation of GameState interface
// Used for CLI and testing - stores game state without rendering
type BaseGameState struct {
	Game  *v1.Game
	State *v1.GameState
}

func (b *BaseGameState) SetGameState(_ context.Context, req *v1.SetGameStateRequest) (*v1.SetGameStateResponse, error) {
	b.Game = req.Game
	b.State = req.State
	return nil, nil
}

func (b *BaseGameState) SetUnitAt(_ context.Context, req *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error) {
	if b.State == nil || b.State.WorldData == nil {
		return nil, fmt.Errorf("game state not initialized")
	}

	// Find and update or add unit
	found := false
	for i, unit := range b.State.WorldData.Units {
		if unit.Q == req.Q && unit.R == req.R {
			b.State.WorldData.Units[i] = req.Unit
			found = true
			break
		}
	}

	if !found {
		b.State.WorldData.Units = append(b.State.WorldData.Units, req.Unit)
	}

	return nil, nil
}

func (b *BaseGameState) RemoveUnitAt(_ context.Context, req *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error) {
	if b.State == nil || b.State.WorldData == nil {
		return nil, fmt.Errorf("game state not initialized")
	}

	// Remove unit at coordinate
	for i, unit := range b.State.WorldData.Units {
		if unit.Q == req.Q && unit.R == req.R {
			b.State.WorldData.Units = append(b.State.WorldData.Units[:i], b.State.WorldData.Units[i+1:]...)
			break
		}
	}

	return nil, nil
}

func (b *BaseGameState) UpdateGameStatus(_ context.Context, req *v1.UpdateGameStatusRequest) (*v1.UpdateGameStatusResponse, error) {
	if b.State == nil {
		return nil, fmt.Errorf("game state not initialized")
	}

	b.State.CurrentPlayer = req.CurrentPlayer
	b.State.TurnCounter = req.TurnCounter

	return nil, nil
}

type BaseUnitPanel struct {
	PanelBase
	Unit *v1.Unit
}

type BaseTilePanel struct {
	PanelBase
	Tile *v1.Tile
}

type BaseGameScene struct {
	PanelBase
	CurrentPathsRequest      *v1.ShowPathRequest
	CurrentHighlightsRequest *v1.ShowHighlightsRequest
}

func (b *BaseGameScene) ClearPaths(context.Context) {
	b.CurrentPathsRequest = nil
}

func (b *BaseGameScene) ClearHighlights(context.Context) {
	b.CurrentHighlightsRequest = nil
}

func (b *BaseGameScene) ShowPath(_ context.Context, p *v1.ShowPathRequest) {
	b.CurrentPathsRequest = p
}

func (b *BaseGameScene) ShowHighlights(_ context.Context, h *v1.ShowHighlightsRequest) {
	b.CurrentHighlightsRequest = h
}

// Animation methods - no-ops for CLI
func (b *BaseGameScene) MoveUnit(_ context.Context, _ *v1.MoveUnitAnimationRequest) (*v1.MoveUnitAnimationResponse, error) {
	return &v1.MoveUnitAnimationResponse{}, nil
}

func (b *BaseGameScene) ShowAttackEffect(_ context.Context, _ *v1.ShowAttackEffectRequest) (*v1.ShowAttackEffectResponse, error) {
	return &v1.ShowAttackEffectResponse{}, nil
}

func (b *BaseGameScene) ShowHealEffect(_ context.Context, _ *v1.ShowHealEffectRequest) (*v1.ShowHealEffectResponse, error) {
	return &v1.ShowHealEffectResponse{}, nil
}

func (b *BaseGameScene) ShowCaptureEffect(_ context.Context, _ *v1.ShowCaptureEffectRequest) (*v1.ShowCaptureEffectResponse, error) {
	return &v1.ShowCaptureEffectResponse{}, nil
}

func (b *BaseGameScene) SetUnitAt(_ context.Context, _ *v1.SetUnitAtAnimationRequest) (*v1.SetUnitAtAnimationResponse, error) {
	return &v1.SetUnitAtAnimationResponse{}, nil
}

func (b *BaseGameScene) RemoveUnitAt(_ context.Context, _ *v1.RemoveUnitAtAnimationRequest) (*v1.RemoveUnitAtAnimationResponse, error) {
	return &v1.RemoveUnitAtAnimationResponse{}, nil
}

type BaseTurnOptionsPanel struct {
	BaseUnitPanel
	Options *v1.GetOptionsAtResponse
}

func (b *BaseTurnOptionsPanel) CurrentOptions() *v1.GetOptionsAtResponse {
	return b.Options
}

func (b *BaseTurnOptionsPanel) SetCurrentUnit(_ context.Context, unit *v1.Unit, options *v1.GetOptionsAtResponse) {
	b.Unit = unit
	if options == nil {
		options = &v1.GetOptionsAtResponse{}
	}
	b.Options = options
}

func (b *BaseUnitPanel) CurrentUnit() *v1.Unit {
	return b.Unit
}

func (b *BaseUnitPanel) SetCurrentUnit(_ context.Context, u *v1.Unit) {
	b.Unit = u
}

func (b *BaseTilePanel) CurrentTile() *v1.Tile {
	return b.Tile
}

func (b *BaseTilePanel) SetCurrentTile(_ context.Context, u *v1.Tile) {
	b.Tile = u
}

type BaseBuildOptionsModal struct {
	PanelBase
	BuildOptions []*v1.BuildUnitAction
	Tile         *v1.Tile
	PlayerCoins  int32
}

func (b *BaseBuildOptionsModal) Show(_ context.Context, tile *v1.Tile, buildOptions []*v1.BuildUnitAction, playerCoins int32) {
	b.Tile = tile
	b.BuildOptions = buildOptions
	b.PlayerCoins = playerCoins
}

func (b *BaseBuildOptionsModal) Hide(_ context.Context) {
	b.Tile = nil
	b.BuildOptions = nil
	b.PlayerCoins = 0
}

// Browser specific panel implementations

func renderPanelTemplate(_ context.Context, templatefile string, data any) (content string) {
	tmpl, err := tmpls.Templates.Loader.Load(templatefile, "")
	if err == nil {
		buf := bytes.NewBufferString("")
		err = tmpls.Templates.RenderHtmlTemplate(buf, tmpl[0], "", data, nil)
		if err == nil {
			content = buf.String()
		}
	}
	if err != nil {
		panic(err)
	}
	return
}

type BrowserTurnOptionsPanel struct {
	BaseTurnOptionsPanel
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserTurnOptionsPanel) SetCurrentUnit(ctx context.Context, unit *v1.Unit, options *v1.GetOptionsAtResponse) {
	b.BaseTurnOptionsPanel.SetCurrentUnit(ctx, unit, options)
	content := renderPanelTemplate(ctx, "TurnOptionsPanel.templar.html", map[string]any{
		"Options": b.Options.Options,
		"Unit":    unit,
		"Theme":   b.Theme,
	})
	go b.GameViewerPage.SetTurnOptionsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}

type BrowserUnitStatsPanel struct {
	BaseUnitPanel
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserUnitStatsPanel) SetCurrentUnit(ctx context.Context, unit *v1.Unit) {
	content := renderPanelTemplate(ctx, "UnitStatsPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": b.RulesEngine,
		"Theme":      b.Theme, // Pass theme to template
	})
	go b.GameViewerPage.SetUnitStatsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}

type BrowserDamageDistributionPanel struct {
	BaseUnitPanel
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserDamageDistributionPanel) SetCurrentUnit(ctx context.Context, unit *v1.Unit) {
	b.BaseUnitPanel.SetCurrentUnit(ctx, unit)
	fmt.Println("Before DDP Set")
	content := renderPanelTemplate(ctx, "DamageDistributionPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": b.RulesEngine,
		"Theme":      b.Theme, // Pass theme to template
	})
	go b.GameViewerPage.SetDamageDistributionContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
	fmt.Println("After DDP Set")
}

type BrowserTerrainStatsPanel struct {
	BaseTilePanel
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserTerrainStatsPanel) SetCurrentTile(ctx context.Context, tile *v1.Tile) {
	b.BaseTilePanel.SetCurrentTile(ctx, tile)
	fmt.Println("Before TSP Set")
	content := renderPanelTemplate(ctx, "TerrainStatsPanel.templar.html", map[string]any{
		"Tile":       tile,
		"RulesTable": b.RulesEngine,
		"Theme":      b.Theme, // Pass theme to template
	})
	go b.GameViewerPage.SetTerrainStatsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
	fmt.Println("After TSP Set")
}

type BrowserBuildOptionsModal struct {
	BaseBuildOptionsModal
	GameViewerPage v1.GameViewerPageClient
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
	go b.GameViewerPage.ShowBuildOptions(ctx, &v1.ShowBuildOptionsRequest{
		InnerHtml: content,
		Hide:      false,
		Q:         tile.Q,
		R:         tile.R,
	})
}

func (b *BrowserBuildOptionsModal) Hide(ctx context.Context) {
	b.BaseBuildOptionsModal.Hide(ctx)
	go b.GameViewerPage.ShowBuildOptions(ctx, &v1.ShowBuildOptionsRequest{
		Hide: true,
	})
}

type BrowserGameScene struct {
	BaseGameScene
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserGameScene) ClearPaths(ctx context.Context) {
	b.BaseGameScene.ClearPaths(ctx)
	go b.GameViewerPage.ClearPaths(ctx, &v1.ClearPathsRequest{})
}

func (b *BrowserGameScene) ClearHighlights(ctx context.Context) {
	b.BaseGameScene.ClearHighlights(ctx)
	go b.GameViewerPage.ClearHighlights(ctx, &v1.ClearHighlightsRequest{})
}

func (b *BrowserGameScene) ShowPath(ctx context.Context, p *v1.ShowPathRequest) {
	b.BaseGameScene.ShowPath(ctx, p)
	go b.GameViewerPage.ShowPath(ctx, p)
}

func (b *BrowserGameScene) ShowHighlights(ctx context.Context, h *v1.ShowHighlightsRequest) {
	b.BaseGameScene.ShowHighlights(ctx, h)
	go b.GameViewerPage.ShowHighlights(ctx, h)
}

// BrowserGameState is a non-UI implementation of GameState interface
// Used for CLI and testing - stores game state without rendering
type BrowserGameState struct {
	BaseGameState
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserGameState) SetGameState(ctx context.Context, req *v1.SetGameStateRequest) (*v1.SetGameStateResponse, error) {
	b.BaseGameState.SetGameState(ctx, req)
	go b.GameViewerPage.SetGameState(ctx, req)
	return nil, nil
}

func (b *BrowserGameState) SetUnitAt(ctx context.Context, req *v1.SetUnitAtRequest) (*v1.SetUnitAtResponse, error) {
	b.BaseGameState.SetUnitAt(ctx, req)
	go b.GameViewerPage.SetUnitAt(ctx, req)
	return nil, nil
}

func (b *BrowserGameState) RemoveUnitAt(ctx context.Context, req *v1.RemoveUnitAtRequest) (*v1.RemoveUnitAtResponse, error) {
	b.BaseGameState.RemoveUnitAt(ctx, req)
	go b.GameViewerPage.RemoveUnitAt(ctx, req)
	return nil, nil
}

func (b *BrowserGameState) UpdateGameStatus(ctx context.Context, req *v1.UpdateGameStatusRequest) (*v1.UpdateGameStatusResponse, error) {
	b.BaseGameState.UpdateGameStatus(ctx, req)
	go b.GameViewerPage.UpdateGameStatus(ctx, req)
	return nil, nil
}
