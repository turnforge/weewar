package weewar

import (
	"fmt"
	"strings"
)

// =============================================================================
// WeeWar CLI Interface Definitions
// =============================================================================
// This file defines CLI-specific interfaces and types for the WeeWar game system.
// These interfaces focus on command-line interaction, text-based gameplay,
// and console-based game management.

// =============================================================================
// CLI Data Types
// =============================================================================

// CLICommand represents a parsed command line input
type CLICommand struct {
	Command   string            `json:"command"`   // Primary command (e.g., "move", "attack")
	Arguments []string          `json:"arguments"` // Command arguments
	Options   map[string]string `json:"options"`   // Command options/flags
	Raw       string            `json:"raw"`       // Original command string
}

// CLIResponse represents the result of a CLI command
type CLIResponse struct {
	Success bool   `json:"success"` // Whether command succeeded
	Message string `json:"message"` // Response message
	Data    string `json:"data"`    // Additional data (e.g., formatted output)
	Error   string `json:"error"`   // Error message if failed
}

// CLIDisplayMode represents different display modes for CLI output
type CLIDisplayMode int

const (
	DisplayCompact CLIDisplayMode = iota
	DisplayDetailed
	DisplayASCII
	DisplayJSON
)

func (d CLIDisplayMode) String() string {
	switch d {
	case DisplayCompact:
		return "compact"
	case DisplayDetailed:
		return "detailed"
	case DisplayASCII:
		return "ascii"
	case DisplayJSON:
		return "json"
	default:
		return "unknown"
	}
}

// CLIGameState represents game state for CLI display
type CLIGameState struct {
	CurrentPlayer  int                    `json:"currentPlayer"`
	TurnNumber     int                    `json:"turnNumber"`
	GameStatus     string                 `json:"gameStatus"`
	MapName        string                 `json:"mapName"`
	Players        []CLIPlayerInfo        `json:"players"`
	UnitsPerPlayer map[int]int            `json:"unitsPerPlayer"`
	LastAction     string                 `json:"lastAction"`
	AvailableActions []string             `json:"availableActions"`
}

// CLIPlayerInfo represents player information for CLI display
type CLIPlayerInfo struct {
	PlayerID     int    `json:"playerId"`
	Name         string `json:"name"`
	UnitCount    int    `json:"unitCount"`
	IsActive     bool   `json:"isActive"`
	IsAI         bool   `json:"isAI"`
	AIDifficulty string `json:"aiDifficulty"`
}

// CLIUnitInfo represents unit information for CLI display
type CLIUnitInfo struct {
	UnitID       int    `json:"unitId"`
	UnitType     string `json:"unitType"`
	PlayerID     int    `json:"playerId"`
	Position     string `json:"position"`     // "A1", "B2", etc.
	Health       int    `json:"health"`
	Movement     int    `json:"movement"`
	CanMove      bool   `json:"canMove"`
	CanAttack    bool   `json:"canAttack"`
	Status       string `json:"status"`       // "ready", "moved", "attacked", "disabled"
}

// CLIMapInfo represents map information for CLI display
type CLIMapInfo struct {
	Name         string            `json:"name"`
	Size         string            `json:"size"`         // "12x8", "16x12", etc.
	PlayerCount  int               `json:"playerCount"`
	TileCount    int               `json:"tileCount"`
	TerrainTypes map[string]int    `json:"terrainTypes"` // terrain type -> count
	Description  string            `json:"description"`
}

// =============================================================================
// CLI Interface
// =============================================================================

// CLIInterface provides command-line interaction capabilities
type CLIInterface interface {
	// Command Processing
	// ExecuteCommand processes text commands
	// Called by: CLI input loop, Batch command processing, Testing
	// Returns: Command response with success/failure and message
	ExecuteCommand(command string) *CLIResponse
	
	// ParseCommand parses command string into structured format
	// Called by: Command processing, Input validation
	// Returns: Parsed command structure
	ParseCommand(commandStr string) *CLICommand
	
	// GetAvailableCommands returns list of valid commands
	// Called by: Help system, Auto-completion, Command validation
	// Returns: Array of command names
	GetAvailableCommands() []string
	
	// GetCommandHelp returns help text for specific command
	// Called by: Help system, Interactive tutorials, Error messages
	// Returns: Help text string
	GetCommandHelp(command string) string
	
	// ValidateCommand checks if command is valid in current context
	// Called by: Input validation, Command preprocessing
	// Returns: Whether command is valid and error message if not
	ValidateCommand(cmd *CLICommand) (bool, string)
	
	// Display Functions
	// PrintGameState outputs current game state to console
	// Called by: CLI status command, Debug output, Turn summaries
	PrintGameState()
	
	// PrintMap outputs map representation to console
	// Called by: CLI map command, Debug output, Save file inspection
	PrintMap()
	
	// PrintUnits outputs unit list to console
	// Called by: CLI units command, Debug output, Planning assistance
	PrintUnits()
	
	// PrintPlayerInfo outputs player statistics
	// Called by: CLI player command, Turn summaries, Statistics
	PrintPlayerInfo(playerID int)
	
	// PrintHelp outputs help information
	// Called by: Help command, Error messages, Tutorial
	PrintHelp(topic string)
	
	// Display Configuration
	// SetDisplayMode changes output format
	// Called by: User preferences, Output formatting
	SetDisplayMode(mode CLIDisplayMode)
	
	// GetDisplayMode returns current display mode
	// Called by: Output formatting, Settings display
	GetDisplayMode() CLIDisplayMode
	
	// SetVerbose enables/disables verbose output
	// Called by: Debug mode, User preferences
	SetVerbose(verbose bool)
	
	// IsVerbose returns whether verbose output is enabled
	// Called by: Output formatting, Debug information
	IsVerbose() bool
	
	// Interactive Functions
	// StartInteractiveMode begins interactive CLI gameplay
	// Called by: CLI main function, Interactive testing, Tutorial mode
	StartInteractiveMode()
	
	// ProcessTurn handles single player turn interactively
	// Called by: Interactive mode, Turn-based automation, AI demonstration
	ProcessTurn(playerID int)
	
	// PromptForInput prompts user for input with validation
	// Called by: Interactive commands, User confirmation
	// Returns: User input string
	PromptForInput(prompt string, validator func(string) bool) string
	
	// ConfirmAction prompts for yes/no confirmation
	// Called by: Destructive actions, Game state changes
	// Returns: Whether user confirmed
	ConfirmAction(message string) bool
	
	// Game Management
	// SaveGameToFile saves current game state to file
	// Called by: Save command, Auto-save, Game persistence
	SaveGameToFile(filename string) error
	
	// LoadGameFromFile loads game state from file
	// Called by: Load command, Game restoration
	LoadGameFromFile(filename string) error
	
	// RenderToFile renders current game state to PNG file
	// Called by: Render command, Screenshot capture, Debug output
	RenderToFile(filename string, width, height int) error
	
	// Batch Processing
	// ExecuteBatchCommands processes multiple commands from file
	// Called by: Script execution, Automated testing, Replay
	ExecuteBatchCommands(filename string) error
	
	// RecordSession records commands to file for replay
	// Called by: Session recording, Testing, Demonstration
	RecordSession(filename string) error
	
	// StopRecording stops current session recording
	// Called by: Recording completion, User request
	StopRecording()
}

