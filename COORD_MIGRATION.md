# Coordinate System Migration: Row/Col → Cube Coordinates

## Overview

This document tracks the migration from legacy row/col coordinates to proper cube coordinates for hexagonal grid mathematics in WeeWar.

## Motivation

The original system used rectangular row/col coordinates with manual hex offsetting, leading to:
- Complex coordinate conversion bugs
- Inconsistent distance calculations  
- Hardcoded tile dimensions scattered throughout code
- Difficult negative coordinate handling

**Solution**: Implement proper cube coordinates (Q/R) with direct hex-to-pixel conversion using Red Blob Games formulas.

## Technical Details

### Coordinate Systems
- **Old**: Row/Col with manual hex offsets (`row % 2` logic)
- **New**: Cube coordinates (Q, R, S) with `S = -Q - R`
- **Layout**: Fixed odd-r layout (odd rows offset right)
- **Conversion**: Direct hex-to-pixel using pointy-topped hex formulas

### Core Changes

#### 1. Data Structures
```go
// Before
type Unit struct {
    Row int
    Col int
    // ...
}

// After  
type Unit struct {
    Coord CubeCoord `json:"coord"`
    // ...
}
```

#### 2. API Methods
```go
// Before
func (m *Map) TileAt(row, col int) *Tile
func (g *Game) IsValidMove(fromRow, fromCol, toRow, toCol int) bool

// After
func (m *Map) TileAt(coord CubeCoord) *Tile  
func (g *Game) IsValidMove(from, to CubeCoord) bool
```

#### 3. Coordinate Conversion
```go
// Direct hex-to-pixel conversion (no row/col intermediate)
func (m *Map) CenterXYForTile(coord CubeCoord, tileWidth, tileHeight, yIncrement, originX, originY float64) (x, y float64) {
    q := float64(coord.Q)
    r := float64(coord.R)
    
    // Pointy-topped hex formulas
    x = originX + tileWidth * 1.732050808 * (q + r/2.0)  // sqrt(3)
    y = originY + tileWidth * 3.0/2.0 * r
    
    return x, y
}
```

## Migration Progress

### ✅ Completed (Phases 1-4)

#### Phase 1: Foundation
- ✅ Added `CubeDistance()` helper function
- ✅ Verified `HexToDisplay()` and `DisplayToHex()` exist
- ✅ Updated Map bounds to use MinQ/MaxQ/MinR/MaxR

#### Phase 2: Core Data Structures  
- ✅ Unit struct now uses `Coord CubeCoord` instead of Row/Col
- ✅ Removed Row/Col fields from Tile struct
- ✅ Added helper methods for backward compatibility

#### Phase 3: API Methods
- ✅ `TileAt()` now takes CubeCoord (primary method)
- ✅ `IsValidMove()` uses cube coordinates
- ✅ `GetMovementCost()` uses cube coordinates  
- ✅ Proper hex distance calculation with `CubeDistance()`

#### Phase 4: Movement System
- ✅ `MoveUnit()` takes CubeCoord parameter
- ✅ `CanMoveUnit()` uses cube coordinates
- ✅ Unit positioning uses cube coordinates internally

### ✅ Completed (Phase 5)

#### Phase 5: CLI Translation Layer
- ✅ CLI preserves chess notation (A1, B2) for user experience
- ✅ Internal conversion: Chess notation → row/col → cube coordinates
- ✅ Game API calls use cube coordinates
- ✅ Display output converts back to chess notation
- ✅ Added `MoveUnitAt()` and `AttackUnitAt()` methods for coordinate-based actions
- ✅ CLI now acts as thin translation layer with centralized validation in Game object

### ✅ Completed (Phase 6)

#### Phase 6: Game-World-Observer Architecture
- ✅ Implemented proper Game-World-Observer architecture separation
- ✅ Game object now focuses on flow control and game logic
- ✅ World object contains pure state (Map, Units by player)
- ✅ Removed all rendering methods from Game object
- ✅ Updated WorldRenderer to work directly with World data using cube coordinates
- ✅ Eliminated CreateGameForRendering approach (architectural violation)
- ✅ Map now includes OriginX/OriginY for coordinate system origin
- ✅ CenterXYForTile method updated to use Map's internal origin
- ✅ All rendering now uses direct Map.Tiles access for efficiency
- ✅ Preserved asset rendering support in WorldRenderer

