# Theme System Usage Example

## In Your Presenter

```go
package services

import (
	"bytes"
	"context"
	"fmt"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/web/assets/themes"
	tmpls "github.com/panyam/turnengine/games/weewar/web/templates"
)

type SingletonGameViewPresenterImpl struct {
	BaseGameViewPresenterImpl
	GameViewerPage v1.GameViewerPageClient
	GamesService   *SingletonGamesServiceImpl
	RulesEngine    *v1.RulesEngine

	// Add the theme
	Theme          themes.Theme
}

func NewSingletonGameViewPresenterImpl() *SingletonGameViewPresenterImpl {
	w := &SingletonGameViewPresenterImpl{
		BaseGameViewPresenterImpl: BaseGameViewPresenterImpl{},
		// Initialize with default theme
		Theme: themes.NewDefaultTheme(),
	}
	return w
}

func (s *SingletonGameViewPresenterImpl) SetUnitStats(ctx context.Context, unit *v1.Unit) {
	content := s.renderPanelTemplate(ctx, "UnitStatsPanel.templar.html", map[string]any{
		"Unit":       unit,
		"RulesTable": s.GamesService.RuntimeGame.rulesEngine,
		"Theme":      s.Theme,  // Pass theme to template
	})
	s.GameViewerPage.SetUnitStatsContent(ctx, &v1.SetContentRequest{
		InnerHtml: content,
	})
}
```

## In Your Template

### Option 1: Using Path (for PNG themes like Default)

```html
<!-- UnitStatsPanel.templar.html -->
{{ if .Unit }}
  <div id="unit-details">
    <div class="mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
      <span id="unit-icon" class="text-2xl w-14">
        {{ $theme := .Theme }}
        {{ $assetPath := $theme.GetAssetPathForTemplate "unit" .Unit.UnitType .Unit.Player }}
        <img src="{{ $assetPath }}"
             alt="{{ $theme.GetUnitName .Unit.UnitType }}"
             class="w-8 h-8 object-contain"
             style="image-rendering: pixelated;"
             onerror="this.style.display='none'; this.nextSibling.style.display='inline';"/>
        <span style="display:none;">⚔️</span>
      </span>
      <h5 class="font-medium text-gray-900 dark:text-white">
        {{ $theme.GetUnitName .Unit.UnitType }}
      </h5>
      <p class="text-sm text-gray-600 dark:text-gray-300 mt-1">
        {{ $theme.GetUnitDescription .Unit.UnitType }}
      </p>
    </div>
  </div>
{{ end }}
```

### Option 2: Using Inline SVG (future, when ThemeAssets is implemented)

```html
{{ if .Unit }}
  {{ $themeAssets := .ThemeAssets }}
  {{ $assetResult := $themeAssets.GetUnitAsset .Unit.UnitType .Unit.Player }}

  {{ if eq $assetResult.Type "TYPE_SVG" }}
    <!-- Inline SVG with player colors already applied -->
    <div class="unit-icon">
      {{ $assetResult.Data }}
    </div>
  {{ else if eq $assetResult.Type "TYPE_PATH" }}
    <!-- PNG path -->
    <img src="{{ $assetResult.Data }}" alt="Unit" />
  {{ end }}
{{ end }}
```

## Switching Themes

```go
// In your presenter
func (s *SingletonGameViewPresenterImpl) SetTheme(themeName string) error {
	switch themeName {
	case "default":
		s.Theme = themes.NewDefaultTheme()
	case "fantasy":
		theme, err := themes.NewBaseTheme("web/static/assets/themes/fantasy/mapping.json")
		if err != nil {
			return err
		}
		s.Theme = theme
	case "modern":
		theme, err := themes.NewBaseTheme("web/static/assets/themes/modern/mapping.json")
		if err != nil {
			return err
		}
		s.Theme = theme
	default:
		return fmt.Errorf("unknown theme: %s", themeName)
	}
	return nil
}
```

## What You Get Now (Phase 1 - Metadata Only)

✅ Theme-specific unit names ("Infantry" vs "Peasant" vs "Knight")
✅ Theme-specific terrain names ("Castle" vs "Military Base")
✅ Descriptions (if theme provides them)
✅ Asset paths for PNG themes
✅ Template-friendly API

## What Comes Later (Phase 2 - Asset Loading)

⏳ Server-side SVG rendering with player colors
⏳ Inline SVG in templates (no separate file requests)
⏳ Go-based asset processing (optional, can keep TS)

## Benefits of This Approach

1. **Minimal disruption**: Templates work now with PNG paths
2. **Phased migration**: Add ThemeAssets later if needed
3. **Clean separation**: Theme metadata ≠ asset loading
4. **TypeScript still works**: No changes needed to existing TS asset system
5. **Both can coexist**: Go templates use Go themes, TS code uses TS themes
