package weewar

import (
	"fmt"
	"image/color"
	"image/draw"
)

// =============================================================================
// GridLayer - Grid Line and Coordinate Rendering
// =============================================================================

// NewGridLayer creates a new grid layer
func NewGridLayer(width, height int, scheduler LayerScheduler) *GridLayer {
	return &GridLayer{
		BaseLayer: NewBaseLayer("grid", width, height, scheduler),
	}
}

// Render renders hex grid lines and coordinates
func (gl *GridLayer) Render(world *World, options LayerRenderOptions) {
	if world == nil || world.Map == nil {
		return
	}

	// Only render if grid or coordinates are enabled
	if !options.ShowGrid && !options.ShowCoordinates {
		if !gl.allDirty {
			return // Nothing to render
		}
		// Clear buffer if switching from visible to hidden
		gl.buffer.Clear()
		gl.allDirty = false
		gl.ClearDirty()
		return
	}

	// Clear buffer for full redraw (grid/coordinates are view-dependent)
	gl.buffer.Clear()

	// Get optimal starting coordinate and position from map bounds
	mapBounds := world.Map.GetMapBounds(options.TileWidth, options.TileHeight, options.YIncrement)
	minX := (options.TileWidth / 2) + mapBounds.MinX
	minY := mapBounds.MinY
	startingCoord := mapBounds.StartingCoord
	startingX := mapBounds.StartingX

	// Use viewport position as content offset
	viewportX := float64(gl.x)
	viewportY := float64(gl.y)
	TW2 := (options.TileWidth / 2.0)
	// startr, startc := HexToRowCol(startingCoord)
	// fmt.Println("ViewportX,Y: ", viewportX, viewportY, "StartingCoord: ", startingCoord, ", StartingCoordRC: ", startr, startc)
	startY := viewportY - minY // + (options.TileHeight / 2.0)
	startX := viewportX - (minX + startingX) + TW2
	startCoord := startingCoord
	// go left till we are < 0
	for startX > 0 {
		startCoord = startCoord.Neighbor(LEFT)
		startX -= options.TileWidth
	}
	// startr, startc = HexToRowCol(startCoord)
	// fmt.Println("AfterMoving Left: StartingCoord: ", startingCoord, ", StartingCoordRC: ", startr, startc, options.TileWidth)

	// Now go up till we are above y = 0
	for i := 0; startY > 0; i++ {
		if i%2 == 0 {
			startCoord = startCoord.Neighbor(TOP_LEFT)
			startX -= TW2
		} else {
			startCoord = startCoord.Neighbor(TOP_RIGHT)
			startX += TW2
		}
		startY -= options.YIncrement
	}

	height, width := float64(gl.height), float64(gl.width)
	// startr, startc = HexToRowCol(startCoord)
	// fmt.Printf("2. Here...., StartX, StartY: %f, %f, StartCoord: %s, StartRC: (%d, %d)\n", startX, startY, startCoord, startr, startc)

	// Now we can start drawing it
	currY := startY
	// fmt.Printf("w,h: ", width, height)
	for i := 0; ; i++ {
		// Draw a row first
		currCoord := startCoord
		currX := startX
		for ; currX < width; currX += options.TileWidth {
			// Draw grid lines if enabled
			// r, c := HexToRowCol(currCoord)
			// fmt.Printf("3. currXcurrY: (%f, %f), CurrCorrd: %s, RowCol: %d,%d, ShowGrid: %t\n", currX, currY, currCoord, r, c, options.ShowGrid)
			if options.ShowGrid {
				gl.drawHexGrid(currX, currY, options)
			}

			// Draw coordinates if enabled
			if options.ShowCoordinates {
				gl.drawCoordinates(currCoord, currX, currY, options)
			}
			currCoord = currCoord.Neighbor(RIGHT)
		}

		if i%2 == 0 {
			startCoord = startCoord.Neighbor(BOTTOM_LEFT)
			startX -= TW2
		} else {
			startCoord = startCoord.Neighbor(BOTTOM_RIGHT)
			startX += TW2
		}

		currY += options.YIncrement
		if currY >= height {
			// out of bounds so stop
			break
		}
	}

	// Mark as clean
	gl.allDirty = false
	gl.ClearDirty()
}

