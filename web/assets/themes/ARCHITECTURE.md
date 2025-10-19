# Theme Architecture - Go + TypeScript

## Overview

This theme system supports both **Go templates** (server-side rendering) and **TypeScript/Phaser** (client-side rendering) with minimal duplication.

## File Structure

```
web/assets/themes/
├── themes.go              # Go interfaces (Theme, ThemeAssets)
├── base.go                # BaseTheme implementation (loads mapping.json)
├── default.go             # DefaultTheme (PNG-based, like TS default.ts)
├── BaseTheme.ts           # TypeScript base class
├── default.ts             # TypeScript default theme
├── modern.ts              # TypeScript modern theme
├── fantasy.ts             # TypeScript fantasy theme
└── providers/
    └── AssetProvider.ts   # Handles asset loading in Phaser

web/static/assets/themes/
├── fantasy/
│   ├── mapping.json      # ✅ Shared by Go AND TypeScript
│   ├── Units/*.svg
│   └── Tiles/*.svg
└── modern/
    ├── mapping.json      # ✅ Shared by Go AND TypeScript
    ├── Units/*.svg
    └── Tiles/*.svg
```

## Key Design Decisions

### 1. Split Interface: Theme vs ThemeAssets

**Theme** = Lightweight, metadata-only
- Unit/terrain names
- Descriptions
- Path resolution
- Classification (city/nature/bridge)
- ✅ **Use this now** in Go templates

**ThemeAssets** = Heavy, I/O operations
- Loading SVG files
- Applying player colors
- Rendering/rasterization
- ⏳ **Implement later** (optional)

**Why?** Separation of concerns + phased migration. Templates only need metadata and paths, not rendered assets.

### 2. Proto Definitions

All theme structs are defined in `protos/weewar/v1/themes.proto`:
- ✅ Consistent types between Go and TS
- ✅ Future-proof for gRPC services
- ✅ Single source of truth
- ✅ Auto-generated code

### 3. Shared mapping.json

Both Go and TypeScript read the **same** `mapping.json` files:
```json
{
  "themeInfo": {
    "name": "Medieval Fantasy",
    "version": "1.0.0",
    "basePath": "/static/assets/themes/fantasy",
    "assetType": "svg",
    "needsPostProcessing": true
  },
  "units": {
    "1": { "name": "Peasant", "image": "Units/Peasant.svg" }
  },
  "terrains": {
    "1": { "name": "Castle", "image": "Tiles/Castle.svg" }
  }
}
```

**Why?** No duplication, single source of truth for theme metadata.

### 4. Co-located Files

Go and TypeScript theme files live together:
- `default.go` next to `default.ts`
- `modern.go` (future) next to `modern.ts`

**Why?** Easier to find related implementations, clear correspondence.

## Current Implementation (Phase 1)

### What Works Now

1. **Go Templates**: Can use `Theme` interface to get:
   - Unit/terrain names
   - Asset paths (for PNG themes)
   - Descriptions

2. **TypeScript**: Unchanged, continues working as before

3. **Default Theme**: Fully implemented in both Go and TS

### Example Flow

```
User clicks unit
  ↓
Go Presenter receives event
  ↓
Presenter calls SetUnitStats(unit)
  ↓
Template renders with:
    - unit.UnitType = 1
    - theme.GetUnitName(1) = "Infantry"
    - theme.GetUnitAssetPath(1, 2) = "/static/assets/v1/Units/1/2.png"
  ↓
HTML sent to browser
  ↓
Browser loads PNG from path
```

## Future Implementation (Phase 2 - Optional)

### ThemeAssets in Go

If you want server-side SVG rendering:

```go
// Implement ThemeAssets interface
type FantasyThemeAssets struct {
    theme *BaseTheme
}

func (f *FantasyThemeAssets) GetUnitAsset(unitId, playerId int) (*v1.AssetResult, error) {
    // 1. Load SVG template from disk
    svgContent, err := os.ReadFile(f.theme.GetUnitPath(unitId))

    // 2. Apply player colors (replace gradient stops)
    processedSVG, err := f.ApplyPlayerColors(svgContent, playerId)

    // 3. Return inline SVG
    return &v1.AssetResult{
        Type: v1.AssetResult_TYPE_SVG,
        Data: processedSVG,
    }, nil
}
```

Then templates can embed SVG directly:
```html
{{ $asset := .ThemeAssets.GetUnitAsset .Unit.UnitType .Unit.Player }}
{{ if eq $asset.Type "TYPE_SVG" }}
  <div class="unit-icon">{{ $asset.Data }}</div>
{{ end }}
```

**Benefits:**
- No separate HTTP requests for assets
- Smaller initial payload (inline SVG)
- Go-controlled rendering

**Drawbacks:**
- More Go code to maintain
- SVG processing in Go (need XML library)
- May not need this if TS asset loading works well

## Concerns & Trade-offs

### Concern 1: Duplication of Logic

**Issue:** Classification logic (`IsCityTile`, etc.) is in both Go and TS

**Mitigation:**
- Constants are simple arrays, unlikely to change
- Could generate from proto if it becomes a problem
- Small surface area (< 50 lines)

### Concern 2: Mapping.json Sync

**Issue:** If mapping.json changes, both Go and TS might need updates

**Mitigation:**
- JSON is the source of truth
- Both read same file
- Breaking changes are rare (add units, not change structure)

### Concern 3: Asset Loading Split

**Issue:** Go does metadata, TS does asset loading - feels split-brained?

**Counter:** This is actually **good design**
- Go templates only need paths/names
- Asset loading is browser/Phaser concern
- Keeps Go layer thin and fast
- TS continues to do what it does best (DOM, canvas)

### Concern 4: Performance (if implementing ThemeAssets)

**Issue:** Loading SVG files in Go for every template render

**Mitigation:**
- Cache loaded/processed SVGs in memory
- Pre-generate all variants at build time
- Or just stick with Phase 1 (paths only)

### Concern 5: SVG Processing in Go

**Issue:** Applying player colors requires XML parsing

**Solutions:**
1. Use `encoding/xml` (standard library)
2. String replacement (fragile but fast)
3. Don't implement ThemeAssets - keep using TS
4. Pre-generate at build time

## Recommended Approach

### For Now (Minimal Work)

1. ✅ Use `Theme` interface in Go templates
2. ✅ Keep asset loading in TypeScript
3. ✅ Templates render paths: `<img src="{{ .Theme.GetAssetPath ... }}">`
4. ✅ Browser loads assets as before

### Later (If Needed)

1. Implement `ThemeAssets` for SVG themes
2. Add caching layer
3. Optionally pre-generate SVGs at build time

### Never (Probably)

- Don't port DOM manipulation to Go (`setUnitImage`)
- Don't port Phaser integration to Go
- Don't duplicate AssetProvider in Go

## Migration Path

1. **Now**: Add `Theme` field to presenters
2. **Now**: Update templates to use `theme.GetUnitName()`, `theme.GetAssetPath()`
3. **Later**: Enhance `mapping.json` with `themeInfo` block
4. **Later**: Create `modern.go`, `fantasy.go` using `BaseTheme`
5. **Optional**: Implement `ThemeAssets` for server-side rendering

## Summary

**This design lets you:**
- ✅ Use theme metadata in Go templates **today**
- ✅ Keep existing TypeScript asset system working
- ✅ Share theme data via `mapping.json`
- ✅ Gradually migrate if needed
- ✅ Avoid big rewrites

**Key insight:** Go templates need **metadata** (names, paths), not **rendered assets**. Separating `Theme` from `ThemeAssets` makes this clean and practical.
