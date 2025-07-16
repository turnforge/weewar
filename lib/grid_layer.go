package weewar

import (
	"fmt"
	"image/color"
	"image/draw"
	"math"
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
	minX, minY, _, _, _, _, _, _, startingCoord, startingX := world.Map.GetMapBounds(options.TileWidth, options.TileHeight, options.YIncrement)

	y := options.ScrollY - minY
	startX := options.ScrollX - (minX + startingX)
	height := float64(gl.height)
	width := float64(gl.width)
	leftCoord := startingCoord.Neighbor(LEFT)
	for i := 0; ; i++ {
		currX := startX
		if i%2 == 1 {
			currX = startX + options.TileWidth/2.0
		}
		rowCoord := leftCoord
		for ; currX < width; currX += options.TileWidth {
			fmt.Printf("currX, currY, Coord: ", currX, y, rowCoord)
			// Draw grid lines if enabled
			if options.ShowGrid {
				gl.drawHexGrid(currX, y, options)
			}

			// Draw coordinates if enabled
			if options.ShowCoordinates {
				gl.drawCoordinates(rowCoord, currX, y, options)
			}
			rowCoord = rowCoord.Neighbor(RIGHT)
		}

		if i%2 == 0 {
			leftCoord = leftCoord.Neighbor(BOTTOM_RIGHT)
		} else {
			leftCoord = leftCoord.Neighbor(BOTTOM_LEFT)
		}
		y += options.YIncrement
		if y >= height {
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
	vertices := gl.getHexVertices(centerX, centerY, options.TileWidth, options.TileHeight)

	// Draw lines between vertices
	gridColor := color.RGBA{R: 64, G: 64, B: 64, A: 255} // Dark gray
	bufferImg := gl.buffer.GetImageData()

	for i := 0; i < len(vertices); i++ {
		x1, y1 := vertices[i][0], vertices[i][1]
		x2, y2 := vertices[(i+1)%len(vertices)][0], vertices[(i+1)%len(vertices)][1]

		gl.drawLine(bufferImg, int(x1), int(y1), int(x2), int(y2), gridColor)
	}
}

// drawCoordinates draws Q,R coordinates in the center of a hex
func (gl *GridLayer) drawCoordinates(coord CubeCoord, centerX, centerY float64, options LayerRenderOptions) {
	// Simple text rendering - draw coordinate text
	text := fmt.Sprintf("%d,%d", coord.Q, coord.R)

	// For now, draw a simple representation (can be enhanced with proper text rendering)
	gl.drawSimpleText(text, centerX, centerY)
}

// getHexVertices returns the vertices of a hexagon centered at (centerX, centerY)
func (gl *GridLayer) getHexVertices(centerX, centerY, tileWidth, tileHeight float64) [][2]float64 {
	// Hexagon vertices (flat-top orientation)
	vertices := make([][2]float64, 6)

	// Use actual tile dimensions for proper hexagon shape
	radiusX := tileWidth / 2
	radiusY := tileHeight / 2

	// Hexagon angles (flat-top)
	for i := 0; i < 6; i++ {
		angle := float64(i) * 60.0 * 3.14159 / 180.0 // Convert to radians
		vertices[i][0] = centerX + radiusX*math.Cos(angle)
		vertices[i][1] = centerY + radiusY*math.Sin(angle)
	}

	return vertices
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
