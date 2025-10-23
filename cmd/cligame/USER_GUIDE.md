# WeeWar CLI User Guide

A comprehensive guide for playing WeeWar using the simplified command-line interface (CLI).

## Table of Contents
- [Getting Started](#getting-started)
- [REPL Interface](#repl-interface)
- [Game Commands](#game-commands)
- [Position System](#position-system)
- [Gameplay Tutorial](#gameplay-tutorial)
- [Advanced Features](#advanced-features)
- [Tips and Tricks](#tips-and-tricks)

## Getting Started

### Installation
```bash
# Build the CLI executable
go build -o weewar-cli ./cmd/weewar-cli

# Make it executable 
chmod +x weewar-cli
```

### Quick Start
```bash
# Start interactive game with a world from storage
./weewar-cli -world $WEEWAR_DATA_ROOT/storage/maps/small-world -interactive

# Load a saved game
./weewar-cli -load my_game.json -interactive

# Execute single commands
./weewar-cli -world $WEEWAR_DATA_ROOT/storage/maps/small-world status units quit
```

## REPL Interface

### Understanding the Prompt
The CLI uses a simple interactive prompt:
```
> 
```

### Getting Help
```bash
# Show all available commands
> help

# View available commands and position formats
> help
```

## Game Commands

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `move <from> <to>` | Move unit | `move A1 3,4` |
| `attack <att> <tgt>` | Attack target | `attack A1 B2` |
| `select <unit>` | Select unit, show options | `select A1` |
| `end` | End current turn | `end` |
| `status` | Show game status | `status` |
| `units` | Show all units | `units` |
| `player [ID]` | Show player info | `player A` |
| `help` | Show help | `help` |
| `quit` | Exit game | `quit` |

### Recording Commands

| Command | Description | Example |
|---------|-------------|---------|
| `record start` | Begin recording moves | `record start` |
| `record stop` | Stop recording | `record stop` |
| `record show` | Display recorded moves | `record show` |
| `record clear` | Clear move list | `record clear` |
| `replay` | Show moves as JSON | `replay` |

## Position System

### Flexible Position Formats
WeeWar CLI supports three position formats:

#### 1. Unit IDs (Player + Unit Number)
- `A1`, `A2`, `A3` - Player A's units 1, 2, 3
- `B1`, `B12`, `B99` - Player B's units
- `C2` - Player C's unit 2

#### 2. Q/R Hex Coordinates  
- `3,4` - Hex coordinate Q=3, R=4
- `-1,2` - Negative coordinates supported
- `0,0` - Origin coordinate

#### 3. Row/Col Coordinates (prefixed with 'r')
- `r4,5` - Row 4, Column 5 
- `r0,0` - Row 0, Column 0

### Position Examples
```bash
# Move unit A1 to hex coordinate 3,4
> move A1 3,4

# Attack unit B2 with unit at row/col 4,5
> attack r4,5 B2

# Move using different coordinate systems
> move 2,3 A1    # Move from Q/R to unit
> move A1 r5,6   # Move from unit to row/col
```

## Gameplay Tutorial

### Starting a Game
```bash
# Start with a world from storage
$ ./weewar-cli -world $WEEWAR_DATA_ROOT/storage/maps/small-world -interactive
Loading world from $WEEWAR_DATA_ROOT/storage/maps/small-world...
Loaded world: Small World
World loaded successfully
WeeWar CLI - Interactive Mode
Type 'help' for available commands, 'quit' to exit

> status
Turn: 1
Current Player: 0
Game Status: playing
Player A: 1 units

> units
Player A units:
  A1: Type 1 at (1,2) (HP: 100)
```

### Basic Turn Sequence
```bash
# 1. Check game status
> status

# 2. See your units  
> units

# 3. Select a unit to see movement/attack options
> select A1
Selected A1 at (1,2)
Can move to 5 positions: (0,2), (1,1), (1,3), (2,1), (2,2)
No valid attacks available

# 4. Move the unit
> move A1 2,2
Moved A1 from (1,2) to (2,2)

# 5. End your turn
> end  
Turn ended. Current player: 1
```

### Combat Example
```bash
# Select unit to see attack options
> select A1
Selected A1 at (2,3)
Can move to 3 positions: (1,3), (2,2), (3,2)  
Can attack 1 positions: (3,3)

# Attack enemy unit
> attack A1 B1
Attack successful! A1 attacked B1. Defender took 25 damage

# Check unit status
> units
Player A units:
  A1: Type 1 at (2,3) (HP: 100)
Player B units:
  B1: Type 1 at (3,3) (HP: 75)
```

## Advanced Features

### Move Recording
```bash
# Start recording your session
> record start
Recording started

# Play some moves
> move A1 3,4
> attack 3,4 B1
> end

# View recorded moves
> record show
Recorded moves (3):
  1. Turn 1, Player 0: move A1 3,4
  2. Turn 1, Player 0: attack 3,4 B1  
  3. Turn 1, Player 0: end

# Export as JSON
> replay
Move list JSON:
{"moves":[{"command":"move A1 3,4","timestamp":"1642781234","turn":1,"player":0},...]}
```

### Batch Commands via Pipe
```bash
# Create command file
echo -e "select A1\nmove A1 3,4\nend\nstatus" > moves.txt

# Pipe commands to CLI
cat moves.txt | ./weewar-cli -world $WEEWAR_DATA_ROOT/storage/maps/small-world -interactive
```

### Game State Analysis
```bash
# Check current game state
> status
Turn: 3
Current Player: 1
Game Status: playing
Player A: 2 units
Player B: 1 units

# Analyze specific player
> player B
Player B:
  Units: 1
  Status: Waiting

# See all units on the battlefield  
> units
Player A units:
  A1: Type 1 at (3,4) (HP: 85)
  A2: Type 1 at (1,3) (HP: 100)
Player B units:
  B1: Type 1 at (5,5) (HP: 60)
```

## Tips and Tricks

### Efficient Workflow
```bash
# Quick status check
> status

# Select unit before moving to see options
> select A1
Selected A1 at (2,2)
Can move to 4 positions: (1,2), (2,1), (2,3), (3,2)

# Use the information to make tactical decisions
> move A1 2,3
```

### Strategic Planning
```bash  
# Check what you can do this turn
> units
> select A1  # See movement options
> select A2  # See attack options

# Plan your moves
> move A1 3,4
> select A1  # Now check attack options from new position  
> attack A1 B1
> end
```

### Session Recording for Analysis
```bash
# Record important games
> record start
> # ... play game ...
> record show  # Review your moves
> replay      # Get JSON for external analysis
```

### Multiple Coordinate Systems
```bash
# Use whatever format is most convenient
> move A1 3,4     # Unit to hex coordinate
> attack r2,3 B1  # Row/col to unit  
> move 4,5 2,2    # Hex to hex
```

## Position Format Reference

### Unit ID Format: `[A-Z][1-99]`
- **Player Letters**: A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R, S, T, U, V, W, X, Y, Z
- **Unit Numbers**: 1, 2, 3, ..., 99
- **Examples**: `A1`, `B12`, `Z99`

### Q/R Coordinate Format: `<q>,<r>`  
- **Q Coordinate**: Column (can be negative)
- **R Coordinate**: Row (can be negative)
- **Examples**: `0,0`, `3,4`, `-1,2`

### Row/Col Format: `r<row>,<col>`
- **Prefix**: Always starts with `r`  
- **Row/Col**: Traditional grid coordinates
- **Examples**: `r0,0`, `r4,5`, `r10,15`

## Troubleshooting

### Common Issues

**Unit Not Found**:
```bash
> move A5 3,4
Invalid unit: unit A5 does not exist
```
Solution: Use `units` command to see available units

**Invalid Coordinate**:  
```bash
> move A1 invalid
Invalid to position: invalid format: invalid (expected unit ID like A1 or coordinate like 3,4 or r4,5)
```
Solution: Use valid position format (see Position System)

**Move Failed**:
```bash
> move A1 10,10
Move failed: no tile at destination
```
Solution: Use `select A1` to see valid movement options

### Getting Help
```bash
# Show all commands and position formats
> help

# Check what units you can move
> units

# See movement options for a specific unit
> select A1
```

## Command Line Options

```bash
# Start with world from storage
./weewar-cli -world $WEEWAR_DATA_ROOT/storage/maps/small-world -interactive

# Load saved game  
./weewar-cli -load my_game.json -interactive

# Execute commands and save
./weewar-cli -world map1 move A1 3,4 end -save game.json

# Record session
./weewar-cli -world map1 -record session.txt -interactive

# Render game to PNG (when implemented)
./weewar-cli -load game.json -render game.png -width 1024 -height 768
```

## Example: Complete Game Session

```bash
$ ./weewar-cli -world $WEEWAR_DATA_ROOT/storage/maps/small-world -interactive
Loading world from $WEEWAR_DATA_ROOT/storage/maps/small-world...
World loaded successfully
WeeWar CLI - Interactive Mode
Type 'help' for available commands, 'quit' to exit

> record start
Recording started

> status
Turn: 1
Current Player: 0
Game Status: playing
Player A: 1 units

> select A1
Selected A1 at (1,2)
Can move to 6 positions: (0,1), (0,2), (1,1), (1,3), (2,1), (2,2)
No valid attacks available

> move A1 2,2  
Moved A1 from (1,2) to (2,2)

> end
Turn ended. Current player: 1

> record show
Recorded moves (3):
  1. Turn 1, Player 0: record start
  2. Turn 1, Player 0: move A1 2,2
  3. Turn 1, Player 0: end

> quit
Goodbye!
```

The simplified WeeWar CLI provides a focused, efficient interface for turn-based strategy gaming with flexible position formats, move recording, and Unix-friendly batch processing capabilities.

---

**For more information**:
- [Developer Guide](../../DEVELOPER_GUIDE.md)
- [Architecture Documentation](../../ARCHITECTURE.md)  
- [Project Summary](../../SUMMARY.md)
