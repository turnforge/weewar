package rendering

import (
	"fmt"
	"image/color"
	"image/draw"
	"log"
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
	if world == nil {
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

	// Simple - get top left r/c and render row by row
	topLeftCoord := world.XYToQR(float64(gl.X), float64(gl.Y), options.TileWidth, options.TileHeight, options.YIncrement)
	bottomRightCoord := world.XYToQR(float64(gl.X+gl.Width), float64(gl.Y+gl.Height), options.TileWidth, options.TileHeight, options.YIncrement)
	tlrow, tlcol := HexToRowCol(topLeftCoord)
	brrow, brcol := HexToRowCol(bottomRightCoord)
	log.Println("TopLeft: ", topLeftCoord, tlrow, tlcol)
	log.Println("BottomRight: ", bottomRightCoord, brrow, brcol)
	for row := tlrow; row <= brrow; row++ {
		for col := tlcol; col <= brcol; col++ {
			coord := RowColToHex(row, col)
			currX, currY := world.CenterXYForTile(coord, options.TileWidth, options.TileHeight, options.YIncrement)
			currX -= float64(gl.X)
			currY -= float64(gl.Y)
			if options.ShowGrid {
				gl.drawHexGrid(currX, currY, options)
			}

			// Draw coordinates if enabled
			if true || options.ShowCoordinates {
				gl.drawCoordinates(coord, currX, currY, options)
			}
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
	if centerX < 0 || centerY < 0 || centerX > float64(gl.Width) || centerY > float64(gl.Height) {
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
			if x >= 0 && y >= 0 && x < gl.Width && y < gl.Height {
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
			if x >= 0 && y >= 0 && x < gl.Width && y < gl.Height {
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
