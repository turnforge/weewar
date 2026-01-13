# Theme System - Go Implementation Summary

## What We Built

### 1. Proto Definitions (`protos/lilbattle/v1/themes.proto`)
- `ThemeInfo` - metadata about a theme
- `ThemeManifest` - complete theme configuration (from mapping.json)
- `UnitMapping` / `TerrainMapping` - unit/terrain display data
- `AssetResult` - for future asset loading
- `PlayerColor` - player color scheme (primary, secondary, name)

### 2. Core Interfaces (`web/assets/themes/themes.go`)
- **`Theme`** - Lightweight metadata interface
  - GetUnitName, GetTerrainName
  - GetUnitDescription, GetTerrainDescription
  - GetUnitPath, GetTilePath
  - GetThemeInfo, HasUnit, HasTerrain
  - GetEffectivePlayer (for terrain color rendering)
  - GetPlayerColor (from mapping.json)

- **`ThemeAssets`** - Heavy asset loading interface
  - GetUnitAsset, GetTileAsset
  - LoadUnit, LoadTile
  - ApplyPlayerColors

### 3. Base Implementation (`web/assets/themes/base.go`)
- `BaseTheme` - Common functionality for SVG themes
- Loads from `ThemeManifest` proto
- Accepts `cityTerrains` map from RulesEngine
- Provides all metadata accessors including GetPlayerColor

### 4. Concrete Themes

**DefaultTheme** (`web/assets/themes/default.go` / `web/assets/themes/default.ts`)
- PNG-based (original v1 assets)
- Uses `mapping.json` with playerColors
- Returns full paths: `/static/assets/themes/default/Units/1/2.png`
- Uses `cityTerrains` from RulesEngine for effective player calculation

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
- `CreateTheme("fantasy", cityTerrains)` - theme creation with cityTerrains
- `GetAvailableThemes()` - lists all registered themes
- Extensible for future themes

### 7. Renderers
- **PNGWorldRenderer** (`png_renderer.go`) - Renders worlds using PNG assets
- **SVGWorldRenderer** (`svg_renderer.go`) - Renders worlds using SVG assets with player color application
- Both use `theme.GetEffectivePlayer()` for terrain color handling
- Both use `theme.GetPlayerColor()` for color lookup

### 8. Label Rendering (PNGWorldRenderer)
The PNG renderer supports optional labels via `RenderOptions`:
- **Unit Labels** (`ShowUnitLabels`): Displays `Shortcut:MP/Health` at bottom of unit with brown background (#3d2817)
- **Tile Labels** (`ShowTileLabels`): Displays tile shortcut at top of tile with teal background (#173d3d)

Used by `ww map` CLI command for terminal map display.

### 9. Shared mapping.json Files
```
web/static/assets/themes/
├── default/
│   └── mapping.json  ✅ Contains units, terrains, playerColors
├── fantasy/
│   ├── mapping.json  ✅ Embedded in Go, read by TS
│   ├── Units/*.svg
│   └── Tiles/*.svg
└── modern/
    ├── mapping.json  ✅ Embedded in Go, read by TS
    ├── Units/*.svg
    └── Tiles/*.svg
```

## Data-Driven Configuration

### Terrain Types (from RulesEngine)
Terrain classification is now data-driven via `lilbattle-rules.json`:
```json
{
  "terrainTypes": {
    "1": "city", "2": "city", "3": "city",  // Bases
    "5": "nature", "7": "nature",            // Grass, Mountains
    "10": "water", "14": "water",            // Water types
    "17": "bridge", "18": "bridge",          // Bridges
    "22": "road"                              // Road
  }
}
```

### Player Colors (defaults in theme loaders)
Player colors are now defined as defaults in the theme loaders (Go and TypeScript).
Themes can optionally override by specifying `playerColors` in their mapping.json.

Default colors (matching sprite sheet order):
```
0: Neutral (gray)
1: Blue      2: Red       3: Yellow    4: White
5: Pink      6: Orange    7: Black     8: Teal
9: Navy Blue 10: Brown    11: Cyan     12: Purple
```

## Usage

### Creating Themes
```go
import (
    "github.com/turnforge/lilbattle/lib"
    "github.com/turnforge/lilbattle/web/assets/themes"
)

// Get city terrains from RulesEngine
cityTerrains := lib.DefaultRulesEngine().GetCityTerrains()

// Create themes
defaultTheme := themes.NewDefaultTheme(cityTerrains)
fantasyTheme, _ := themes.NewFantasyTheme(cityTerrains)
modernTheme, _ := themes.NewModernTheme(cityTerrains)

// Or use registry
theme, _ := themes.CreateTheme("fantasy", cityTerrains)
```

### In Templates
```html
{{ $theme := .Theme }}

<!-- Get unit display name from theme -->
<h5>{{ $theme.GetUnitName .Unit.UnitType }}</h5>

<!-- Get asset path (works for PNG themes) -->
{{ $assetPath := $theme.GetAssetPathForTemplate "unit" .Unit.UnitType .Unit.Player }}
<img src="{{ $assetPath }}" alt="{{ $theme.GetUnitName .Unit.UnitType }}" />
```

## Key Design Wins

1. **Data-driven**: Terrain types from RulesEngine, player colors from mapping.json
2. **No duplication**: Single source of truth for terrain classification
3. **Dependency injection**: cityTerrains passed to theme constructors
4. **Theme-specific colors**: Each theme can customize player colors
5. **Separation**: RulesEngine (game logic) vs Theme (rendering)
6. **Type safety**: Proto-generated structs
7. **Tested**: All themes load and work correctly

## File Locations

```
Root Level:
├── protos/lilbattle/v1/models/
│   └── themes.proto              # Proto definitions
├── assets/
│   ├── embed.go                  # Embeds mapping.json files
│   ├── lilbattle-rules.json         # Contains terrainTypes
│   └── themes/
│       ├── fantasy/mapping.json  # Embedded at compile time
│       └── modern/mapping.json   # Embedded at compile time
│
├── lib/
│   ├── rules_engine.go           # GetCityTerrains(), terrain type methods
│   └── rules_loader.go           # Parses terrainTypes from JSON
│
└── web/
    ├── assets/themes/
    │   ├── themes.go             # Interfaces
    │   ├── base.go               # BaseTheme
    │   ├── default.go            # DefaultTheme
    │   ├── fantasy.go            # FantasyTheme
    │   ├── modern.go             # ModernTheme
    │   ├── registry.go           # Theme factory
    │   ├── png_renderer.go       # PNG world rendering
    │   ├── svg_renderer.go       # SVG world rendering
    │   ├── BaseTheme.ts          # TS base class
    │   ├── default.ts            # TS default
    │   ├── fantasy.ts            # TS fantasy
    │   └── modern.ts             # TS modern
    │
    └── static/assets/themes/
        ├── default/
        │   └── mapping.json      # Units/terrains only (playerColors from loader defaults)
        ├── fantasy/
        │   ├── mapping.json      # Units/terrains only (playerColors from loader defaults)
        │   ├── Units/*.svg
        │   └── Tiles/*.svg
        └── modern/
            ├── mapping.json      # Units/terrains only (playerColors from loader defaults)
            ├── Units/*.svg
            └── Tiles/*.svg
```
