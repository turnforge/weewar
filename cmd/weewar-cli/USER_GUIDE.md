# WeeWar CLI User Guide

A comprehensive guide for playing WeeWar using the command-line interface (CLI).

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
go build -o /tmp/weewar-cli ./cmd/weewar-cli

# Make it executable and add to PATH (optional)
chmod +x /tmp/weewar-cli
```

### Quick Start
```bash
# Start a new interactive game
/tmp/weewar-cli -new -interactive

# Start with specific number of players
/tmp/weewar-cli -new -players 4 -interactive

# Run single commands
/tmp/weewar-cli -new status map
```

## REPL Interface

### Understanding the Prompt
The CLI uses a smart prompt that shows game state:

```
weewar[T1:P0]> 
```

- `T1` = Turn number (1, 2, 3, etc.)
- `P0` = Current player (0, 1, 2, etc.)

### Special Game States
```
weewar[GAME ENDED - Player 1 Won]>     # Game over
weewar>                                # No game loaded
```

### Getting Help
```bash
# Show all available commands
weewar[T1:P0]> help

# Get help for specific command
weewar[T1:P0]> help move

# Show available actions for current player
weewar[T1:P0]> actions
```

## Game Commands

### Basic Commands

| Command | Description | Example |
|---------|-------------|---------|
| `new [players]` | Start new game | `new 2` |
| `status` / `s` | Show game status | `status` |
| `map` | Display game map | `map` |
| `units` | Show all units | `units` |
| `actions` | Show available actions | `actions` |
| `help` | Show help | `help move` |
| `quit` | Exit game | `quit` |

### Movement Commands

| Command | Description | Example |
|---------|-------------|---------|
| `move <from> <to>` | Move unit | `move A1 B2` |
| `attack <from> <to>` | Attack unit | `attack A1 B2` |
| `end` | End current turn | `end` |

### Prediction Commands

| Command | Description | Example |
|---------|-------------|---------|
| `predict <from> <to>` | Show damage prediction | `predict A1 B2` |
| `attackoptions <unit>` | Show attack targets | `attackoptions A1` |
| `moveoptions <unit>` | Show movement options | `moveoptions A1` |

### Information Commands

| Command | Description | Example |
|---------|-------------|---------|
| `turn` | Detailed turn info | `turn` |
| `player [id]` | Player information | `player 1` |
| `refresh` / `r` | Refresh display | `refresh` |

### File Operations

| Command | Description | Example |
|---------|-------------|---------|
| `save <file>` | Save game | `save game.json` |
| `load <file>` | Load game | `load game.json` |
| `render <file>` | Render PNG | `render game.png` |

### Display Options

| Command | Description | Example |
|---------|-------------|---------|
| `verbose` | Toggle verbose mode | `verbose` |
| `compact` | Set compact display | `compact` |

## Position System

### Chess Notation
WeeWar CLI uses chess-style notation for positions on a hex grid:

```
         A     B     C     D     E     F     G     H     I     J     K     L
 1    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱
      --    --    --    --    --    --    --    --    --    --    --    --

 2      🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱
        --    P0    P0    --    --    --    --    --    --    --    --    --

 3    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱
      --    --    --    --    --    --    --    --    --    --    --    --

 7    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱
      --    --    --    --    --    --    --    --    --    P1    P1    --

 8      🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱    🌱
        --    --    --    --    --    --    --    --    --    --    --    --
```

### Position Examples
- `A1` = Top-left corner
- `B2` = Second column, second row  
- `L8` = Bottom-right corner

### Moving Units
```bash
# Move unit from B2 to B3
weewar[T1:P0]> move B2 B3
✓ Unit moved from B2 to B3

# Attack enemy unit at C4 with unit at B3
weewar[T1:P0]> attack B3 C4
✓ Attack from B3 to C4: 25 damage dealt