### 📋 Remaining (Phases 7-8)

#### Phase 7: Update Remaining Components
- [ ] Update canvas_buffer.go to use cube coordinates
- [ ] Update editor.go to use cube coordinates
- [ ] Update all pixel-to-coordinate conversions

#### Phase 7: Update Tests
- [ ] Test files updated to use cube coordinate APIs
- [ ] CLI tests updated for new coordinate system
- [ ] Rendering tests updated

#### Phase 8: Final Integration
- [ ] Extensive CLI testing with existing save files
- [ ] Performance validation
- [ ] Documentation updates
- [ ] Remove deprecated methods

## Key Design Decisions

### 1. Naming Convention
- **Primary methods**: `TileAt(coord)`, `IsValidMove(from, to)`
- **Legacy methods**: Removed to avoid confusion
- **Helpers**: `HexToDisplay()`, `DisplayToHex()`, `CubeDistance()`

### 2. CLI Abstraction
- Users continue using familiar chess notation (A1, B2)
- CLI handles all coordinate conversions internally
- Zero impact on user experience

### 3. Rendering System
- `CenterXYForTile()` takes cube coordinates + origin
- Origin represents pixel center of Q=0, R=0 tile
- Direct hex-to-pixel conversion (no row/col intermediate)

### 4. Layout Standardization
- Fixed odd-r layout eliminates configuration complexity
- Consistent meaning for negative coordinates
- Simplified coordinate validation

## Testing Strategy

### 1. Unit Tests
- Coordinate conversion accuracy
- Distance calculation correctness
- Bounds validation with negative coordinates

### 2. Integration Tests
- CLI commands work identically to before
- Save/load compatibility preserved
- Rendering output matches expected positions

### 3. Performance Tests
- Coordinate conversion performance
- Memory usage with cube coordinates
- Rendering speed with direct conversion

## Benefits Achieved

### 1. Mathematical Correctness
- Proper hex distance calculations
- Accurate coordinate conversions
- Support for arbitrary map regions

### 2. Code Simplification
- Eliminated hardcoded tile dimensions
- Removed complex offset calculations
- Centralized coordinate logic

### 3. User Experience
- CLI remains unchanged for users
- Consistent coordinate behavior
- Support for negative coordinates

### 4. Maintainability
- Single source of truth for coordinates
- Clear separation of concerns
- Easier to add new features

## Migration Commands

```bash
# Run CLI tests to verify compatibility
go test ./lib -run TestCLI

# Test coordinate conversion
go test ./lib -run TestCoordinateConversion

# Verify rendering output
go test ./lib -run TestRendering

# End-to-end CLI testing
./cmd/weewar-cli/weewar-cli -new -interactive
```

## Next Steps

1. **Complete Phase 6**: Update rendering system to use cube coordinates
2. **Phase 7**: Update all tests to use cube coordinate APIs
3. **Phase 8**: Final integration and testing
4. **Performance**: Benchmark and optimize if needed
5. **Documentation**: Update user guides and API docs

## Recent Progress

### Phase 6 Completion (Game-World-Observer Architecture)
- Successfully implemented proper Game-World-Observer architecture separation
- Game object now focuses purely on flow control and game logic
- World object contains pure state (Map, Units organized by player)
- Removed all rendering methods from Game object to prevent architectural violations
- Updated WorldRenderer to work directly with World data using cube coordinates
- Eliminated CreateGameForRendering approach which violated separation of concerns
- Map now includes OriginX/OriginY fields for coordinate system origin management
- CenterXYForTile method updated to use Map's internal origin automatically
- All rendering now uses direct Map.Tiles access for efficiency (no copying)
- Preserved asset rendering support while maintaining clean architecture

## Notes

- All changes maintain backward compatibility during transition
- Chess notation (A1, B2) interface preserved for CLI users
- Cube coordinates provide foundation for future hex-based features
- Migration can be completed incrementally without breaking existing functionality