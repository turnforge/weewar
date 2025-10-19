# Theme System - Go Implementation Summary

## What We Built

### 1. Proto Definitions (`protos/weewar/v1/themes.proto`)
- `ThemeInfo` - metadata about a theme
- `ThemeManifest` - complete theme configuration (from mapping.json)
- `UnitMapping` / `TerrainMapping` - unit/terrain display data
- `AssetResult` - for future asset loading
- `PlayerColor` - player color scheme

### 2. Core Interfaces (`web/assets/themes/themes.go`)
- **`Theme`** - Lightweight metadata interface (ready to use NOW)
  - GetUnitName, GetTerrainName
  - GetUnitDescription, GetTerrainDescription
  - GetUnitPath, GetTilePath
  - IsCityTile, IsNatureTile, IsBridgeTile
  - GetThemeInfo, HasUnit, HasTerrain

- **`ThemeAssets`** - Heavy asset loading interface (for Phase 2)
  - GetUnitAsset, GetTileAsset
  - LoadUnit, LoadTile
  - ApplyPlayerColors

### 3. Base Implementation (`web/assets/themes/base.go`)
- `BaseTheme` - Common functionality for SVG themes
- Loads from `ThemeManifest` proto
- Provides all metadata accessors

### 4. Concrete Themes

**DefaultTheme** (`web/assets/themes/default.go`)
- PNG-based (original v1 assets)
- No mapping.json needed
- Returns full paths: `/static/assets/v1/Units/1/2.png`
- Hardcoded unit/terrain names

**FantasyTheme** (`web/assets/themes/fantasy.go`)
- SVG-based (Medieval Fantasy)
- Loads from `assets/themes/fantasy/mapping.json` (embedded)
- Returns template paths + metadata
- Unit example: "Peasant" instead of "Infantry"

**ModernTheme** (`web/assets/themes/modern.go`)
- SVG-based (Modern Military)
- Loads from `assets/themes/modern/mapping.json` (embedded)
- Returns template paths + metadata
- Unit example: "Infantry" (but different from default)

### 5. Embedded Assets (`assets/embed.go`)
- Uses `embed.FS` to bundle mapping.json files at compile time
- `LoadThemeManifest(themeName)` - loads fantasy/modern manifests
- `ParseThemeManifest(data)` - converts JSON to proto

### 6. Theme Registry (`web/assets/themes/registry.go`)
- Factory pattern for creating themes by name
- `CreateTheme("fantasy")` - easy theme creation
- `GetAvailableThemes()` - lists all registered themes
- Extensible for future themes

### 7. Shared mapping.json Files
```
assets/themes/
├── fantasy/
│   ├── mapping.json  ✅ Embedded in Go, read by TS
│   ├── Units/*.svg
│   └── Tiles/*.svg
└── modern/
    ├── mapping.json  ✅ Embedded in Go, read by TS
    ├── Units/*.svg
    └── Tiles/*.svg
```

## Usage

### In Your Presenter

```go
import "github.com/panyam/turnengine/games/weewar/web/assets/themes"

type SingletonGameViewPresenterImpl struct {
    Theme themes.Theme
    // ...
}

func NewSingletonGameViewPresenterImpl() *SingletonGameViewPresenterImpl {
    return &SingletonGameViewPresenterImpl{
        Theme: themes.NewDefaultTheme(),
        // Or: Theme: themes.CreateTheme("fantasy")
    }
}

func (s *SingletonGameViewPresenterImpl) SetUnitStats(ctx context.Context, unit *v1.Unit) {
    content := s.renderPanelTemplate(ctx, "UnitStatsPanel.templar.html", map[string]any{
        "Unit":  unit,
        "Theme": s.Theme,  // ✅ Pass theme to template
    })
    // ...
}
```

### In Your Template

```html
{{ if .Unit }}
  {{ $theme := .Theme }}

  <!-- Get unit display name from theme -->
  <h5>{{ $theme.GetUnitName .Unit.UnitType }}</h5>

  <!-- Get unit description -->
  <p>{{ $theme.GetUnitDescription .Unit.UnitType }}</p>

  <!-- Get asset path (works for PNG themes) -->
  {{ $assetPath := $theme.GetAssetPathForTemplate "unit" .Unit.UnitType .Unit.Player }}
  <img src="{{ $assetPath }}" alt="{{ $theme.GetUnitName .Unit.UnitType }}" />
{{ end }}
```

### Switching Themes

```go
// Simple
presenter.Theme = themes.NewDefaultTheme()

// Or using registry
theme, err := themes.CreateTheme("fantasy")
if err == nil {
    presenter.Theme = theme
}
```

## What This Gives You NOW

✅ **Theme-specific names** in Go templates
- Default: "Infantry", "Tank", "Grass"
- Fantasy: "Peasant", "War Cart", "Meadow"
- Modern: "Infantry", "Humvee", "Grassland"

✅ **Theme-specific descriptions** (if theme provides them)

✅ **Asset paths** for PNG themes (Default)
- Full paths ready for `<img src="...">`

✅ **Metadata queries**
- Check if tile is city/nature/bridge
- Get available units/terrains
- Theme version info

✅ **Shared data** between Go and TypeScript
- Both read same `mapping.json` files
- No duplication of unit names, terrain names

✅ **Type safety** via proto definitions

## What Comes Later (Optional - Phase 2)

⏳ **ThemeAssets implementation**
- Server-side SVG rendering
- Apply player colors in Go
- Return inline SVG in templates

⏳ **SVG Processing**
- Load SVG files
- Replace gradient colors
- Convert to data URLs or inline

⏳ **Asset caching**
- Cache processed SVGs
- Pre-generate at build time

**But you don't need Phase 2 yet!** The current implementation gives you theme metadata in Go templates while keeping asset loading in TypeScript where it works great.

## File Locations

```
Root Level:
├── protos/weewar/v1/themes.proto        # Proto definitions
├── assets/
│   ├── embed.go                          # Embeds mapping.json files
│   └── themes/
│       ├── fantasy/mapping.json          # ✅ Embedded at compile time
│       └── modern/mapping.json           # ✅ Embedded at compile time
│
└── web/
    ├── assets/themes/
    │   ├── themes.go                     # Interfaces
    │   ├── base.go                       # BaseTheme
    │   ├── default.go                    # DefaultTheme
    │   ├── fantasy.go                    # FantasyTheme
    │   ├── modern.go                     # ModernTheme
    │   ├── registry.go                   # Theme factory
    │   ├── BaseTheme.ts                  # TS base class (unchanged)
    │   ├── default.ts                    # TS default (unchanged)
    │   ├── fantasy.ts                    # TS fantasy (unchanged)
    │   └── modern.ts                     # TS modern (unchanged)
    │
    └── static/assets/themes/
        ├── fantasy/
        │   ├── mapping.json              # ✅ Read by TS, copied to assets/
        │   ├── Units/*.svg
        │   └── Tiles/*.svg
        └── modern/
            ├── mapping.json              # ✅ Read by TS, copied to assets/
            ├── Units/*.svg
            └── Tiles/*.svg
```

## Key Design Wins

1. **Separation**: Theme (metadata) vs ThemeAssets (I/O)
2. **Shared data**: Single `mapping.json` for Go + TS
3. **Embedded**: mapping.json compiled into binary
4. **Phased**: Use metadata NOW, add assets LATER
5. **Co-located**: Go + TS files together
6. **Type-safe**: Proto-generated structs
7. **Tested**: All themes load and work correctly

## Next Steps

Ready to integrate into your presenter! The theme system is complete and tested. TypeScript code continues working unchanged, and Go templates now have access to theme metadata.
