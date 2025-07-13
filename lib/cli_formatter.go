package weewar

import (
	"fmt"
	"strings"
)

// =============================================================================
// CLI Formatter Implementation
// =============================================================================

// DefaultFormatter implements the CLIFormatter interface
type DefaultFormatter struct {
	colorEnabled bool
}

// NewDefaultFormatter creates a new default formatter
func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{
		colorEnabled: true, // TODO: Auto-detect terminal support
	}
}

// =============================================================================
// Game State Formatting
// =============================================================================

// FormatGameState returns formatted game state string
func (f *DefaultFormatter) FormatGameState(state CLIGameState) string {
	var builder strings.Builder
	
	builder.WriteString(f.Bold("=== Game Status ===\n"))
	builder.WriteString(fmt.Sprintf("Turn: %s\n", f.Colorize(fmt.Sprintf("%d", state.TurnNumber), "cyan")))
	builder.WriteString(fmt.Sprintf("Current Player: %s\n", f.Colorize(fmt.Sprintf("%d", state.CurrentPlayer), "yellow")))
	builder.WriteString(fmt.Sprintf("Game Status: %s\n", f.Colorize(state.GameStatus, "green")))
	builder.WriteString(fmt.Sprintf("Map: %s\n", state.MapName))
	
	if len(state.Players) > 0 {
		builder.WriteString(f.Bold("\nPlayers:\n"))
		for _, player := range state.Players {
			status := "waiting"
			if player.IsActive {
				status = f.Colorize("active", "green")
			}
			builder.WriteString(fmt.Sprintf("  Player %d: %d units (%s)\n", 
				player.PlayerID, player.UnitCount, status))
		}
	}
	
	if len(state.AvailableActions) > 0 {
		builder.WriteString(f.Bold("\nAvailable Actions:\n"))
		for _, action := range state.AvailableActions {
			builder.WriteString(fmt.Sprintf("  • %s\n", action))
		}
	}
	
	return builder.String()
}

// FormatMap returns formatted map representation
func (f *DefaultFormatter) FormatMap(mapInfo CLIMapInfo) string {
	var builder strings.Builder
	
	builder.WriteString(f.Bold("=== Map Information ===\n"))
	builder.WriteString(fmt.Sprintf("Name: %s\n", f.Colorize(mapInfo.Name, "cyan")))
	builder.WriteString(fmt.Sprintf("Size: %s\n", mapInfo.Size))
	builder.WriteString(fmt.Sprintf("Players: %d\n", mapInfo.PlayerCount))
	builder.WriteString(fmt.Sprintf("Tiles: %d\n", mapInfo.TileCount))
	
	if len(mapInfo.TerrainTypes) > 0 {
		builder.WriteString(f.Bold("\nTerrain Distribution:\n"))
		for terrain, count := range mapInfo.TerrainTypes {
			builder.WriteString(fmt.Sprintf("  %s: %d\n", terrain, count))
		}
	}
	
	if mapInfo.Description != "" {
		builder.WriteString(f.Bold("\nDescription:\n"))
		builder.WriteString(fmt.Sprintf("%s\n", mapInfo.Description))
	}
	
	return builder.String()
}

// FormatUnits returns formatted unit list
func (f *DefaultFormatter) FormatUnits(units []CLIUnitInfo) string {
	if len(units) == 0 {
		return "No units found.\n"
	}
	
	var builder strings.Builder
	builder.WriteString(f.Bold("=== Units ===\n"))
	
	// Create table
	headers := []string{"ID", "Type", "Player", "Position", "Health", "Movement", "Status"}
	rows := make([][]string, len(units))
	
	for i, unit := range units {
		healthColor := "green"
		if unit.Health < 50 {
			healthColor = "yellow"
		}
		if unit.Health < 25 {
			healthColor = "red"
		}
		
		statusColor := "green"
		if unit.Status == "moved" {
			statusColor = "yellow"
		} else if unit.Status == "disabled" {
			statusColor = "red"
		}
		
		rows[i] = []string{
			fmt.Sprintf("%d", unit.UnitID),
			unit.UnitType,
			fmt.Sprintf("%d", unit.PlayerID),
			unit.Position,
			f.Colorize(fmt.Sprintf("%d", unit.Health), healthColor),
			fmt.Sprintf("%d", unit.Movement),
			f.Colorize(unit.Status, statusColor),
		}
	}
	
	builder.WriteString(f.Table(headers, rows))
	return builder.String()
}

// FormatPlayerInfo returns formatted player information
func (f *DefaultFormatter) FormatPlayerInfo(player CLIPlayerInfo) string {
	var builder strings.Builder
	
	builder.WriteString(f.Bold(fmt.Sprintf("=== Player %d ===\n", player.PlayerID)))
	builder.WriteString(fmt.Sprintf("Name: %s\n", player.Name))
	builder.WriteString(fmt.Sprintf("Units: %d\n", player.UnitCount))
	
	status := "Waiting"
	statusColor := "yellow"
	if player.IsActive {
		status = "Active"
		statusColor = "green"
	}
	builder.WriteString(fmt.Sprintf("Status: %s\n", f.Colorize(status, statusColor)))
	
	if player.IsAI {
		builder.WriteString(fmt.Sprintf("AI Difficulty: %s\n", player.AIDifficulty))
	}
	
	return builder.String()
}

