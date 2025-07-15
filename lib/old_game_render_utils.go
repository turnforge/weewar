package weewar

import "fmt"

// RenderTerrain renders the terrain tiles to a buffer
func (g *Game) RenderTerrain(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	g.RenderTerrainTo(buffer, tileWidth, tileHeight, yIncrement)
}

// RenderTerrainTo renders the terrain tiles to any drawable surface
func (g *Game) RenderTerrainTo(drawable Drawable, tileWidth, tileHeight, yIncrement float64) {
	if g.Map == nil {
		return
	}

	// Render terrain tiles
	for _, tile := range g.Map.Tiles {
		if tile != nil {
			// Calculate tile position
			x, y := g.Map.XYForTile(tile.Row, tile.Col, tileWidth, tileHeight, yIncrement)

			// Try to load real tile asset first
			if g.assetProvider != nil && g.assetProvider.HasTileAsset(tile.TileType) {
				if tileImg, err := g.assetProvider.GetTileImage(tile.TileType); err == nil {
					// Render real tile image (XYForTile already returns centered coordinates)
					drawable.DrawImage(x-tileWidth/2, y-tileHeight/2, tileWidth, tileHeight, tileImg)
					continue
				}
			}

			// Fallback to colored hexagon if asset not available
			hexPath := g.createHexagonPath(x, y, tileWidth, tileHeight, yIncrement)
			tileColor := g.getTerrainColor(tile.TileType)
			drawable.FillPath(hexPath, tileColor)

			// Add border
			borderColor := Color{R: 100, G: 100, B: 100, A: 100}
			strokeProps := StrokeProperties{Width: 1.0, LineCap: "round", LineJoin: "round"}
			drawable.StrokePath(hexPath, borderColor, strokeProps)
		}
	}
}

// RenderUnitsTo renders units to any drawable surface using AssetManager
func (g *Game) RenderUnitsTo(drawable Drawable, tileWidth, tileHeight, yIncrement float64) {
	if g.Map == nil {
		return
	}

	// Define colors for different players
	playerColors := []Color{
		{R: 255, G: 0, B: 0, A: 255},   // Player 0 - red
		{R: 0, G: 0, B: 255, A: 255},   // Player 1 - blue
		{R: 0, G: 255, B: 0, A: 255},   // Player 2 - green
		{R: 255, G: 255, B: 0, A: 255}, // Player 3 - yellow
		{R: 255, G: 0, B: 255, A: 255}, // Player 4 - magenta
		{R: 0, G: 255, B: 255, A: 255}, // Player 5 - cyan
	}

	// Render units for each player
	for playerID, units := range g.Units {
		for _, unit := range units {
			if unit != nil {
				// Calculate unit position (same as tile position)
				x, y := g.Map.XYForTile(unit.Row, unit.Col, tileWidth, tileHeight, yIncrement)

				// Try to load real unit sprite first
				if g.assetProvider != nil && g.assetProvider.HasUnitAsset(unit.UnitType, playerID) {
					if unitImg, err := g.assetProvider.GetUnitImage(unit.UnitType, playerID); err == nil {
						// Render real unit sprite (XYForTile already returns centered coordinates)
						drawable.DrawImage(x-tileWidth/2, y-tileHeight/2, tileWidth, tileHeight, unitImg)

						// Add health indicator if unit is damaged
						if unit.AvailableHealth < 100 {
							g.renderHealthBarTo(drawable, x, y, tileWidth, tileHeight, unit.AvailableHealth, 100)
						}
						continue
					}
				}

				// Fallback to colored circle if asset not available
				var unitColor Color
				if playerID < len(playerColors) {
					unitColor = playerColors[playerID]
				} else {
					unitColor = Color{R: 128, G: 128, B: 128, A: 255} // Default gray
				}

				// Make unit hex slightly smaller than terrain hex
				unitHexPath := g.createHexagonPath(x, y, tileWidth*0.8, tileHeight*0.8, yIncrement*0.8)

				// Fill unit hex
				drawable.FillPath(unitHexPath, unitColor)

				// Draw unit border
				borderColor := Color{R: 255, G: 255, B: 255, A: 255} // White border for contrast
				strokeProps := StrokeProperties{Width: 2.0, LineCap: "round", LineJoin: "round"}
				drawable.StrokePath(unitHexPath, borderColor, strokeProps)

				// Add unit type text overlay (since no UnitID field exists)
				unitIDText := fmt.Sprintf("U%d", unit.UnitType)
				textColor := Color{R: 255, G: 255, B: 255, A: 255}
				fontSize := tileWidth / 6
				if fontSize < 8 {
					fontSize = 8
				}
				drawable.DrawTextWithStyle(x-10, y, unitIDText, fontSize, textColor, true, Color{R: 0, G: 0, B: 0, A: 128})

				// Add health indicator if unit is damaged
				if unit.AvailableHealth < 100 {
					g.renderHealthBarTo(drawable, x, y, tileWidth, tileHeight, unit.AvailableHealth, 100)
				}
			}
		}
	}
}

