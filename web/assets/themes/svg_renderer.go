package themes

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// SVGWorldRenderer renders worlds using SVG assets (for fantasy/modern themes)
// Generates a single SVG file with:
// - <defs> section containing each unique tile/unit as a <symbol>
// - <use> elements positioned at each hex coordinate
type SVGWorldRenderer struct {
	theme     Theme
	assetRoot string // Root directory for SVG assets on filesystem

	// Cache for loaded SVG content
	tileCache  map[string]string // key: "tile_{type}_{player}" -> processed SVG content
	unitCache  map[string]string // key: "unit_{type}_{player}" -> processed SVG content
	cacheMutex sync.RWMutex
}

// NewSVGWorldRenderer creates a new SVG-based world renderer
func NewSVGWorldRenderer(theme Theme) (*SVGWorldRenderer, error) {
	info := theme.GetThemeInfo()
	if info == nil || info.AssetType != "svg" {
		return nil, fmt.Errorf("SVGWorldRenderer requires an SVG-based theme")
	}

	// Determine asset root from theme's base path
	// Theme basePath is like "/static/assets/themes/fantasy"
	// We need "web/static/assets/themes/fantasy" or relative "web" + basePath
	assetRoot := "web" + info.BasePath

	return &SVGWorldRenderer{
		theme:     theme,
		assetRoot: assetRoot,
		tileCache: make(map[string]string),
		unitCache: make(map[string]string),
	}, nil
}

// SetAssetRoot allows overriding the asset root directory
func (r *SVGWorldRenderer) SetAssetRoot(root string) {
	r.assetRoot = root
}

