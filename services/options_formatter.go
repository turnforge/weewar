package services

import (
	"fmt"
	"strings"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// OptionFormatter provides utilities for formatting game options for display
type OptionFormatter struct {
	ShowPaths     bool // Whether to show movement paths
	DetailedPaths bool // Whether to show detailed path information
}

// FormatOption formats a single game option for display
func (f *OptionFormatter) FormatOption(option *v1.GameOption, allPaths *v1.AllPaths) string {
	switch opt := option.OptionType.(type) {
	case *v1.GameOption_Move:
		return f.FormatMoveUnitAction(opt.Move, allPaths)
	case *v1.GameOption_Attack:
		return f.FormatAttackUnitAction(opt.Attack)
	case *v1.GameOption_EndTurn:
		return "End turn"
	case *v1.GameOption_Build:
		return f.FormatBuildUnitAction(opt.Build)
	default:
		return "Unknown option"
	}
}

// FormatMoveUnitAction formats a movement option with optional path visualization
func (f *OptionFormatter) FormatMoveUnitAction(moveOpt *v1.MoveUnitAction, allPaths *v1.AllPaths) string {
	targetCoord := CoordFromInt32(moveOpt.To.Q, moveOpt.To.R)

	// Basic format: "move to (q,r) (cost: X)"
	result := fmt.Sprintf("move to %s (cost: %f)",
		targetCoord.String(), moveOpt.MovementCost)

	// Add path visualization if available and requested
	if f.ShowPaths && allPaths != nil {
		path, err := ReconstructPath(allPaths, moveOpt.To.Q, moveOpt.To.R)
		if err == nil && path != nil {
			if f.DetailedPaths {
				pathStr := FormatPathDetailed(path, "   ")
				result += "\n" + pathStr
			} else {
				pathStr := FormatPathCompact(path)
				result += fmt.Sprintf("\n   Path: %s", pathStr)
			}
		}
	}

	return result
}

// FormatAttackUnitAction formats an attack option with damage estimate
func (f *OptionFormatter) FormatAttackUnitAction(attackOpt *v1.AttackUnitAction) string {
	targetCoord := CoordFromInt32(attackOpt.Defender.Q, attackOpt.Defender.R)

	// Include target unit type and damage estimate
	result := fmt.Sprintf("attack %s", targetCoord.String())

	if attackOpt.TargetUnitType > 0 {
		result += fmt.Sprintf(" (type %d", attackOpt.TargetUnitType)
		if attackOpt.DamageEstimate > 0 {
			result += fmt.Sprintf(", damage est: %d", attackOpt.DamageEstimate)
		}
		result += ")"
	}

	return result
}

// FormatBuildUnitAction formats a build option
func (f *OptionFormatter) FormatBuildUnitAction(buildOpt *v1.BuildUnitAction) string {
	if buildOpt == nil {
		return "build"
	}

	// Format based on what building information is available
	var parts []string

	if buildOpt.UnitType > 0 {
		parts = append(parts, fmt.Sprintf("unit type %d", buildOpt.UnitType))
	}

	if buildOpt.Cost > 0 {
		parts = append(parts, fmt.Sprintf("cost: %d", buildOpt.Cost))
	}

	if len(parts) > 0 {
		return fmt.Sprintf("build (%s)", strings.Join(parts, ", "))
	}

	return "build"
}

// FormatOptions formats a list of game options
func (f *OptionFormatter) FormatOptions(options []*v1.GameOption, allPaths *v1.AllPaths) []string {
	var formatted []string
	for _, option := range options {
		formatted = append(formatted, f.FormatOption(option, allPaths))
	}
	return formatted
}

// FormatOptionsNumbered formats options with numbers for menu selection
func (f *OptionFormatter) FormatOptionsNumbered(options []*v1.GameOption, allPaths *v1.AllPaths) []string {
	var formatted []string
	for i, option := range options {
		optionStr := f.FormatOption(option, allPaths)
		// Add number prefix to first line, indent subsequent lines
		lines := strings.Split(optionStr, "\n")
		if len(lines) > 0 {
			lines[0] = fmt.Sprintf("%d. %s", i+1, lines[0])
			for j := 1; j < len(lines); j++ {
				// Preserve existing indentation for multi-line options
				if !strings.HasPrefix(lines[j], "   ") {
					lines[j] = "   " + lines[j]
				}
			}
			formatted = append(formatted, strings.Join(lines, "\n"))
		}
	}
	return formatted
}
