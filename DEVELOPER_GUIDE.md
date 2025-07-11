# WeeWar Developer Guide

A comprehensive guide for developing, building, testing, and running the WeeWar turn-based strategy game built on the TurnEngine framework.

## Table of Contents
- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Building](#building)
- [Testing](#testing)
- [Development Workflow](#development-workflow)
- [Architecture Overview](#architecture-overview)
- [Common Tasks](#common-tasks)
- [Troubleshooting](#troubleshooting)
- [Known Issues](#known-issues)

## Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd turnengine

# Install dependencies
go mod download

# Test map system
go run games/weewar/cmd/test-map-system/main.go

# Run unit tests
go test ./...

# Build WASM version
make wasm
```

## Prerequisites

### Required Software
- **Go 1.24.0+** - Core development language
- **Python 3.8+** - For map processing tools
- **Git** - Version control
- **Make** - Build automation

### Optional Tools
- **Visual Studio Code** - Recommended IDE with Go extension
- **Docker** - For containerized development
- **Node.js** - For web development (if extending web interface)

## Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd turnengine
```

### 2. Install Go Dependencies
```bash
go mod download
go mod verify
```

### 3. Set up Python Environment (for map tools)
```bash
cd games/weewar/maps
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### 4. Verify Installation
```bash
# Test Go build
go build ./...

# Test map system
go run games/weewar/cmd/test-map-system/main.go

# Test Python tools
cd games/weewar/maps
python grid_analyzer.py --help
```

## Building

### Local Development Build
```bash
# Build all components
make

# Build server only
make server

# Build WebAssembly
make wasm
```

### Available Build Targets

| Target | Command | Description |
|--------|---------|-------------|
| `server` | `make server` | Build server binary |
| `wasm` | `make wasm` | Build WebAssembly version |
| `dev` | `make dev` | Build WASM and run dev server |
| `test` | `make test` | Run all tests |

### Manual Build Commands
```bash
# Build server
go build -o bin/server cmd/server/main.go

# Build WASM
GOOS=js GOARCH=wasm go build -o web/wasm/game-engine.wasm cmd/wasm/main.go

# Build WeeWar tools
go build -o bin/test-map-system games/weewar/cmd/test-map-system/main.go
go build -o bin/extract-data games/weewar/cmd/extract-data/main.go
go build -o bin/extract-map-data games/weewar/cmd/extract-map-data/main.go
```

## Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/turnengine/...
```

### Integration Tests
```bash
# Test map system
go run games/weewar/cmd/test-map-system/main.go

# Test data extraction
go run games/weewar/cmd/extract-data/main.go

# Test map data extraction
go run games/weewar/cmd/extract-map-data/main.go
```

### Manual Testing
```bash
# Create a simple test game
cd games/weewar
go run -c '
package main

import (
    "fmt"
    "log"
    "github.com/panyam/turnengine/games/weewar"
)

func main() {
    game, err := weewar.CreateWeeWarGameWithMapName("Small World")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Game created successfully with %d players\n", len(game.GetGameState().Players))
}'
```

### Python Map Tools Testing
```bash
cd games/weewar/maps

# Test grid analysis
python grid_analyzer.py --image ../data/Maps/1_files/map-og.png --debug

# Test hex generation
python hex_generator.py --image ../data/Maps/1_files/map-og.png --debug

# Test tile extraction
python hex_splitter.py --image ../data/Maps/1_files/map-og.png --output-dir test_tiles
```

## Development Workflow

### Code Organization

```
turnengine/
├── internal/turnengine/     # Core game engine (80% reusable)
│   ├── entity.go           # Entity Component System
│   ├── gamestate.go        # Game state management
│   ├── board.go            # Abstract board system
│   └── ...
├── games/weewar/           # WeeWar implementation (20% specific)
│   ├── core.go             # Core game API (NEW)
│   ├── buffer.go           # Rendering system (NEW)
│   ├── game.go             # Main game logic
│   ├── board.go            # Hex board implementation
│   ├── components.go       # WeeWar components
│   ├── combat.go           # Combat system
│   ├── movement.go         # Movement system
│   ├── map.go              # Map system
│   └── data/               # Game data
└── cmd/                    # Main applications
    ├── server/             # Server application
    └── wasm/               # WebAssembly application
```

### New Development Architecture (v2.0)

#### Core API Design
- **Clean Separation**: Static data (`UnitData`, `TerrainData`) vs runtime state (`Unit`, `Tile`, `Game`)
- **Programmatic Interface**: Direct object manipulation instead of ECS lookups
- **Headless Gameplay**: Easy testing and AI development
- **Deterministic Games**: Seeded RNG for reproducible gameplay

#### Buffer-Based Rendering
- **Composable Layers**: Separate terrain, units, and UI rendering
- **Scaling Support**: Professional image scaling with bilinear interpolation
- **Alpha Compositing**: Proper transparency and blending
- **Multi-format Output**: PNG generation with flexible dimensions

### Development Principles

1. **Clean Architecture**: Separate concerns clearly (static vs runtime, rendering vs logic)
2. **Programmatic API**: Easy to test and extend programmatically
3. **Composable Systems**: Modular design with clear interfaces
4. **Test-First Development**: Write tests before implementing features
5. **Visual Debugging**: Use Buffer system for debugging and visualization

### Adding New Features

#### 1. Core Game Features
```go
// Example: Adding a new unit type
func TestNewUnitType(t *testing.T) {
    // Create unit
    unit := NewUnit(ENGINEER_TYPE, 0)
    unit.Row = 5
    unit.Col = 3
    
    // Test unit behavior
    game.AddUnit(unit, 0)
    
    // Verify placement
    tile := game.Map.TileAt(5, 3)
    assert.Equal(t, tile.Unit, unit)
}
```

#### 2. Rendering Features
```go
// Example: Adding visual effects
func (g *Game) RenderEffects(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
    // Create effect image
    explosionImg := createExplosionSprite()
    
    // Draw at combat locations
    for _, combat := range g.GetCombatEffects() {
        x, y := g.Map.XYForTile(combat.Row, combat.Col, tileWidth, tileHeight, yIncrement)
        buffer.DrawImage(x, y, tileWidth, tileHeight, explosionImg)
    }
}
```

#### 3. Integration Testing
```go
// Example: Full gameplay test
func TestCompleteGame(t *testing.T) {
    // Create game
    game := createTestGame()
    
    // Play through several turns
    for turn := 0; turn < 10; turn++ {
        // Move units
        moveUnits(game)
        
        // Render state
        buffer := NewBuffer(800, 600)
        game.RenderToBuffer(buffer, 80, 70, 50)
        buffer.Save(fmt.Sprintf("/tmp/game_turn_%d.png", turn))
        
        // Next turn
        game.NextTurn()
    }
}

### Debugging

#### Command Line Debugging
```bash
# Enable debug logging
export DEBUG=1
go run games/weewar/cmd/test-map-system/main.go

# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug games/weewar/cmd/test-map-system/main.go

# Print game state
fmt.Printf("Game State: %+v\n", game.GetGameState())
```

#### Visual Debugging (NEW)
```go
// Create debug visualization
func debugGameState(game *Game) {
    buffer := NewBuffer(800, 600)
    
    // Render layers separately for debugging
    terrainBuffer := NewBuffer(800, 600)
    unitBuffer := NewBuffer(800, 600)
    
    game.RenderTerrain(terrainBuffer, 80, 70, 50)
    game.RenderUnits(unitBuffer, 80, 70, 50)
    
    // Save individual layers
    terrainBuffer.Save("/tmp/debug_terrain.png")
    unitBuffer.Save("/tmp/debug_units.png")
    
    // Save composite
    buffer.RenderBuffer(terrainBuffer)
    buffer.RenderBuffer(unitBuffer)
    buffer.Save("/tmp/debug_complete.png")
}
```

#### Unit Testing with Visuals
```go
func TestUnitMovement(t *testing.T) {
    // Create test scenario
    game := createTestGame()
    
    // Move unit
    unit := game.GetUnitsForPlayer(0)[0]
    unit.Row = 3
    unit.Col = 4
    
    // Visualize result
    buffer := NewBuffer(400, 300)
    game.RenderToBuffer(buffer, 60, 50, 40)
    buffer.Save("/tmp/test_movement.png")
    
    // Verify movement
    tile := game.Map.TileAt(3, 4)
    assert.Equal(t, tile.Unit, unit)
}
```

## Architecture Overview

### TurnEngine Framework (Reusable)
- **Entity Component System**: Flexible game object composition
- **Game State Management**: Version control, player management, turn handling
- **Abstract Board System**: Supports hex, grid, graph coordinates
- **Command Processing**: Validation and execution pipeline
- **Serialization**: JSON-based state persistence

### WeeWar Implementation (Game-Specific)
- **Hex Board**: Axial coordinate system with pathfinding
- **Components**: Health, Movement, Combat, Team, UnitType, Position
- **Systems**: Combat calculations, movement validation, turn processing
- **Data Integration**: Real WeeWar maps, units, and combat matrices

### Key Interfaces

```go
// Core abstractions
type Position interface {
    GetCoordinates() (int, int, int)
    Distance(Position) float64
}

type Board interface {
    IsValidPosition(Position) bool
    GetNeighbors(Position) []Position
    FindPath(from, to Position) []Position
}

type Component interface {
    Type() string
}

type System interface {
    Update(*GameState) error
}
```

### Data Flow

1. **Game Creation**: Load map data → Create game config → Initialize game
2. **Command Processing**: Validate command → Process command → Update state
3. **State Management**: Track entities → Update components → Persist changes
4. **Turn Management**: Process player actions → Update game state → Check victory

## Common Tasks

### Adding a New Map

1. **Extract Map Data**: Use Python tools to analyze map image
2. **Update Maps JSON**: Add map data to `weewar-maps.json`
3. **Test Loading**: Verify map loads correctly
4. **Validate Gameplay**: Ensure proper unit placement and terrain

```bash
# Extract map data
cd games/weewar/maps
python grid_analyzer.py --image new_map.png --debug
python hex_splitter.py --image new_map.png --output-dir new_map_tiles

# Update maps JSON file
# Add new map entry to games/weewar/data/weewar-maps.json

# Test loading
go run games/weewar/cmd/test-map-system/main.go
```

### Creating New Unit Types

1. **Update Data**: Add unit to `weewar-data.json`
2. **Extend Components**: Add new component types if needed
3. **Update Systems**: Modify combat/movement systems
4. **Test Integration**: Verify unit works in game

```go
// Add new component type
type SpecialAbilityComponent struct {
    AbilityType string `json:"abilityType"`
    Cooldown    int    `json:"cooldown"`
    Range       int    `json:"range"`
}

func (sac SpecialAbilityComponent) Type() string { return "specialAbility" }
```

### Debugging Game State

```go
// Print entity components
for _, entity := range game.GetGameState().World.GetAllEntities() {
    fmt.Printf("Entity %s:\n", entity.ID)
    for compType, comp := range entity.Components {
        fmt.Printf("  %s: %+v\n", compType, comp)
    }
}

// Print board state
board := game.GetBoard()
for q := 0; q < board.Width; q++ {
    for r := 0; r < board.Height; r++ {
        pos := &weewar.HexPosition{Q: q, R: r}
        if entityID, exists := board.GetEntityAt(pos); exists {
            fmt.Printf("Position (%d,%d): %s\n", q, r, entityID)
        }
    }
}
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile cpu.prof -bench=. ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile mem.prof -bench=. ./...
go tool pprof mem.prof

# Trace execution
go test -trace trace.out ./...
go tool trace trace.out
```

## Troubleshooting

### Common Build Errors

**Error**: `package github.com/panyam/turnengine/games/weewar: cannot find package`
**Solution**: Ensure you're in the correct directory and Go module is initialized
```bash
go mod init github.com/panyam/turnengine
go mod tidy
```

**Error**: `undefined: HexPosition`
**Solution**: Import the correct package
```go
import "github.com/panyam/turnengine/games/weewar"
```

### Runtime Issues

**Error**: Map loading fails with "file not found"
**Solution**: Ensure data files are in correct location
```bash
ls games/weewar/data/
# Should show: weewar-data.json, weewar-maps.json, Maps/
```

**Error**: Position validation fails
**Solution**: Check hex coordinate bounds
```go
if !board.IsValidPosition(pos) {
    fmt.Printf("Invalid position: %+v\n", pos)
}
```

### Python Tool Issues

**Error**: `ModuleNotFoundError: No module named 'cv2'`
**Solution**: Install OpenCV
```bash
pip install opencv-python
```

**Error**: Map analysis fails
**Solution**: Check image file format and path
```bash
file your_map.png  # Should show PNG image data
```

### Performance Issues

**Issue**: Slow pathfinding
**Solution**: Implement A* caching
```go
// Cache pathfinding results
type PathCache struct {
    cache map[string][]Position
}
```

**Issue**: Memory leaks
**Solution**: Proper entity cleanup
```go
// Remove entities properly
world.RemoveEntity(entityID)
```

## Known Issues

### Current Status (v2.0)

The core architecture has been significantly improved with the new Buffer-based rendering system and clean API design. Most critical blocking issues have been resolved.

### Minor Issues

1. **AI System**: Framework exists but no AI implementation
2. **Web Interface**: WASM builds but no interactive UI
3. **Advanced Features**: Game persistence, tournaments, etc.

### Recent Fixes ✅

1. **~~Board Position Validation~~**: Fixed with new core API
2. **~~Command Processing~~**: Resolved with simplified game state management
3. **~~Unit Placement~~**: Fixed with direct object manipulation
4. **~~Map Terrain~~**: Implemented with proper terrain rendering
5. **~~Victory Conditions~~**: Can be implemented with current architecture

### Development Roadmap

1. **Phase 1**: AI implementation and game persistence
2. **Phase 2**: Web interface and real-time features
3. **Phase 3**: Advanced features and community tools
4. **Phase 4**: Performance optimization and scaling

## Contributing

### Development Setup
1. Fork the repository
2. Create feature branch: `git checkout -b feature/new-feature`
3. Make changes with tests
4. Run full test suite: `make test`
5. Submit pull request

### Code Style
- Follow Go conventions: `gofmt`, `go vet`, `golint`
- Use meaningful variable names
- Add comments for complex logic
- Write tests for new features

### Documentation
- Update this guide for new features
- Add inline documentation for complex functions
- Include examples in docstrings

## Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [WebAssembly with Go](https://github.com/golang/go/wiki/WebAssembly)
- [Entity Component System](https://en.wikipedia.org/wiki/Entity_component_system)

### Tools
- [Delve Debugger](https://github.com/go-delve/delve)
- [Go Profiler](https://golang.org/pkg/runtime/pprof/)
- [Visual Studio Code Go Extension](https://github.com/golang/vscode-go)

### Game Development
- [Original WeeWar](http://weewar.com/) - Reference implementation
- [Hex Grid Guide](https://www.redblobgames.com/grids/hexagons/)
- [Turn-Based Game Design](https://gamedevelopment.tutsplus.com/articles/turn-based-game-mechanics--gamedev-11175)

---

**Last Updated**: 2025-01-11
**Version**: 2.0.0
**Status**: WeeWar core architecture complete - enhanced API, advanced rendering, comprehensive testing