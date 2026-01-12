package lib

import (
	"fmt"
	"strings"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// FormatPathCompact formats a path as a single line showing directions and key info
func FormatPathCompact(path *v1.Path) string {
	if path == nil || len(path.Edges) == 0 {
		return "direct"
	}

	var parts []string
	for i, edge := range path.Edges {
		fromCoord := AxialCoord{Q: int(edge.FromQ), R: int(edge.FromR)}
		toCoord := AxialCoord{Q: int(edge.ToQ), R: int(edge.ToR)}

		dir := GetDirection(fromCoord, toCoord)
		arrow := DirectionToString(dir)

		// Format: arrow (q,r,terrain,cost)
		part := fmt.Sprintf("%s (%d,%d,%s,%.0f)",
			arrow, edge.ToQ, edge.ToR,
			ShortenTerrain(edge.TerrainType), edge.MovementCost)
		parts = append(parts, part)

		// Limit to first 5 steps for very long paths
		if i >= 4 && len(path.Edges) > 6 {
			parts = append(parts, fmt.Sprintf("... +%d more", len(path.Edges)-5))
			break
		}
	}

	return strings.Join(parts, " ")
}

// FormatPathDetailed formats a path with one line per step, including explanations
func FormatPathDetailed(path *v1.Path, indent string) string {
	if path == nil || len(path.Edges) == 0 {
		return indent + "Direct adjacent move"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%sPath (total cost: %.0f):", indent, path.TotalCost))

	for i, edge := range path.Edges {
		fromCoord := AxialCoord{Q: int(edge.FromQ), R: int(edge.FromR)}
		toCoord := AxialCoord{Q: int(edge.ToQ), R: int(edge.ToR)}

		dir := GetDirection(fromCoord, toCoord)
		arrow := DirectionToString(dir)
		dirLong := DirectionToLongString(dir)

		// Format each step with arrow
		line := fmt.Sprintf("%s  %d. %s %s to (%d,%d) - %s (cost: %.0f)",
			indent, i+1, arrow, dirLong,
			edge.ToQ, edge.ToR, edge.TerrainType, edge.MovementCost)

		// Add explanation if available
		if edge.Explanation != "" {
			line += fmt.Sprintf(" - %s", edge.Explanation)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// ShortenTerrain shortens terrain names for compact display
func ShortenTerrain(terrain string) string {
	// Common terrain abbreviations
	switch strings.ToLower(terrain) {
	case "grass", "plains":
		return "grs"
	case "mountain", "mountains":
		return "mtn"
	case "forest", "woods":
		return "for"
	case "water", "ocean", "sea":
		return "wtr"
	case "desert", "sand":
		return "des"
	case "road", "path":
		return "rd"
	case "city", "town":
		return "cty"
	case "swamp", "marsh":
		return "swp"
	default:
		// Return first 3 chars if unknown
		if len(terrain) > 3 {
			return terrain[:3]
		}
		return terrain
	}
}