// Render produces a composite SVG of the world
func (r *SVGWorldRenderer) Render(tiles map[string]*v1.Tile, units map[string]*v1.Unit, options *lib.RenderOptions) ([]byte, string, error) {
	if options == nil {
		options = lib.DefaultRenderOptions()
	}

	if len(tiles) == 0 {
		return nil, "", fmt.Errorf("no tiles to render")
	}

	// Compute bounds
	minX, minY, width, height := computeBounds(tiles, units, options)

	// Track unique symbols we need to define
	tileSymbols := make(map[string]string) // symbolId -> SVG content
	unitSymbols := make(map[string]string)

	// Build the SVG document
	var svg bytes.Buffer

	// Start SVG with computed viewBox
	svg.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
`, width, height, width, height))

	// First pass: collect all unique symbols
	for _, tile := range tiles {
		symbolId, svgContent, err := r.getTileSymbol(tile.TileType, tile.Player, options)
		if err != nil {
			fmt.Printf("Warning: failed to load tile symbol for type %d: %v\n", tile.TileType, err)
			continue
		}
		tileSymbols[symbolId] = svgContent
	}

	for _, unit := range units {
		symbolId, svgContent, err := r.getUnitSymbol(unit.UnitType, unit.Player, options)
		if err != nil {
			fmt.Printf("Warning: failed to load unit symbol for type %d: %v\n", unit.UnitType, err)
			continue
		}
		unitSymbols[symbolId] = svgContent
	}

	// Write defs section with all symbols
	svg.WriteString("  <defs>\n")
	for symbolId, content := range tileSymbols {
		svg.WriteString(fmt.Sprintf("    <symbol id=\"%s\" viewBox=\"0 0 100 100\">\n", symbolId))
		svg.WriteString("      ")
		svg.WriteString(content)
		svg.WriteString("\n    </symbol>\n")
	}
	for symbolId, content := range unitSymbols {
		svg.WriteString(fmt.Sprintf("    <symbol id=\"%s\" viewBox=\"0 0 100 100\">\n", symbolId))
		svg.WriteString("      ")
		svg.WriteString(content)
		svg.WriteString("\n    </symbol>\n")
	}
	svg.WriteString("  </defs>\n\n")

	// Second pass: place tiles using <use> elements
	svg.WriteString("  <!-- Tiles -->\n")
	for _, tile := range tiles {
		symbolId := r.tileSymbolId(tile.TileType, tile.Player)
		if _, ok := tileSymbols[symbolId]; !ok {
			continue // Skip if symbol wasn't loaded
		}

		x, y := lib.HexToPixelInt32(tile.Q, tile.R, options)
		x -= minX
		y -= minY

		svg.WriteString(fmt.Sprintf("  <use href=\"#%s\" x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\"/>\n",
			symbolId, x, y, options.TileWidth, options.TileHeight))
	}

	// Third pass: place units on top
	svg.WriteString("\n  <!-- Units -->\n")
	for _, unit := range units {
		symbolId := r.unitSymbolId(unit.UnitType, unit.Player)
		if _, ok := unitSymbols[symbolId]; !ok {
			continue
		}

		x, y := lib.HexToPixelInt32(unit.Q, unit.R, options)
		x -= minX
		y -= minY

		// Units are slightly smaller (90% of tile) and centered within the tile
		unitWidth := int(float64(options.TileWidth) * 0.9)
		unitHeight := int(float64(options.TileHeight) * 0.9)
		useX := x + (options.TileWidth-unitWidth)/2
		useY := y + (options.TileHeight-unitHeight)/2

		svg.WriteString(fmt.Sprintf("  <use href=\"#%s\" x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\"/>\n",
			symbolId, useX, useY, unitWidth, unitHeight))
	}

	svg.WriteString("</svg>\n")

	return svg.Bytes(), "image/svg+xml", nil
}

// tileSymbolId generates a unique symbol ID for a tile type+player
func (r *SVGWorldRenderer) tileSymbolId(tileType, player int32) string {
	effectivePlayer := r.theme.GetEffectivePlayer(tileType, player)
	return fmt.Sprintf("tile_%d_%d", tileType, effectivePlayer)
}

// unitSymbolId generates a unique symbol ID for a unit type+player
func (r *SVGWorldRenderer) unitSymbolId(unitType, player int32) string {
	return fmt.Sprintf("unit_%d_%d", unitType, player)
}

// getTileSymbol loads and processes a tile SVG, returning symbol ID and inner content
func (r *SVGWorldRenderer) getTileSymbol(tileType, player int32, options *lib.RenderOptions) (string, string, error) {
	effectivePlayer := r.theme.GetEffectivePlayer(tileType, player)
	symbolId := r.tileSymbolId(tileType, player) // tileSymbolId also calls GetEffectivePlayer

	// Check cache
	r.cacheMutex.RLock()
	if content, ok := r.tileCache[symbolId]; ok {
		r.cacheMutex.RUnlock()
		return symbolId, content, nil
	}
	r.cacheMutex.RUnlock()

	// Load SVG file
	tilePath := r.theme.GetTilePath(tileType)
	if tilePath == "" {
		return "", "", fmt.Errorf("no tile path for type %d", tileType)
	}

	fullPath := filepath.Join(r.assetRoot, tilePath)
	svgContent, err := r.loadAndProcessSVG(fullPath, effectivePlayer)
	if err != nil {
		return "", "", err
	}

	// Cache it
	r.cacheMutex.Lock()
	r.tileCache[symbolId] = svgContent
	r.cacheMutex.Unlock()

	return symbolId, svgContent, nil
}

// getUnitSymbol loads and processes a unit SVG
func (r *SVGWorldRenderer) getUnitSymbol(unitType, player int32, options *lib.RenderOptions) (string, string, error) {
	symbolId := r.unitSymbolId(unitType, player)

	// Check cache
	r.cacheMutex.RLock()
	if content, ok := r.unitCache[symbolId]; ok {
		r.cacheMutex.RUnlock()
		return symbolId, content, nil
	}
	r.cacheMutex.RUnlock()

	// Load SVG file
	unitPath := r.theme.GetUnitPath(unitType)
	if unitPath == "" {
		return "", "", fmt.Errorf("no unit path for type %d", unitType)
	}

	fullPath := filepath.Join(r.assetRoot, unitPath)
	svgContent, err := r.loadAndProcessSVG(fullPath, player)
	if err != nil {
		return "", "", err
	}

	// Cache it
	r.cacheMutex.Lock()
	r.unitCache[symbolId] = svgContent
	r.cacheMutex.Unlock()

	return symbolId, svgContent, nil
}

// loadAndProcessSVG loads an SVG file and applies player colors
func (r *SVGWorldRenderer) loadAndProcessSVG(path string, player int32) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read SVG: %w", err)
	}

	svgStr := string(content)

	// Apply player colors by replacing the playerColor gradient stops
	colors := r.theme.GetPlayerColor(player)
	if colors != nil {
		svgStr = r.applyPlayerColors(svgStr, colors)
	}

	// Extract inner content (remove outer <svg> tag)
	svgStr = r.extractSVGContent(svgStr)

	return svgStr, nil
}

// applyPlayerColors replaces the playerColor gradient with player-specific colors
func (r *SVGWorldRenderer) applyPlayerColors(svg string, colors *v1.PlayerColor) string {
	// Match the linearGradient with id="playerColor" and replace stop colors
	// Pattern: <linearGradient id="playerColor">...<stop ... stop-color="..."/>...<stop ... stop-color="..."/>...</linearGradient>

	// Replace first stop color (primary)
	primaryPattern := regexp.MustCompile(`(<linearGradient[^>]*id="playerColor"[^>]*>[\s\S]*?<stop[^>]*offset="0%"[^>]*stop-color=")#[0-9a-fA-F]{6}(")`)
	svg = primaryPattern.ReplaceAllString(svg, "${1}"+colors.Primary+"${2}")

	// Replace second stop color (secondary)
	secondaryPattern := regexp.MustCompile(`(<linearGradient[^>]*id="playerColor"[^>]*>[\s\S]*?<stop[^>]*offset="100%"[^>]*stop-color=")#[0-9a-fA-F]{6}(")`)
	svg = secondaryPattern.ReplaceAllString(svg, "${1}"+colors.Secondary+"${2}")

	return svg
}

// extractSVGContent removes the outer <svg> wrapper and returns inner content
func (r *SVGWorldRenderer) extractSVGContent(svg string) string {
	// Remove <?xml...?> declaration if present
	svg = regexp.MustCompile(`<\?xml[^?]*\?>`).ReplaceAllString(svg, "")

	// Extract content between <svg...> and </svg>
	startPattern := regexp.MustCompile(`<svg[^>]*>`)
	endPattern := regexp.MustCompile(`</svg>`)

	startLoc := startPattern.FindStringIndex(svg)
	endLoc := endPattern.FindStringIndex(svg)

	if startLoc != nil && endLoc != nil && endLoc[0] > startLoc[1] {
		return strings.TrimSpace(svg[startLoc[1]:endLoc[0]])
	}

	return svg
}

// ClearCache clears the SVG cache
func (r *SVGWorldRenderer) ClearCache() {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	r.tileCache = make(map[string]string)
	r.unitCache = make(map[string]string)
}