# Check updated map to see new positions
weewar[T1:P0]> map
=== Game Map ===
         A     B     C     D     E     F
 1    🌱    🌱    🌱    🌱    🌱    🌱
      --    --    --    --    --    --

 2      🌱    🌱    🌱    🌱    🌱    🌱
        --    --    P1    --    --    --

 3    🌱    🌱    🌱    🌱    🌱    🌱
      --    P0    --    --    --    --

 4      🌱    🌱    🌱    🌱    🌱    🌱
        --    --    --    --    --    --
```

## Gameplay Tutorial

### Starting a New Game
```bash
# Start interactive game
weewar[T1:P0]> new

# Check initial game state
weewar[T1:P0]> status
=== Game Status ===
Turn: 1
Current Player: 0
Game Status: playing
Map: DefaultMap
Players: 2
  Player 0: 2 units
  Player 1: 2 units

# See what actions are available
weewar[T1:P0]> actions
=== Available Actions (Player 0) ===
  Move unit at B2 (movement: 3)
  Move unit at C2 (movement: 3)
  No attack opportunities
  End turn (use 'end' command)
```

### Basic Turn Sequence
```bash
# 1. Check available actions
weewar[T1:P0]> actions

# 2. View the map
weewar[T1:P0]> map

# 3. Move units
weewar[T1:P0]> move B2 B3

# 4. Check for attack opportunities
weewar[T1:P0]> actions

# 5. End turn when done
weewar[T1:P0]> end
```

### Understanding the Map
The map now uses rich emoji symbols for terrain types with hex grid layout. Each tile spans 2 lines:

```bash
weewar[T1:P0]> map
=== Game Map ===
Size: 8x12
         A     B     C     D     E     F     G     H     I     J     K     L
 1    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱
      --    --    --    --    --    --    --    --    --    --    --    --

 2      🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️
        --    P0    P0    --    --    --    --    --    --    --    --    --

 3    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱
      --    --    --    --    --    --    --    --    --    --    --    --

 4      🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱
        --    --    --    --    --    --    --    --    --    --    --    --

 5    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱
      --    --    --    --    --    --    --    --    --    --    --    --

 6      🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️
        --    --    --    --    --    --    --    --    --    --    --    --

 7    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🌱    🌱
      --    --    --    --    --    --    --    --    --    P1    P1    --

 8      🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱    🌱    🏜️    🌱    🌱
        --    --    --    --    --    --    --    --    --    --    --    --

Terrain Key:
🌱=Grass  🏜️=Desert  🌊=Water  ⛰️=Mountains  🗿=Rock  🏥=Hospital
🌾=Swamp  🌲=Forest  🌋=Lava  💧=Shallow  🚀=Missile  🌉=Bridge
⛏️=Mines  🏙️=City  🛣️=Road  🗼=Tower  ❄️=Snow  🏰=Land Base
🏛️=Naval Base  ✈️=Airport  ❓=Unknown

Units: P0, P1, etc. (Player number), -- = No unit
Hex Layout: Offset rows based on EvenRowsOffset flag
```

**Key Features:**
- **2-Line Format**: Each tile has terrain emoji on top line, unit info on bottom line
- **Emoji Terrain**: Each terrain type has a distinctive emoji for easy recognition
- **Hex Grid**: Notice how alternate rows are offset to show the hexagonal structure
- **Clear Unit Display**: Units appear as "P0", "P1", etc. on separate line from terrain
- **Better Spacing**: More space around each tile for improved readability
- **Rich Variety**: Supports up to 99 different terrain types with clear visual distinction

### Combat Example
```bash
# Check if units can attack
weewar[T1:P0]> actions
=== Available Actions (Player 0) ===
  Attack with unit at B3 -> enemy at C4

# Execute attack
weewar[T1:P0]> attack B3 C4
✓ Attack from B3 to C4: 25 damage dealt

# Check result
weewar[T1:P0]> units
=== Units ===
Player 0: 2 units
  1. B3 - Type:1 Health:100 Movement:2
  2. C2 - Type:1 Health:100 Movement:3
Player 1: 2 units
  1. C4 - Type:1 Health:75 Movement:3
  2. K7 - Type:1 Health:100 Movement:3
