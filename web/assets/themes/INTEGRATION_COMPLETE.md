# Theme System Integration - Complete

## What We Built

A hybrid Go + TypeScript theme system that:
- Separates **Go** (metadata, server-side rendering) from **TypeScript** (asset loading, DOM)
- Uses **data-driven configuration** for terrain types and player colors
- Supports three themes: Default (PNG), Fantasy (SVG), Modern (SVG)

## Files Created/Modified

### Core Go Files
```
web/assets/themes/
├── themes.go           # Theme interface with GetEffectivePlayer, GetPlayerColor
├── base.go             # BaseTheme for SVG themes (accepts cityTerrains)
├── default.go          # DefaultTheme for PNG assets
├── fantasy.go          # FantasyTheme (SVG-based, embedded)
├── modern.go           # ModernTheme (SVG-based, embedded)
├── registry.go         # Theme factory with cityTerrains injection
├── png_renderer.go     # PNG world rendering
└── svg_renderer.go     # SVG world rendering with player colors

lib/
└── rules_engine.go     # GetCityTerrains() helper method

assets/
├── embed.go            # Embeds mapping.json files
└── lilbattle-rules.json   # Contains terrainTypes map
```

### TypeScript Files
```
web/assets/themes/
├── BaseTheme.ts        # Base class with playerColors from mapping
├── default.ts          # PNG theme using isCityTile()
├── fantasy.ts          # SVG fantasy theme
└── modern.ts           # SVG modern theme
```

### Configuration Files
```
web/static/assets/themes/
├── default/mapping.json   # With playerColors
├── fantasy/mapping.json   # With playerColors
└── modern/mapping.json    # With playerColors
```

## Data Flow

### Terrain Classification
```
lilbattle-rules.json
    │
    ▼
RulesEngine.GetCityTerrains()
    │
    ▼
Theme constructor (cityTerrains map)
    │
    ▼
theme.GetEffectivePlayer(terrainId, playerId)
    │
    ▼
Returns owner's color for cities, neutral for nature
```

### Player Colors
```
mapping.json (per theme)
    │
    ▼
Theme.GetPlayerColor(playerId)
    │
    ▼
Renderer applies colors to SVG gradients
```

## Key API

### Theme Interface (Go)
```go
type Theme interface {
    GetUnitName(unitId int32) string
    GetTerrainName(terrainId int32) string
    GetEffectivePlayer(terrainId, playerId int32) int32
    GetPlayerColor(playerId int32) *v1.PlayerColor
    // ... other methods
}
```

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

### TypeScript Theme
```typescript
export interface ITheme {
    getUnitName(unitId: number): string | undefined;
    getTerrainName(terrainId: number): string | undefined;
    getPlayerColor(playerId: number): PlayerColor | undefined;
    // ... other methods
}
```

## Design Wins

1. **Single Source of Truth**
   - Terrain types: `lilbattle-rules.json`
   - Player colors: each theme's `mapping.json`
   - No hardcoded arrays in code

2. **Dependency Injection**
   - Themes don't import RulesEngine
   - `cityTerrains` passed at construction time
   - Clean separation of concerns

3. **Theme-Specific Customization**
   - Each theme can define its own player colors
   - Fantasy could have medieval color names
   - Modern could have military color schemes

4. **Shared Configuration**
   - Both Go and TypeScript read same mapping.json
   - Changes propagate automatically

5. **Type Safety**
   - Proto-generated structs in Go
   - TypeScript interfaces match

## Tests

```bash
go test -v ./web/assets/themes/...
# PASS: TestAllThemes
# PASS: TestGetEffectivePlayer
# PASS: ExampleCreateTheme
# PASS: Example_assetPaths
```

## What's Next (Phase 5)

Clean up remaining hardcoded values in TypeScript:
- `CITY_TERRAIN_IDS` array in BaseTheme.ts
- `NATURE_TERRAIN_IDS` array in BaseTheme.ts
- Move these to mapping.json or load from rules
