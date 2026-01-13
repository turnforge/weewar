# Theme System Usage Example

## Creating Themes in Go

```go
package services

import (
    "github.com/turnforge/lilbattle/lib"
    "github.com/turnforge/lilbattle/web/assets/themes"
)

type SingletonGameViewPresenterImpl struct {
    BaseGameViewPresenterImpl
    Theme themes.Theme
}

func NewSingletonGameViewPresenterImpl() *SingletonGameViewPresenterImpl {
    // Get city terrains from RulesEngine (single source of truth)
    cityTerrains := lib.DefaultRulesEngine().GetCityTerrains()

    return &SingletonGameViewPresenterImpl{
        Theme: themes.NewDefaultTheme(cityTerrains),
    }
}

// Switch themes at runtime
func (s *SingletonGameViewPresenterImpl) SetTheme(themeName string) error {
    cityTerrains := lib.DefaultRulesEngine().GetCityTerrains()
    theme, err := themes.CreateTheme(themeName, cityTerrains)
    if err != nil {
        return err
    }
    s.Theme = theme
    return nil
}
```

## Using Theme in Templates

### PNG Theme (Default)

```html
{{ $theme := .Theme }}
{{ $unit := .Unit }}

<!-- Unit name from theme -->
<h5>{{ $theme.GetUnitName $unit.UnitType }}</h5>

<!-- Asset path for PNG themes -->
{{ $assetPath := $theme.GetAssetPathForTemplate "unit" $unit.UnitType $unit.Player }}
<img src="{{ $assetPath }}"
     alt="{{ $theme.GetUnitName $unit.UnitType }}"
     class="w-8 h-8 object-contain"
     style="image-rendering: pixelated;" />
```

### Terrain with Player Colors

```html
{{ $theme := .Theme }}
{{ $tile := .Tile }}

<!-- Terrain name -->
<h5>{{ $theme.GetTerrainName $tile.TileType }}</h5>

<!-- Get effective player for terrain coloring -->
{{ $effectivePlayer := $theme.GetEffectivePlayer $tile.TileType $tile.Player }}
{{ $assetPath := $theme.GetAssetPathForTemplate "tile" $tile.TileType $effectivePlayer }}
<img src="{{ $assetPath }}" alt="{{ $theme.GetTerrainName $tile.TileType }}" />
```

## Using Theme in Renderers

### SVG Renderer

```go
func (r *SVGWorldRenderer) RenderTile(terrainId, playerId int32) (string, error) {
    // Get effective player (city = owner, nature = neutral)
    effectivePlayer := r.theme.GetEffectivePlayer(terrainId, playerId)

    // Get player colors from theme's mapping.json
    colors := r.theme.GetPlayerColor(effectivePlayer)

    // Load and process SVG
    svgPath := r.theme.GetTilePath(terrainId)
    svgContent, err := os.ReadFile(svgPath)
    if err != nil {
        return "", err
    }

    // Apply player colors to SVG gradients
    if colors != nil {
        svgContent = r.applyPlayerColors(svgContent, colors)
    }

    return string(svgContent), nil
}
```

## TypeScript Usage

### Creating Theme

```typescript
import DefaultTheme from './default';
import FantasyTheme from './fantasy';

// Default theme (PNG)
const defaultTheme = new DefaultTheme();

// Fantasy theme (SVG)
const fantasyTheme = new FantasyTheme();
```

### Getting Player Colors

```typescript
const playerId = 1;
const colors = theme.getPlayerColor(playerId);

if (colors) {
    console.log(`Player ${playerId}: ${colors.name}`);
    console.log(`Primary: ${colors.primary}`);
    console.log(`Secondary: ${colors.secondary}`);
}
```

### Rendering with Theme

```typescript
// Get effective player for terrain
const terrainId = 5; // grass
const playerId = 2;

// For city terrain (like base), returns playerId
// For nature terrain (like grass), returns 0 (neutral)
const effectivePlayer = theme.isCityTile(terrainId) ? playerId : 0;

// Get asset path
const path = theme.getTileAssetPath(terrainId, effectivePlayer);
```

## Available Themes

| Theme | Asset Type | Description |
|-------|------------|-------------|
| default | PNG | Original v1 pre-colored assets |
| fantasy | SVG | Medieval units (Peasant, Knight, Castle) |
| modern | SVG | Military units (Infantry, Tank, Base) |

## Theme Registry

```go
// List available themes
themes := themes.GetAvailableThemes()
// ["default", "fantasy", "modern"]

// Create by name
cityTerrains := lib.DefaultRulesEngine().GetCityTerrains()
theme, err := themes.CreateTheme("fantasy", cityTerrains)
```

## Player Colors in mapping.json

Each theme can customize player colors:

```json
{
  "playerColors": {
    "0": { "primary": "#888888", "secondary": "#666666", "name": "Neutral" },
    "1": { "primary": "#f87171", "secondary": "#dc2626", "name": "Red" },
    "2": { "primary": "#60a5fa", "secondary": "#2563eb", "name": "Blue" },
    "3": { "primary": "#4ade80", "secondary": "#16a34a", "name": "Green" },
    "4": { "primary": "#facc15", "secondary": "#ca8a04", "name": "Yellow" },
    "5": { "primary": "#fb923c", "secondary": "#ea580c", "name": "Orange" },
    "6": { "primary": "#c084fc", "secondary": "#9333ea", "name": "Purple" },
    "7": { "primary": "#f472b6", "secondary": "#db2777", "name": "Pink" },
    "8": { "primary": "#22d3ee", "secondary": "#0891b2", "name": "Cyan" }
  }
}
```

## Key Concepts

1. **cityTerrains**: Map of terrain IDs that should show owner colors (bases, cities, etc.)
2. **GetEffectivePlayer**: Returns owner for cities, 0 (neutral) for nature
3. **GetPlayerColor**: Returns color scheme from theme's mapping.json
4. **isCityTile**: TypeScript equivalent for checking if terrain uses player colors