// RenderUnits renders the units to a buffer (delegates to RenderUnitsTo)
func (g *Game) RenderUnits(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	g.RenderUnitsTo(buffer, tileWidth, tileHeight, yIncrement)
}

// RenderUITo renders UI elements to any drawable surface
func (g *Game) RenderUITo(drawable Drawable, tileWidth, tileHeight, yIncrement float64) {
	// Create a simple indicator for current player
	indicatorSize := 20.0

	// Get current player color
	playerColors := []Color{
		{R: 255, G: 0, B: 0, A: 200},   // Player 0 - red
		{R: 0, G: 0, B: 255, A: 200},   // Player 1 - blue
		{R: 0, G: 255, B: 0, A: 200},   // Player 2 - green
		{R: 255, G: 255, B: 0, A: 200}, // Player 3 - yellow
		{R: 255, G: 0, B: 255, A: 200}, // Player 4 - magenta
		{R: 0, G: 255, B: 255, A: 200}, // Player 5 - cyan
	}

	var currentPlayerColor Color
	if g.CurrentPlayer < len(playerColors) {
		currentPlayerColor = playerColors[g.CurrentPlayer]
	} else {
		currentPlayerColor = Color{R: 255, G: 255, B: 255, A: 200} // Default white
	}

	// Create indicator rectangle in top-left corner
	indicatorPath := []Point{
		{X: 5, Y: 5},
		{X: 5 + indicatorSize, Y: 5},
		{X: 5 + indicatorSize, Y: 5 + indicatorSize},
		{X: 5, Y: 5 + indicatorSize},
	}

	// Fill indicator with current player color
	drawable.FillPath(indicatorPath, currentPlayerColor)

	// Add border
	borderColor := Color{R: 0, G: 0, B: 0, A: 255}
	strokeProps := StrokeProperties{Width: 2.0, LineCap: "round", LineJoin: "round"}
	drawable.StrokePath(indicatorPath, borderColor, strokeProps)
}

// RenderUI renders UI elements to a buffer (delegates to RenderUITo)
func (g *Game) RenderUI(buffer *Buffer, tileWidth, tileHeight, yIncrement float64) {
	g.RenderUITo(buffer, tileWidth, tileHeight, yIncrement)
}

// renderHealthBarTo renders a health bar for any drawable (generic version)
func (g *Game) renderHealthBarTo(drawable Drawable, x, y, tileWidth, tileHeight float64, currentHealth, maxHealth int) {
	if currentHealth >= maxHealth {
		return // Don't render health bar for full health
	}

	// Calculate health bar dimensions
	barWidth := tileWidth * 0.8
	barHeight := 6.0
	barX := x - barWidth/2
	barY := y + tileHeight/2 - barHeight - 2

	// Background bar (red)
	backgroundBar := []Point{
		{X: barX, Y: barY},
		{X: barX + barWidth, Y: barY},
		{X: barX + barWidth, Y: barY + barHeight},
		{X: barX, Y: barY + barHeight},
	}
	redColor := Color{R: 255, G: 0, B: 0, A: 255}
	drawable.FillPath(backgroundBar, redColor)

	// Health bar (green, proportional to health)
	healthRatio := float64(currentHealth) / float64(maxHealth)
	healthWidth := barWidth * healthRatio
	healthBar := []Point{
		{X: barX, Y: barY},
		{X: barX + healthWidth, Y: barY},
		{X: barX + healthWidth, Y: barY + barHeight},
		{X: barX, Y: barY + barHeight},
	}
	greenColor := Color{R: 0, G: 255, B: 0, A: 255}
	drawable.FillPath(healthBar, greenColor)

	// Health bar border
	borderColor := Color{R: 0, G: 0, B: 0, A: 255}
	strokeProps := StrokeProperties{Width: 1.0, LineCap: "round", LineJoin: "round"}
	drawable.StrokePath(backgroundBar, borderColor, strokeProps)
}

