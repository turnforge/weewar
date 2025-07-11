# WeeWar Developer Guide

A comprehensive guide for developing, testing, and running the WeeWar turn-based strategy game.

## Table of Contents
- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)
- [Testing Strategy](#testing-strategy)
- [Development Workflow](#development-workflow)
- [CLI Interface](#cli-interface)
- [Common Tasks](#common-tasks)
- [Troubleshooting](#troubleshooting)

## Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd turnengine/games/weewar

# Install dependencies
go mod download

# Run all tests
go test -v ./...

# Build CLI executable
go build -o /tmp/weewar-cli ./cmd/weewar-cli

# Start interactive game
/tmp/weewar-cli -new -interactive
```

## Architecture Overview

### Current Architecture (2024)

The WeeWar implementation has evolved into a unified, interface-driven architecture:

```
Core Game System (GameInterface)
├── GameController (lifecycle, turns, save/load)
├── MapInterface (hex grid, pathfinding, coordinates)  
└── UnitInterface (units, combat, actions)
     ↓
Unified Game Implementation
├── Comprehensive state management
├── Integrated hex pathfinding
├── Real WeeWar data integration
├── PNG rendering capabilities
├── Asset management system
└── Combat prediction system
     ↓
Multiple Interfaces
├── CLI (REPL with chess notation)
├── PNG Renderer (hex graphics)
└── Web Interface (future)
```

### Key Design Principles

1. **Interface-Driven Design**: Clean contracts for all operations
2. **Unified State Management**: Single source of truth in Game struct
3. **Data-Driven Authenticity**: Real WeeWar data integration
4. **Comprehensive Testing**: All major functionality tested
5. **Multiple Interface Support**: CLI, PNG, Web (future)

## Testing Strategy

### Test Categories

#### 1. Core Game Tests (`*_test.go`)
- **Basic Operations**: Game creation, state management
- **Combat System**: Damage calculations, unit interactions
- **Map Navigation**: Hex pathfinding, coordinate conversion
- **Save/Load**: Game persistence and restoration

#### 2. Interface Tests
- **CLI Tests**: Command parsing, REPL functionality
- **PNG Rendering**: Visual output generation
- **Integration Tests**: Full game scenarios

#### 3. Data Integration Tests
- **Real Data**: Unit stats, terrain, combat matrices
- **Map Loading**: WeeWar map configurations
- **Position Handling**: Chess notation (A1, B2, etc.)

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific test categories
go test -v -run TestGame          # Core game tests
go test -v -run TestCLI           # CLI interface tests
go test -v -run TestCombat        # Combat system tests
go test -v -run TestMap           # Map and pathfinding tests
go test -v -run TestPNG           # PNG rendering tests

# Run with coverage
go test -cover ./...

# Run with verbose output and save test images
go test -v -run TestPNGRendering
# Test images saved to /tmp/turnengine/test/
```

### Test Organization

```go
// Core game functionality
func TestGameBasicOperations(t *testing.T)
func TestCombatSystem(t *testing.T)
func TestMapNavigation(t *testing.T)
func TestSaveLoad(t *testing.T)

// Interface functionality
func TestCLIBasicOperations(t *testing.T)
func TestCLIREPLCommands(t *testing.T)
func TestCLIGameStateIntegration(t *testing.T)
func TestPNGRendering(t *testing.T)

// Integration tests
func TestFullGameplayScenario(t *testing.T)
func TestRealDataIntegration(t *testing.T)
```

## Development Workflow

### Code Organization

```
games/weewar/
├── game_interface.go           # Core interface contracts
├── game.go                     # Unified game implementation
├── map.go, tile.go            # Hex map system
├── unit.go, combat.go         # Unit management and combat
├── assets.go                  # Asset management system
├── predict.go                 # Combat prediction system
├── rendering.go, buffer.go    # PNG generation
├── cli_impl.go                # CLI interface implementation
├── cli_formatter.go           # CLI text formatting
├── cli_test.go                # CLI tests
├── game_test.go               # Core game tests
├── *_test.go                  # Other test files
└── cmd/
    └── weewar-cli/main.go     # CLI executable
```

### Development Process

1. **Interface First**: Define interfaces before implementation
2. **Test-Driven**: Write tests before implementing features
3. **Visual Debugging**: Use PNG rendering for game state visualization
4. **Comprehensive Testing**: Test all major functionality
5. **Documentation**: Update guides and architecture docs

## CLI Interface

### REPL Features

The CLI provides a sophisticated Read-Eval-Print Loop (REPL) for interactive gameplay:

```bash
# Start interactive session
/tmp/weewar-cli -new -interactive

# REPL provides:
weewar[T1:P0]> actions        # Show available actions
weewar[T1:P0]> move B2 B3     # Move unit using chess notation
weewar[T1:P0]> s              # Quick status (shortcut)
weewar[T1:P0]> map            # Display game map
weewar[T1:P0]> end            # End turn
weewar[T2:P1]> quit           # Exit game
```

### REPL Commands

| Command | Description | Example |
|---------|-------------|---------|
| `actions` | Show available actions | `actions` |
| `move <from> <to>` | Move unit | `move A1 B2` |
| `attack <from> <to>` | Attack unit | `attack A1 B2` |
| `s` / `state` | Quick status | `s` |
| `map` | Display map | `map` |
| `units` | Show units | `units` |
| `turn` | Turn information | `turn` |
| `predict <from> <to>` | Damage prediction | `predict A1 B2` |
| `attackoptions <unit>` | Show attack targets | `attackoptions A1` |
| `moveoptions <unit>` | Show movement options | `moveoptions A1` |
| `end` | End turn | `end` |
| `save <file>` | Save game | `save game.json` |
| `render <file>` | Render PNG | `render game.png` |
| `help` | Show help | `help move` |
| `quit` | Exit | `quit` |

### CLI Modes

```bash
# Interactive REPL
/tmp/weewar-cli -new -interactive

# Single commands
/tmp/weewar-cli -new status map

# Batch processing
/tmp/weewar-cli -new -batch commands.txt

# Save and render
/tmp/weewar-cli -new -save game.json -render game.png
```

## Common Tasks

### Adding New Tests

```go
// Example: Adding a new combat test
func TestNewCombatFeature(t *testing.T) {
    // Create test game
    testMap := NewMap(8, 12, false)
    for row := 0; row < 8; row++ {
        for col := 0; col < 12; col++ {
            tile := NewTile(row, col, 1)
            testMap.AddTile(tile)
        }
    }
    testMap.ConnectHexNeighbors()

    game, err := NewGame(2, testMap, 12345)
    require.NoError(t, err)

    // Test specific combat scenario
    // ... test implementation
    
    // Optional: Generate visual output
    if testing.Verbose() {
        buffer := NewBuffer(400, 300)
        game.RenderToBuffer(buffer, 60, 50, 40)
        buffer.Save("/tmp/test_combat_feature.png")
    }
}
```

### Debugging Game State

```go
// Visual debugging
func debugGameState(game *Game) {
    buffer := NewBuffer(800, 600)
    game.RenderToBuffer(buffer, 80, 70, 50)
    buffer.Save("/tmp/debug_state.png")
}

// CLI debugging
func debugCLI(game *Game) {
    cli := NewWeeWarCLI(game)
    cli.SetVerbose(true)
    
    // Print detailed state
    cli.PrintGameState()
    cli.PrintUnits()
    cli.PrintMap()
}
```

### Performance Testing

```bash
# Benchmark tests
go test -bench=. ./...

# Memory profiling
go test -memprofile mem.prof -bench=. ./...
go tool pprof mem.prof

# CPU profiling
go test -cpuprofile cpu.prof -bench=. ./...
go tool pprof cpu.prof
```

### Adding New Features

1. **Define Interface**: Add methods to appropriate interface
2. **Implement Method**: Add implementation to Game struct
3. **Write Tests**: Create comprehensive test coverage
4. **Update CLI**: Add CLI commands if needed
5. **Update Documentation**: Update guides and help text

## Troubleshooting

### Common Issues

**Build Errors**:
```bash
# Missing dependencies
go mod download
go mod tidy

# Import issues
go mod verify
```

**Test Failures**:
```bash
# Run specific failing test
go test -v -run TestSpecificFunction

# Check test output directories
ls /tmp/turnengine/test/
ls /tmp/turnengine/cli_test/
```

**CLI Issues**:
```bash
# Rebuild CLI
go build -o /tmp/weewar-cli ./cmd/weewar-cli

# Test CLI help
/tmp/weewar-cli --help

# Test CLI commands
echo "new" | /tmp/weewar-cli -interactive
```

### Debug Logging

```go
// Add debug output to tests
if testing.Verbose() {
    fmt.Printf("Debug: %+v\n", gameState)
}

// Use t.Logf for test-specific logging
t.Logf("Game state: %+v", game.GetGameState())
```

### Visual Debug Output

```go
// Generate debug visuals
func TestWithVisualDebug(t *testing.T) {
    game := createTestGame()
    
    // Save initial state
    buffer := NewBuffer(400, 300)
    game.RenderToBuffer(buffer, 60, 50, 40)
    buffer.Save("/tmp/debug_initial.png")
    
    // Perform operations
    // ... test operations
    
    // Save final state
    game.RenderToBuffer(buffer, 60, 50, 40)
    buffer.Save("/tmp/debug_final.png")
    
    t.Logf("Debug images saved to /tmp/debug_*.png")
}
```

## File Structure

```
games/weewar/
├── Core Implementation
│   ├── game_interface.go      # Interface contracts
│   ├── game.go               # Unified game implementation
│   ├── map.go, tile.go       # Hex map system
│   ├── unit.go, combat.go    # Unit and combat systems
│   └── rendering.go, buffer.go # PNG rendering
├── CLI Interface
│   ├── cli_impl.go           # CLI implementation
│   ├── cli_formatter.go      # Text formatting
│   └── cmd/weewar-cli/       # CLI executable
├── Testing
│   ├── game_test.go          # Core game tests
│   ├── cli_test.go           # CLI tests
│   ├── combat_test.go        # Combat tests
│   └── *_test.go             # Other test files
├── Data Integration
│   ├── weewar_data.go        # Real WeeWar data
│   └── cmd/extract-data/     # Data extraction tools
└── Documentation
    ├── ARCHITECTURE.md       # Architecture overview
    ├── DEVELOPER_GUIDE.md    # This file
    └── cmd/weewar-cli/USER_GUIDE.md # CLI user guide
```

## Contributing

### Development Setup
1. Fork the repository
2. Create feature branch: `git checkout -b feature/new-feature`
3. Write tests for new functionality
4. Implement feature with comprehensive testing
5. Run full test suite: `go test -v ./...`
6. Update documentation as needed
7. Submit pull request

### Code Standards
- Follow Go conventions (`gofmt`, `go vet`)
- Write comprehensive tests for all new features
- Use meaningful variable names and add comments
- Maintain interface compatibility
- Update documentation for user-facing changes

## Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [WeeWar Architecture](ARCHITECTURE.md)
- [CLI User Guide](cmd/weewar-cli/USER_GUIDE.md)

### Development Tools
- [Delve Debugger](https://github.com/go-delve/delve)
- [Visual Studio Code Go Extension](https://github.com/golang/vscode-go)
- [Go Profiler](https://golang.org/pkg/runtime/pprof/)

### Game Development
- [Hex Grid Guide](https://www.redblobgames.com/grids/hexagons/)
- [Turn-Based Game Design](https://gamedevelopment.tutsplus.com/articles/turn-based-game-mechanics--gamedev-11175)

---

**Last Updated**: 2025-01-11  
**Version**: 3.0.0  
**Status**: Production-ready with comprehensive CLI REPL interface