```

### Prediction and Planning Examples

```bash
# Show movement options for a unit
weewar[T1:P0]> moveoptions B2
=== Movement Options for Soldier (Basic) at B2 ===
Movement Points: 3
Available positions (4):
  1. A2 - Grass (Move Cost: 1)
  2. B1 - Grass (Move Cost: 1)
  3. B3 - Grass (Move Cost: 1)
  4. C2 - Grass (Move Cost: 1)

Use 'move <unit> <destination>' to move the unit.

# Show attack options for a unit
weewar[T1:P0]> attackoptions B3
=== Attack Options for Soldier (Basic) at B3 ===
Available targets (1):
  1. C4 - Soldier (Basic) (Player 1, Health: 100)

Use 'predict <unit> <target>' to see damage prediction.

# Predict damage before attacking
weewar[T1:P0]> predict B3 C4
=== Damage Prediction: B3 attacking C4 ===
Attacker: Soldier (Basic) at B3 (Player 0, Health: 100)
Target: Soldier (Basic) at C4 (Player 1, Health: 100)

Damage Range: 20-30 damage
Expected Damage: 25.0
Damage Probabilities:
  20 damage: 33.3%
  25 damage: 33.3%
  30 damage: 33.3%

Predicted Target Health: 75
```

## Advanced Features

### Batch Processing
```bash
# Create command file
echo "new
move B2 B3
attack B3 C4
end
status" > commands.txt

# Execute batch commands
/tmp/weewar-cli -batch commands.txt
```

### Save and Load Games
```bash
# Save current game
weewar[T1:P0]> save my_game.json
✓ Game saved to my_game.json

# Load saved game
weewar> load my_game.json
✓ Game loaded from my_game.json

# Start CLI with saved game
/tmp/weewar-cli -load my_game.json -interactive
```

### PNG Rendering
```bash
# Render current game state
weewar[T1:P0]> render game.png
✓ Game rendered to game.png (800x600)

# Render with custom size
weewar[T1:P0]> render game_large.png 1200 900
✓ Game rendered to game_large.png (1200x900)

# Render from command line
/tmp/weewar-cli -load game.json -render game.png
```

### Session Recording
```bash
# Start recording session
/tmp/weewar-cli -record session.txt -interactive

# All commands will be recorded to session.txt
# Later replay with:
/tmp/weewar-cli -batch session.txt
```

### Multiple CLI Modes
```bash
# Interactive mode
/tmp/weewar-cli -new -interactive

# Single commands
/tmp/weewar-cli -new status map units

# Batch with save
/tmp/weewar-cli -new -batch commands.txt -save final_game.json

# Load, render, and save
/tmp/weewar-cli -load game.json -render game.png -save updated_game.json
```

## Tips and Tricks

### Efficient Gameplay
```bash
# Use shortcuts for common commands
weewar[T1:P0]> s          # Quick status
weewar[T1:P0]> r          # Refresh display
weewar[T1:P0]> actions    # See available actions

# Chain commands in command line
/tmp/weewar-cli -new status map actions
```

### Visual Debugging
```bash
# Render game state to see what's happening
weewar[T1:P0]> render debug.png

# Use verbose mode for detailed output
weewar[T1:P0]> verbose
weewar[T1:P0]> move B2 B3  # More detailed feedback
```

### Strategic Planning
```bash
# Check turn information
weewar[T1:P0]> turn
=== Turn Information ===
Turn Number: 1
Current Player: 0
Game Status: playing
Can End Turn: true
Your Units: 2
Units with Movement: 2

# Analyze available actions
weewar[T1:P0]> actions
=== Available Actions (Player 0) ===
  Move unit at B2 (movement: 3)
  Move unit at C2 (movement: 3)
  No attack opportunities
  End turn (use 'end' command)
```

### Error Recovery
```bash
# If you make a mistake, check the error message
weewar[T1:P0]> move A1 Z9
✗ Invalid to position: Z9
  Error: Use format like A1, B2, etc.