// =============================================================================
// Utility Formatting
// =============================================================================

// FormatPosition converts row/col to chess notation (A1, B2, etc.)
func (f *DefaultFormatter) FormatPosition(row, col int) string {
	return FormatPositionToString(row, col)
}

// ParsePosition converts chess notation to row/col
func (f *DefaultFormatter) ParsePosition(position string) (row, col int, valid bool) {
	return ParsePositionFromString(position)
}

// FormatHealth returns formatted health display
func (f *DefaultFormatter) FormatHealth(current, max int) string {
	percentage := float64(current) / float64(max) * 100
	
	var color string
	if percentage > 75 {
		color = "green"
	} else if percentage > 50 {
		color = "yellow"
	} else if percentage > 25 {
		color = "red"
	} else {
		color = "red"
	}
	
	healthBar := f.createHealthBar(current, max, 10)
	return fmt.Sprintf("%s %s/%d (%.0f%%)", 
		f.Colorize(healthBar, color), 
		f.Colorize(fmt.Sprintf("%d", current), color),
		max, percentage)
}

// FormatMovement returns formatted movement display
func (f *DefaultFormatter) FormatMovement(current, max int) string {
	if current == 0 {
		return f.Colorize("No movement", "red")
	}
	
	color := "green"
	if current < max {
		color = "yellow"
	}
	
	return f.Colorize(fmt.Sprintf("%d/%d", current, max), color)
}

// =============================================================================
// Text Styling
// =============================================================================

// Colorize applies color to text (if terminal supports it)
func (f *DefaultFormatter) Colorize(text, color string) string {
	if !f.colorEnabled {
		return text
	}
	
	colorCode := ""
	switch strings.ToLower(color) {
	case "red":
		colorCode = ColorRed
	case "green":
		colorCode = ColorGreen
	case "yellow":
		colorCode = ColorYellow
	case "blue":
		colorCode = ColorBlue
	case "purple":
		colorCode = ColorPurple
	case "cyan":
		colorCode = ColorCyan
	case "white":
		colorCode = ColorWhite
	default:
		return text
	}
	
	return fmt.Sprintf("%s%s%s", colorCode, text, ColorReset)
}

// Bold applies bold formatting to text
func (f *DefaultFormatter) Bold(text string) string {
	if !f.colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", ColorBold, text, ColorReset)
}

// Italic applies italic formatting to text
func (f *DefaultFormatter) Italic(text string) string {
	if !f.colorEnabled {
		return text
	}
	return fmt.Sprintf("%s%s%s", ColorItalic, text, ColorReset)
}

// Table creates formatted table from data
func (f *DefaultFormatter) Table(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}
	
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				// Strip color codes for width calculation
				cellLen := len(f.stripColorCodes(cell))
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}
	
	var builder strings.Builder
	
	// Header
	builder.WriteString(f.Bold(f.formatTableRow(headers, widths)))
	
	// Separator
	separator := make([]string, len(headers))
	for i, width := range widths {
		separator[i] = strings.Repeat("-", width)
	}
	builder.WriteString(f.formatTableRow(separator, widths))
	
	// Data rows
	for _, row := range rows {
		builder.WriteString(f.formatTableRow(row, widths))
	}
	
	return builder.String()
}

// =============================================================================
// Helper Functions
// =============================================================================

// createHealthBar creates a visual health bar
func (f *DefaultFormatter) createHealthBar(current, max, width int) string {
	if max == 0 {
		return strings.Repeat("_", width)
	}
	
	filled := int(float64(current) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// formatTableRow formats a single table row
func (f *DefaultFormatter) formatTableRow(row []string, widths []int) string {
	var builder strings.Builder
	
	for i, cell := range row {
		if i < len(widths) {
			// Calculate padding needed (accounting for color codes)
			cellLen := len(f.stripColorCodes(cell))
			padding := widths[i] - cellLen
			if padding < 0 {
				padding = 0
			}
			
			builder.WriteString(cell)
			builder.WriteString(strings.Repeat(" ", padding))
			
			if i < len(row)-1 {
				builder.WriteString(" | ")
			}
		}
	}
	
	builder.WriteString("\n")
	return builder.String()
}

// stripColorCodes removes ANSI color codes from text for length calculation
func (f *DefaultFormatter) stripColorCodes(text string) string {
	// Simple implementation - remove common ANSI codes
	codes := []string{
		ColorReset, ColorRed, ColorGreen, ColorYellow, ColorBlue,
		ColorPurple, ColorCyan, ColorWhite, ColorBold, ColorItalic,
	}
	
	result := text
	for _, code := range codes {
		result = strings.ReplaceAll(result, code, "")
	}
	
	return result
}

// SetColorEnabled enables or disables color output
func (f *DefaultFormatter) SetColorEnabled(enabled bool) {
	f.colorEnabled = enabled
}

// IsColorEnabled returns whether color output is enabled
func (f *DefaultFormatter) IsColorEnabled() bool {
	return f.colorEnabled
}