// =============================================================================
// CLI Formatter Interface
// =============================================================================

// CLIFormatter provides text formatting capabilities
type CLIFormatter interface {
	// Game State Formatting
	// FormatGameState returns formatted game state string
	// Called by: Status display, Debug output
	FormatGameState(state CLIGameState) string
	
	// FormatMap returns formatted map representation
	// Called by: Map display, ASCII art generation
	FormatMap(mapInfo CLIMapInfo) string
	
	// FormatUnits returns formatted unit list
	// Called by: Unit display, Status reports
	FormatUnits(units []CLIUnitInfo) string
	
	// FormatPlayerInfo returns formatted player information
	// Called by: Player display, Statistics
	FormatPlayerInfo(player CLIPlayerInfo) string
	
	// Utility Formatting
	// FormatPosition converts row/col to chess notation (A1, B2, etc.)
	// Called by: Position display, Command parsing
	FormatPosition(row, col int) string
	
	// ParsePosition converts chess notation to row/col
	// Called by: Command parsing, Input validation
	ParsePosition(position string) (row, col int, valid bool)
	
	// FormatHealth returns formatted health display
	// Called by: Unit status, Health bars
	FormatHealth(current, max int) string
	
	// FormatMovement returns formatted movement display
	// Called by: Unit status, Movement indicators
	FormatMovement(current, max int) string
	
	// Text Styling
	// Colorize applies color to text (if terminal supports it)
	// Called by: Status display, Error messages, Highlights
	Colorize(text, color string) string
	
	// Bold applies bold formatting to text
	// Called by: Headers, Important messages
	Bold(text string) string
	
	// Italic applies italic formatting to text
	// Called by: Descriptions, Secondary information
	Italic(text string) string
	
	// Table creates formatted table from data
	// Called by: Unit lists, Statistics display
	Table(headers []string, rows [][]string) string
}

// =============================================================================
// CLI Utilities
// =============================================================================

// Command constants for CLI commands
const (
	CmdMove     = "move"
	CmdAttack   = "attack"
	CmdStatus   = "status"
	CmdMap      = "map"
	CmdUnits    = "units"
	CmdPlayer   = "player"
	CmdHelp     = "help"
	CmdSave     = "save"
	CmdLoad     = "load"
	CmdRender   = "render"
	CmdEnd      = "end"
	CmdQuit     = "quit"
	CmdUndo     = "undo"
	CmdRedo     = "redo"
	CmdHint     = "hint"
	CmdReplay   = "replay"
	CmdRecord   = "record"
	CmdNew      = "new"
	CmdClear    = "clear"
	CmdVerbose  = "verbose"
	CmdCompact  = "compact"
)

// Color constants for CLI output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorItalic = "\033[3m"
)

// ParsePositionFromString converts chess notation to row/col coordinates
func ParsePositionFromString(pos string) (row, col int, valid bool) {
	pos = strings.ToUpper(strings.TrimSpace(pos))
	if len(pos) < 2 {
		return 0, 0, false
	}
	
	// Parse column (A-Z)
	colChar := pos[0]
	if colChar < 'A' || colChar > 'Z' {
		return 0, 0, false
	}
	col = int(colChar - 'A')
	
	// Parse row (1-99)
	rowStr := pos[1:]
	var rowNum int
	if _, err := fmt.Sscanf(rowStr, "%d", &rowNum); err != nil {
		return 0, 0, false
	}
	
	if rowNum < 1 || rowNum > 99 {
		return 0, 0, false
	}
	
	row = rowNum - 1 // Convert to 0-based
	return row, col, true
}

// FormatPositionToString converts row/col coordinates to chess notation
func FormatPositionToString(row, col int) string {
	if row < 0 || col < 0 || col > 25 {
		return "??"
	}
	
	colChar := string(rune('A' + col))
	rowNum := row + 1 // Convert to 1-based
	
	return fmt.Sprintf("%s%d", colChar, rowNum)
}