// drawHexGrid draws hexagonal grid lines around a tile
func (gl *GridLayer) drawHexGrid(centerX, centerY float64, options LayerRenderOptions) {
	// Get hexagon vertices
	vertices := gl.GetHexVertices(centerX, centerY, options.TileWidth, options.TileHeight)

	// Draw lines between vertices
	gridColor := color.RGBA{R: 128, G: 128, B: 128, A: 255} // Dark gray
	bufferImg := gl.buffer.GetImageData()

	for i := range len(vertices) {
		x1, y1 := vertices[i][0], vertices[i][1]
		x2, y2 := vertices[(i+1)%len(vertices)][0], vertices[(i+1)%len(vertices)][1]

		gl.drawLine(bufferImg, int(x1), int(y1), int(x2), int(y2), gridColor)
	}
}

// drawCoordinates draws Q,R coordinates in the center of a hex
func (gl *GridLayer) drawCoordinates(coord AxialCoord, centerX, centerY float64, options LayerRenderOptions) {
	// Format coordinate text
	text := fmt.Sprintf("%d,%d", coord.Q, coord.R)

	// Only draw text if it's within the visible area
	if centerX < 0 || centerY < 0 || centerX > float64(gl.width) || centerY > float64(gl.height) {
		return // Skip off-screen text
	}

	// Use the buffer's DrawText method with embedded font
	fontSize := 12.0
	textColor := Color{R: 20, G: 25, B: 25, A: 255}          // Black text for better visibility
	backgroundColor := Color{R: 255, G: 255, B: 255, A: 255} // Semi-transparent white background

	// Draw text at hex center with background
	gl.buffer.DrawTextWithStyle(centerX, centerY, text, fontSize, textColor, false, backgroundColor)

	// Log the coordinate for debugging
	// fmt.Printf("DEBUG: Drew coordinate text '%s' at (%.1f, %.1f)\n", text, centerX, centerY)
}

// drawLine draws a line between two points using Bresenham's algorithm
func (gl *GridLayer) drawLine(img draw.Image, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)

	x, y := x1, y1

	var xInc, yInc int
	if x1 < x2 {
		xInc = 1
	} else {
		xInc = -1
	}
	if y1 < y2 {
		yInc = 1
	} else {
		yInc = -1
	}

	var err int
	if dx > dy {
		err = dx / 2
		for x != x2 {
			if x >= 0 && y >= 0 && x < gl.width && y < gl.height {
				img.Set(x, y, c)
			}
			err -= dy
			if err < 0 {
				y += yInc
				err += dx
			}
			x += xInc
		}
	} else {
		err = dy / 2
		for y != y2 {
			if x >= 0 && y >= 0 && x < gl.width && y < gl.height {
				img.Set(x, y, c)
			}
			err -= dx
			if err < 0 {
				x += xInc
				err += dy
			}
			y += yInc
		}
	}
}

// drawSimpleText draws simple text at the given position
func (gl *GridLayer) drawSimpleText(text string, centerX, centerY float64) {
	// For now, draw simple dots to represent coordinates
	// This can be enhanced with proper text rendering later
	bufferImg := gl.buffer.GetImageData()
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

	x, y := int(centerX), int(centerY)

	// Draw a small cross or dot to indicate coordinates
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			px, py := x+i, y+j
			if px >= 0 && py >= 0 && px < gl.width && py < gl.height {
				if (i == 0 && abs(j) <= 2) || (j == 0 && abs(i) <= 2) {
					bufferImg.Set(px, py, textColor)
				}
			}
		}
	}
}