# Use help to understand commands
weewar[T1:P0]> help move
move <from> <to> - Move unit from one position to another (e.g., 'move A1 B2')
```

## Common Command Patterns

### Start of Turn
```bash
weewar[T1:P0]> s          # Check status
weewar[T1:P0]> actions    # See available actions
weewar[T1:P0]> map        # View battlefield
```

### During Turn
```bash
weewar[T1:P0]> move B2 B3  # Move units
weewar[T1:P0]> actions     # Check for attacks
weewar[T1:P0]> attack B3 C4 # Attack if possible
```

### End of Turn
```bash
weewar[T1:P0]> actions     # Verify no more actions
weewar[T1:P0]> end         # End turn
```

### Analysis
```bash
weewar[T1:P0]> units       # Review unit status
weewar[T1:P0]> player 1    # Check opponent
weewar[T1:P0]> render analysis.png # Save visual record
```

## Troubleshooting

### Common Issues

**Invalid Position Error**:
```bash
weewar[T1:P0]> move A1 Z9
✗ Invalid to position: Z9
```
Solution: Use valid chess notation (A1-L8 for standard map)

**Unit Not Found**:
```bash
weewar[T1:P0]> move A1 B2
✗ No unit found at position A1
```
Solution: Check unit positions with `units` or `map` command

**Wrong Player Turn**:
```bash
weewar[T1:P0]> move J7 J6
✗ Unit at J7 belongs to player 1, but it's player 0's turn
```
Solution: Only move your own units during your turn

### Getting Help
```bash
# Show all commands
weewar[T1:P0]> help

# Show command-specific help
weewar[T1:P0]> help move

# Check available actions
weewar[T1:P0]> actions

# View current game state
weewar[T1:P0]> status
```

## Command Reference

### Complete Command List
```
Game Management:
  new [players]        - Start new game
  load <file>         - Load saved game
  save <file>         - Save current game
  quit                - Exit

Movement:
  move <from> <to>    - Move unit
  attack <from> <to>  - Attack unit
  end                 - End turn

Information:
  status / s          - Game status
  map                 - Display map
  units               - Show units
  player [id]         - Player info
  turn                - Turn details
  actions             - Available actions
  help [command]      - Show help
  refresh / r         - Refresh display

Display:
  verbose             - Toggle verbose mode
  compact             - Set compact display

Files:
  render <file> [w] [h] - Render PNG
```

### Exit Codes
- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - File operation failed

## Examples

### Complete Game Session
```bash
# Start new game
$ /tmp/weewar-cli -new -interactive

# Game begins
weewar[T1:P0]> actions
=== Available Actions (Player 0) ===
  Move unit at B2 (movement: 3)
  Move unit at C2 (movement: 3)

weewar[T1:P0]> move B2 B3
✓ Unit moved from B2 to B3

weewar[T1:P0]> move C2 C3
✓ Unit moved from C2 to C3

weewar[T1:P0]> end
✓ Turn ended. Now player 1's turn (Turn 1)

weewar[T1:P1]> actions
=== Available Actions (Player 1) ===
  Move unit at J7 (movement: 3)
  Move unit at K7 (movement: 3)

weewar[T1:P1]> move J7 J6
✓ Unit moved from J7 to J6

weewar[T1:P1]> end
✓ Turn ended. Now player 0's turn (Turn 2)

weewar[T2:P0]> save game.json
✓ Game saved to game.json

weewar[T2:P0]> render game.png
✓ Game rendered to game.png (800x600)

weewar[T2:P0]> quit
✓ Goodbye!
```

The WeeWar CLI provides a comprehensive, professional interface for playing turn-based strategy games. With its chess notation system, rich feedback, and powerful features, it offers an excellent command-line gaming experience.

---

**For more information**:
- [Developer Guide](../../DEVELOPER_GUIDE.md)
- [Architecture Documentation](../../ARCHITECTURE.md)
- [Project Summary](../../SUMMARY.md)