// createHexagonPath creates a hexagon path for a tile
func (g *Game) createHexagonPath(x, y, tileWidth, tileHeight, yIncrement float64) []Point {
	return []Point{
		{X: x + tileWidth/2, Y: y},                         // Top
		{X: x + tileWidth, Y: y + tileHeight - yIncrement}, // Top-right
		{X: x + tileWidth, Y: y + yIncrement},              // Bottom-right
		{X: x + tileWidth/2, Y: y + tileHeight},            // Bottom
		{X: x, Y: y + yIncrement},                          // Bottom-left
		{X: x, Y: y + tileHeight - yIncrement},             // Top-left
	}
}

// createUnitCircle creates a circular path for a unit
func (g *Game) createUnitCircle(x, y, tileWidth, tileHeight float64) []Point {
	// Create a circular approximation using polygon
	centerX := x + tileWidth/2
	centerY := y + tileHeight/2
	radius := minFloat(tileWidth, tileHeight) * 0.3 // Unit size relative to tile

	segments := 12
	points := make([]Point, segments)

	for i := 0; i < segments; i++ {
		angle := 2 * 3.14159 * float64(i) / float64(segments)
		unitX := centerX + radius*approximateCos(angle)
		unitY := centerY + radius*approximateSin(angle)
		points[i] = Point{X: unitX, Y: unitY}
	}

	return points
}

// renderHealthBar renders a health bar for a unit
func (g *Game) renderHealthBar(buffer *Buffer, x, y, tileWidth, tileHeight float64, currentHealth, maxHealth int) {
	if currentHealth >= maxHealth {
		return // Don't render health bar for full health
	}

	// Calculate health bar dimensions
	barWidth := tileWidth * 0.8
	barHeight := 6.0
	barX := x + (tileWidth-barWidth)/2
	barY := y + tileHeight - barHeight - 2

	// Background bar (red)
	backgroundBar := []Point{
		{X: barX, Y: barY},
		{X: barX + barWidth, Y: barY},
		{X: barX + barWidth, Y: barY + barHeight},
		{X: barX, Y: barY + barHeight},
	}
	buffer.FillPath(backgroundBar, Color{R: 255, G: 0, B: 0, A: 200})

	// Health bar (green)
	healthPercent := float64(currentHealth) / float64(maxHealth)
	healthBarWidth := barWidth * healthPercent

	if healthBarWidth > 0 {
		healthBar := []Point{
			{X: barX, Y: barY},
			{X: barX + healthBarWidth, Y: barY},
			{X: barX + healthBarWidth, Y: barY + barHeight},
			{X: barX, Y: barY + barHeight},
		}
		buffer.FillPath(healthBar, Color{R: 0, G: 255, B: 0, A: 200})
	}
}

// renderUnitText renders unit ID and health text overlay on PNG output
func (g *Game) renderUnitText(buffer *Buffer, unit *Unit, x, y, tileWidth, tileHeight float64) {
	// Get unit ID
	unitID := g.GetUnitID(unit)

	// Render unit ID below the unit with bold font and dark background
	idTextColor := Color{R: 255, G: 255, B: 255, A: 255} // White text for visibility
	idBackgroundColor := Color{R: 0, G: 0, B: 0, A: 180} // Semi-transparent black background
	idFontSize := 28.0                                   // Large font size for readability
	idX := x - 15                                        // Slightly left of center
	idY := y + (tileHeight * 0.4)                        // Below the unit
	buffer.DrawTextWithStyle(idX, idY, unitID, idFontSize, idTextColor, true, idBackgroundColor)

	// Render health with bold font and dark background (upper right)
	healthText := fmt.Sprintf("%d", unit.AvailableHealth)
	healthTextColor := Color{R: 255, G: 255, B: 0, A: 255}   // Yellow text for better visibility
	healthBackgroundColor := Color{R: 0, G: 0, B: 0, A: 180} // Semi-transparent black background
	healthFontSize := 22.0                                   // Large font for health
	healthX := x + 15                                        // Upper right area
	healthY := y - (tileHeight * 0.3)                        // Above center
	buffer.DrawTextWithStyle(healthX, healthY, healthText, healthFontSize, healthTextColor, true, healthBackgroundColor)
}

// getTerrainColor returns color for terrain type
func (g *Game) getTerrainColor(terrainType int) Color {
	switch terrainType {
	case 1: // Grass
		return Color{R: 50, G: 150, B: 50, A: 255}
	case 2: // Desert
		return Color{R: 200, G: 180, B: 100, A: 255}
	case 3: // Water
		return Color{R: 50, G: 50, B: 200, A: 255}
	case 4: // Mountain
		return Color{R: 150, G: 100, B: 50, A: 255}
	case 5: // Rock
		return Color{R: 150, G: 150, B: 150, A: 255}
	default: // Unknown
		return Color{R: 200, G: 200, B: 200, A: 255}
	}
}
