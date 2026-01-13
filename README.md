# LilBattle - Turn-Based Hexagonal Strategy Game

A modern turn-based strategy game inspired by classic hex-grid wargames, built with Go backend, TypeScript/Phaser.js frontend, and WebAssembly for game logic.

## What is LilBattle?

LilBattle is a multiplayer turn-based strategy game featuring:
- **Hexagonal Grid Combat** - Tactical warfare on hex-based maps
- **Unit-Based Strategy** - Infantry, tanks, aircraft, and more with unique capabilities
- **Formula-Based Combat** - Probabilistic damage system with wound bonuses and counter-attacks
- **Browser-Based Gameplay** - No installation required, play in your browser
- **Map Editor** - Create custom maps with Phaser.js-based visual editor
- **CLI Support** - Play via command line with chess notation (A1, B2, etc.)

## Architecture Overview

LilBattle follows a clean, modern architecture:

```
┌────────────────────────────────────────────────────────────┐
│                         Browser                            │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────────┐   │
│  │   Phaser.js  │  │ TypeScript  │  │  WASM (Go)       │   │
│  │   Renderer   │←─│   UI Layer  │←─│  Game Logic      │   │
│  └──────────────┘  └─────────────┘  └──────────────────┘   │
└────────────────────────────────────────────────────────────┘
                            ↕ gRPC
┌────────────────────────────────────────────────────────────┐
│                       Go Backend                           │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────────┐   │
│  │  Web Server  │  │   Services  │  │  Rules Engine    │   │
│  │  (Templar)   │←─│   (gRPC)    │←─│  (Data-Driven)   │   │
│  └──────────────┘  └─────────────┘  └──────────────────┘   │
└────────────────────────────────────────────────────────────┘
```

### Core Components

- **Backend (`services/`)** - Go-based game engine with rules-driven mechanics
- **Frontend (`web/`)** - TypeScript + Phaser.js for rendering and UI
- **WASM Bridge** - Go compiled to WebAssembly for client-side game logic
- **Protos (`protos/`)** - Protocol Buffers define all data structures and APIs
- **CLI (`cmd/cli/`)** - Command-line interface for headless gameplay

## Key Technologies

- **Backend**: Go 1.24+, gRPC, Protocol Buffers
- **Frontend**: TypeScript, Phaser.js 3, Tailwind CSS
- **Templates**: Go html/template with Templar engine (namespace, include, extend directives)
- **Build**: esbuild for TypeScript, buf for protos
- **Live Reload**: devloop for continuous builds
- **Template Library**: goapplib for shared page components

## Getting Started

### Prerequisites

- Go 1.24 or later
- Node.js 18+ and npm
- buf CLI for protobuf generation
- devloop for live reloading (optional but recommended)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd lilbattle
   ```

2. **Install dependencies**
   ```bash
   # Go dependencies
   go mod download

   # Node dependencies for frontend
   cd web && npm install && cd ..
   ```

3. **Generate proto code**
   ```bash
   buf generate
   ```

4. **Start development servers**
   ```bash
   # Start devloop for continuous builds (recommended)
   devloop

   # Or manually:
   # Terminal 1: Backend
   go run main.go serve

   # Terminal 2: Frontend build
   cd web && npm run watch
   ```

5. **Open browser**
   Navigate to `http://localhost:8080`

### CLI Gameplay

Build and use the CLI for command-line gameplay:

```bash
# Build CLI
make cli

# Start a game
ww status

# View units
ww units

# Get available moves for a unit
ww options A1

# Move a unit
ww move A1 R    # Move right
ww move A1 0,1  # Move to coordinates

# Attack
ww attack A1 B2

# End turn
ww endturn
```

## Project Structure

