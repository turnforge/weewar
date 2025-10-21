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
	b.GameViewerPage.SetTurnOptionsContent(ctx, &v1.SetContentRequest{
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
	b.GameViewerPage.SetUnitStatsContent(ctx, &v1.SetContentRequest{
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
	b.GameViewerPage.SetDamageDistributionContent(ctx, &v1.SetContentRequest{
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
	b.GameViewerPage.SetTerrainStatsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
	fmt.Println("After TSP Set")
}

type BrowserGameScene struct {
	BaseGameScene
	GameViewerPage v1.GameViewerPageClient
}

func (b *BrowserGameScene) ClearPaths(ctx context.Context) {
	b.BaseGameScene.ClearPaths(ctx)
	b.GameViewerPage.ClearPaths(ctx, &v1.ClearPathsRequest{})
}

func (b *BrowserGameScene) ClearHighlights(ctx context.Context) {
	b.BaseGameScene.ClearHighlights(ctx)
	b.GameViewerPage.ClearHighlights(ctx, &v1.ClearHighlightsRequest{})
}

func (b *BrowserGameScene) ShowPath(ctx context.Context, p *v1.ShowPathRequest) {
	b.BaseGameScene.ShowPath(ctx, p)
	b.GameViewerPage.ShowPath(ctx, p)
}

func (b *BrowserGameScene) ShowHighlights(ctx context.Context, h *v1.ShowHighlightsRequest) {
	b.BaseGameScene.ShowHighlights(ctx, h)
	b.GameViewerPage.ShowHighlights(ctx, h)
}
