# Theme Architecture - Go + TypeScript

## Overview

This theme system supports both **Go templates** (server-side rendering) and **TypeScript/Phaser** (client-side rendering) with minimal duplication. Configuration is data-driven via shared JSON files.

## File Structure

```
web/assets/themes/
├── themes.go              # Go interfaces (Theme, ThemeAssets)
├── base.go                # BaseTheme implementation (SVG themes)
├── default.go             # DefaultTheme (PNG-based)
├── fantasy.go             # FantasyTheme (SVG-based)
├── modern.go              # ModernTheme (SVG-based)
├── registry.go            # Theme factory with cityTerrains injection
├── png_renderer.go        # PNG world rendering
├── svg_renderer.go        # SVG world rendering with color application
├── BaseTheme.ts           # TypeScript base class
├── default.ts             # TypeScript default theme (PNG)
├── modern.ts              # TypeScript modern theme (SVG)
├── fantasy.ts             # TypeScript fantasy theme (SVG)
└── providers/
    └── AssetProvider.ts   # Handles asset loading in Phaser

web/static/assets/themes/
├── default/
│   └── mapping.json      # Contains units, terrains, playerColors
├── fantasy/
│   ├── mapping.json      # Shared by Go AND TypeScript
│   ├── Units/*.svg
│   └── Tiles/*.svg
└── modern/
    ├── mapping.json      # Shared by Go AND TypeScript
    ├── Units/*.svg
    └── Tiles/*.svg

lib/
└── rules_engine.go       # GetCityTerrains() - terrain classification

assets/
├── embed.go              # Embeds mapping.json files at compile time
└── lilbattle-rules.json     # Contains terrainTypes (single source of truth)
```

## Key Design Decisions

### 1. Data-Driven Configuration

**Terrain Types** (from `lilbattle-rules.json`):
```json
{
  "terrainTypes": {
    "1": "city", "2": "city", "3": "city",
    "5": "nature", "7": "nature",
    "10": "water", "14": "water",
    "17": "bridge", "18": "bridge",
    "22": "road"
  }
}
```

**Player Colors** (from each theme's `mapping.json`):
```json
{
  "playerColors": {
    "0": { "primary": "#888888", "secondary": "#666666", "name": "Neutral" },
    "1": { "primary": "#f87171", "secondary": "#dc2626", "name": "Red" }
  }
}
```

### 2. Dependency Injection

Themes receive `cityTerrains` map from RulesEngine at construction time:
```go
cityTerrains := lib.DefaultRulesEngine().GetCityTerrains()
theme, _ := themes.CreateTheme("fantasy", cityTerrains)
```

This avoids circular dependencies and keeps themes decoupled from game rules.

### 3. Split Interface: Theme vs ThemeAssets

**Theme** = Lightweight, metadata-only
- Unit/terrain names and descriptions
- Path resolution
- `GetEffectivePlayer()` - determines player color for terrain
- `GetPlayerColor()` - retrieves color scheme from mapping

**ThemeAssets** = Heavy, I/O operations
- Loading SVG files
- Applying player colors
- Rendering/rasterization

### 4. Shared mapping.json

Both Go and TypeScript read the **same** `mapping.json` files:
```json
{
  "themeInfo": {
    "name": "Medieval Fantasy",
    "version": "1.0.0",
    "base_path": "/static/assets/themes/fantasy",
    "asset_type": "svg",
    "needs_post_processing": true
  },
  "units": {
    "1": { "name": "Peasant", "image": "Units/Peasant.svg" }
  },
  "terrains": {
    "1": { "name": "Castle", "image": "Tiles/Castle.svg" }
  },
  "playerColors": {
    "0": { "primary": "#888888", "secondary": "#666666", "name": "Neutral" }
  }
}
```

### 5. Co-located Files

Go and TypeScript theme files live together:
- `default.go` next to `default.ts`
- `fantasy.go` next to `fantasy.ts`

## Architecture Diagram

```
                    ┌─────────────────────┐
                    │  lilbattle-rules.json  │
                    │   (terrainTypes)    │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │    RulesEngine      │
                    │  GetCityTerrains()  │
                    └──────────┬──────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
┌─────────▼─────────┐ ┌────────▼────────┐ ┌────────▼────────┐
│   DefaultTheme    │ │  FantasyTheme   │ │   ModernTheme   │
│   (PNG assets)    │ │  (SVG assets)   │ │  (SVG assets)   │
└─────────┬─────────┘ └────────┬────────┘ └────────┬────────┘
          │                    │                    │
          │       ┌────────────┴────────────┐      │
          │       │      mapping.json       │      │
          │       │  (units, terrains,      │      │
          │       │   playerColors)         │      │
          │       └─────────────────────────┘      │
          │                                        │
┌─────────▼─────────┐                    ┌────────▼────────┐
│   PNG Renderer    │                    │   SVG Renderer  │
│  (pre-colored)    │                    │ (applies colors)│
└───────────────────┘                    └─────────────────┘
```

## Key Methods

### GetEffectivePlayer(terrainId, playerId) int32

Determines which player color to use when rendering terrain:
- City terrains (bases, airports, etc.) use the owner's color
- Nature terrains (grass, mountains, etc.) use neutral (player 0)

```go
func (b *BaseTheme) GetEffectivePlayer(terrainId, playerId int32) int32 {
    if b.cityTerrains[terrainId] {
        return playerId
    }
    return 0
}
```

### GetPlayerColor(playerId) *v1.PlayerColor

Returns the color scheme for a player from the theme's mapping.json:
```go
func (b *BaseTheme) GetPlayerColor(playerId int32) *v1.PlayerColor {
    if color, ok := b.manifest.PlayerColors[playerId]; ok {
        return color
    }
    return b.manifest.PlayerColors[0] // fallback to neutral
}
```

## Separation of Concerns

| Concern | Owner | File |
|---------|-------|------|
| Terrain classification | RulesEngine | `lilbattle-rules.json` |
| Player colors | Theme | `mapping.json` |
| Unit/terrain names | Theme | `mapping.json` |
| Effective player logic | Theme | `base.go`, `default.go` |
| Color application | Renderer | `svg_renderer.go` |

## Benefits

1. **Single source of truth**: Terrain types in rules, colors in mapping
2. **No duplication**: Both Go and TS read same JSON files
3. **Dependency injection**: Themes don't import RulesEngine
4. **Theme-specific customization**: Each theme can have different colors
5. **Type safety**: Proto-generated structs throughout
6. **Testable**: Clear boundaries for unit testing