```
lilbattle/
├── cmd/
│   └── cli/              # Command-line interface
├── services/             # Go backend services
│   ├── game.go          # Core game logic
│   ├── moves.go         # Move processing
│   ├── rules_engine.go  # Data-driven game mechanics
│   └── *_test.go        # Test files
├── protos/              # Protocol Buffer definitions
│   └── lilbattle/v1/
│       └── models.proto # Core data structures
├── web/
│   ├── src/             # TypeScript source
│   │   ├── pages/       # Page-specific modules
│   │   └── lib/         # Shared utilities
│   ├── templates/       # Go HTML templates
│   └── frontend/        # Web server
├── data/
│   ├── Tiles/           # Terrain definitions (JSON)
│   └── Units/           # Unit definitions (JSON)
└── docs/                # Documentation
```

## Key Features

### Combat System
- **Formula-Based Damage**: Probabilistic combat with dice rolling simulation
- **Wound Bonus**: Accumulating damage from multiple attackers in same turn
- **Counter-Attacks**: Defenders can strike back if in range
- **Splash Damage**: Area effect damage from certain unit types
- **Terrain Effects**: Defense bonuses from forests, mountains, etc.

### Movement System
- **Hexagonal Grid**: Proper axial coordinate system
- **Terrain Costs**: Different terrain types affect movement
- **Pathfinding**: Dijkstra's algorithm for optimal paths
- **Line of Sight**: Range-based visibility and attack capabilities

### Action Progression
- **Index-Based State Machine**: Units track progression through action sequence
- **Configurable Actions**: Define allowed action orders per unit type
- **Pipe-Separated Alternatives**: Mutually exclusive choices (e.g., "attack|capture")
- **Natural Limiting**: Movement limited by points, attacks by limits

### Map Editor
- **Visual Terrain Editing**: Phaser.js-based canvas with real-time updates
- **Brush Tools**: Paint, erase, eyedropper modes
- **Grid Visualization**: Toggle grid lines and coordinate labels
- **Undo/Redo**: Full history management
- **Export**: Save maps to JSON format

## Development

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./services/ -v

# With coverage
go test ./services/ -cover

# Specific test
go test ./services/ -run TestActionProgression -v
```

### Building

```bash
# Backend binary
go build -o lilbattle main.go

# CLI tool
make cli

# WASM module
cd cmd/wasm && GOOS=js GOARCH=wasm go build -o ../../web/public/wasm/game.wasm

# Frontend assets
cd web && npm run build
```

### Code Generation

```bash
# Regenerate protos
buf generate

# Update rules data
# Edit JSON files in data/Tiles/ and data/Units/
# Changes detected automatically by devloop
```

## Documentation

- **[Architecture](./docs/ARCHITECTURE.md)** - Detailed technical architecture
- **[Developer Guide](./docs/DEVELOPER_GUIDE.md)** - Development workflows and patterns
- **[Roadmap](./docs/ROADMAP.md)** - Development phases and planned features
- **[Project Summary](./PROJECT.md)** - Current status and recent achievements
- **[Attack System](./docs/ATTACK.md)** - Combat mechanics documentation
- **[Game Log](./docs/GAMELOG.md)** - Move history and replay system

## Current Status

**Phase 1 (Core Gameplay)**: ✅ Complete
- Hex grid system, movement, combat, rules engine, CLI

**Phase 2 (Multiplayer)**: 🔄 90% Complete
- Coordination protocol implemented, needs testing

**Phase 3 (Polish & Production)**: 🚧 In Progress
- UI improvements, animations, templar template migration complete

See [docs/ROADMAP.md](./docs/ROADMAP.md) for detailed development phases.

## Contributing

This project follows the conventions in [CLAUDE.md](./CLAUDE.md) for AI-assisted development.

Key principles:
- Proto files are source of truth for all data structures
- Rules engine drives game mechanics (data over code)
- Test coverage for all game logic
- Clean separation: Go backend, TS frontend, WASM bridge

## License

[Add license information]

## Credits

Inspired by the original LilBattle game and classic turn-based strategy games.
