//go:build js && wasm
// +build js,wasm

package weewar

import (
	"image"
	"image/color"
	"syscall/js"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/htmlcanvas"
)

// =============================================================================
// CanvasBuffer - Direct HTML Canvas Rendering
// =============================================================================

// CanvasBuffer represents a drawable canvas that renders directly to HTML canvas
// Implements the Drawable interface to be a drop-in replacement for Buffer
type CanvasBuffer struct {
	canvasElement js.Value        // HTML canvas element
	renderer      *htmlcanvas.HTMLCanvas // HTML canvas renderer
	context       *canvas.Context // canvas drawing context
	width         int
	height        int
	canvasID      string
}

// NewCanvasBuffer creates a new canvas buffer that renders to the specified HTML canvas
func NewCanvasBuffer(canvasID string, width, height int) *CanvasBuffer {
	// Get canvas element from DOM
	canvasElement := js.Global().Get("document").Call("getElementById", canvasID)
	if canvasElement.IsUndefined() {
		return nil // Canvas element not found
	}

	// Create HTML canvas renderer with dimensions in mm
	canvasWidth := float64(width) / PixelsPerMM
	canvasHeight := float64(height) / PixelsPerMM
	renderer := htmlcanvas.New(canvasElement, canvasWidth, canvasHeight, PixelsPerMM)

	// Create drawing context
	context := canvas.NewContext(renderer)

	return &CanvasBuffer{
		canvasElement: canvasElement,
		renderer:      renderer,
		context:       context,
		width:         width,
		height:        height,
		canvasID:      canvasID,
	}
}

// Clear clears the HTML canvas
func (cb *CanvasBuffer) Clear() {
	// Recreate the renderer and context to clear everything
	canvasWidth := float64(cb.width) / PixelsPerMM
	canvasHeight := float64(cb.height) / PixelsPerMM
	cb.renderer = htmlcanvas.New(cb.canvasElement, canvasWidth, canvasHeight, PixelsPerMM)
	cb.context = canvas.NewContext(cb.renderer)
}

// Size returns the dimensions of the canvas buffer
func (cb *CanvasBuffer) Size() (width, height float64) {
	return float64(cb.width), float64(cb.height)
}

// bufferToCanvasX converts buffer X coordinate (pixels) to canvas X coordinate (mm)
func (cb *CanvasBuffer) bufferToCanvasX(x float64) float64 {
	return x / PixelsPerMM
}

// bufferToCanvasY converts buffer Y coordinate (pixels) to canvas Y coordinate (mm)
// Note: Also flips Y-axis (buffer top-left to canvas bottom-left) - same as Buffer
func (cb *CanvasBuffer) bufferToCanvasY(y float64) float64 {
	return (float64(cb.height) - y) / PixelsPerMM
}

// bufferToCanvasXY converts buffer coordinates (pixels) to canvas coordinates (mm)
func (cb *CanvasBuffer) bufferToCanvasXY(x, y float64) (float64, float64) {
	return cb.bufferToCanvasX(x), cb.bufferToCanvasY(y)
}

// getCanvasSize returns canvas dimensions in millimeters
func (cb *CanvasBuffer) getCanvasSize() (width, height float64) {
	return float64(cb.width) / PixelsPerMM, float64(cb.height) / PixelsPerMM
}

// Render flushes all drawing operations to the HTML canvas
func (cb *CanvasBuffer) Render() error {
	// Drawing operations are automatically rendered to the HTML canvas
	// through the htmlcanvas renderer, so this is essentially a no-op
	return nil
}

// FillPath fills a given path with the given color (same interface as Buffer)
func (cb *CanvasBuffer) FillPath(points []Point, fillColor Color) {
	if len(points) < 2 {
		return // Need at least 2 points to create a path
	}

	// Set fill color
	rgba := color.RGBA{R: fillColor.R, G: fillColor.G, B: fillColor.B, A: fillColor.A}
	cb.context.SetFillColor(rgba)

	// Build path using coordinate conversion helpers (same as Buffer)
	canvasX, canvasY := cb.bufferToCanvasXY(points[0].X, points[0].Y)
	cb.context.MoveTo(canvasX, canvasY)
	for i := 1; i < len(points); i++ {
		canvasX, canvasY := cb.bufferToCanvasXY(points[i].X, points[i].Y)
		cb.context.LineTo(canvasX, canvasY)
	}
	cb.context.Close()

	// Fill the path
	cb.context.Fill()
}

// StrokePath strokes a given path with a given color and stroke properties
func (cb *CanvasBuffer) StrokePath(points []Point, strokeColor Color, strokeProperties StrokeProperties) {
	if len(points) < 2 {
		return // Need at least 2 points to create a path
	}

	// Set stroke color
	rgba := color.RGBA{R: strokeColor.R, G: strokeColor.G, B: strokeColor.B, A: strokeColor.A}
	cb.context.SetStrokeColor(rgba)

	// Set stroke width (convert pixels to mm)
	cb.context.SetStrokeWidth(strokeProperties.Width / PixelsPerMM)

	// Set line cap
	switch strokeProperties.LineCap {
	case "round":
		cb.context.SetStrokeCapper(canvas.RoundCapper{})
	case "square":
		cb.context.SetStrokeCapper(canvas.SquareCapper{})
	default: // "butt" or unspecified
		cb.context.SetStrokeCapper(canvas.ButtCapper{})
	}

	// Set line join
	switch strokeProperties.LineJoin {
	case "round":
		cb.context.SetStrokeJoiner(canvas.RoundJoiner{})
	case "bevel":
		cb.context.SetStrokeJoiner(canvas.BevelJoiner{})
	default: // "miter" or unspecified
		cb.context.SetStrokeJoiner(canvas.MiterJoiner{})
	}

	// Set dash pattern if specified (convert pixels to mm)
	if len(strokeProperties.DashPattern) > 0 {
		scaledDashes := make([]float64, len(strokeProperties.DashPattern))
		for i, dash := range strokeProperties.DashPattern {
			scaledDashes[i] = dash / PixelsPerMM
		}
		cb.context.SetDashes(strokeProperties.DashOffset/PixelsPerMM, scaledDashes...)
	}

	// Build path using coordinate conversion helpers
	canvasX, canvasY := cb.bufferToCanvasXY(points[0].X, points[0].Y)
	cb.context.MoveTo(canvasX, canvasY)
	for i := 1; i < len(points); i++ {
		canvasX, canvasY := cb.bufferToCanvasXY(points[i].X, points[i].Y)
		cb.context.LineTo(canvasX, canvasY)
	}

	// Stroke the path
	cb.context.Stroke()
}

// DrawText renders text at the specified position with the given font size and color
func (cb *CanvasBuffer) DrawText(x, y float64, text string, fontSize float64, textColor Color) {
	cb.DrawTextWithStyle(x, y, text, fontSize, textColor, false, Color{})
}

// DrawTextWithStyle renders text with optional bold and background
func (cb *CanvasBuffer) DrawTextWithStyle(x, y float64, text string, fontSize float64, textColor Color, bold bool, backgroundColor Color) {
	if text == "" {
		return
	}

	// Load font family
	fontFamily := canvas.NewFontFamily("Arial")
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontRegular); err != nil {
		// Fallback to sans-serif if Arial not available
		fontFamily = canvas.NewFontFamily("sans-serif")
		fontFamily.LoadSystemFont("DejaVu Sans", canvas.FontRegular)
	}

	// Choose font weight
	fontWeight := canvas.FontRegular
	if bold {
		fontWeight = canvas.FontBold
	}

	// Set text color and create face
	rgba := color.RGBA{R: textColor.R, G: textColor.G, B: textColor.B, A: textColor.A}

	// Convert font size from pixels to mm
	face := fontFamily.Face(fontSize/PixelsPerMM, rgba, fontWeight, canvas.FontNormal)

	// Create text line for rendering
	textLine := canvas.NewTextLine(face, text, canvas.Left)

	// Convert buffer coordinates to canvas coordinates (once)
	canvasX, canvasY := cb.bufferToCanvasXY(x, y)

	// Draw background rectangle if specified
	if backgroundColor.A > 0 {
		// Add padding around text (in mm)
		padding := 2.0 / PixelsPerMM // Convert 2 pixels to mm

		// Get text bounds to position background properly
		bounds := textLine.Bounds()

		// Position background to properly contain text
		// Canvas DrawText positions text at baseline, estimate descender space
		textWidth := bounds.W()
		textHeight := bounds.H()

		bgX := canvasX - padding
		bgY := canvasY - (textHeight * 0.2) - padding // Account for descenders below baseline
		bgWidth := textWidth + (padding * 2)
		bgHeight := textHeight + (padding * 2)

		cb.context.SetFillColor(color.RGBA{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B, A: backgroundColor.A})
		cb.context.DrawPath(bgX, bgY, canvas.Rectangle(bgWidth, bgHeight))
		cb.context.Fill()
	}

	// Draw the text using converted coordinates
	cb.context.DrawText(canvasX, canvasY, textLine)
}

// DrawImage draws an image at the specified position
func (cb *CanvasBuffer) DrawImage(x, y, width, height float64, img image.Image) {
	if img == nil {
		return
	}
	
	// Convert buffer coordinates to canvas coordinates
	canvasX, canvasY := cb.bufferToCanvasXY(x, y)
	
	// Convert image to canvas-compatible format
	// For now, extract image data and draw as a filled rectangle with average color
	// This is a simplified implementation - in production we'd convert to ImageData
	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}
	
	// Sample the center pixel to get a representative color
	centerX := bounds.Min.X + bounds.Dx()/2
	centerY := bounds.Min.Y + bounds.Dy()/2
	colorSample := img.At(centerX, centerY)
	r, g, b, a := colorSample.RGBA()
	
	// Convert from 16-bit to 8-bit color values
	avgColor := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8), 
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
	
	// Draw as a filled rectangle with the sampled color
	cb.context.SetFillColor(avgColor)
	cb.context.DrawPath(canvasX, canvasY, canvas.Rectangle(width, height))
	cb.context.Fill()
}

// RenderToCanvasBuffer creates a canvas buffer-compatible version of RenderToBuffer
func (g *Game) RenderToCanvasBuffer(canvasBuffer *CanvasBuffer, tileWidth, tileHeight, yIncrement float64) error {
	// Use the new RenderTo method that works with any Drawable
	return g.RenderTo(canvasBuffer, tileWidth, tileHeight, yIncrement)
}

// renderMapToCanvas renders the game map directly to a canvas buffer
func (g *Game) renderMapToCanvas(canvasBuffer *CanvasBuffer, tileWidth, tileHeight, yIncrement float64) error {
	ctx := canvasBuffer.context

	// Convert pixel measurements to mm for canvas coordinates
	hexRadius := (tileWidth * 0.4) / PixelsPerMM
	
	// Render each tile
	for row := 0; row < g.Map.NumRows; row++ {
		for col := 0; col < g.Map.NumCols; col++ {
			// Calculate hex center position
			x := float64(col) * (tileWidth / PixelsPerMM)
			y := float64(row) * (yIncrement / PixelsPerMM)
			
			// Offset even rows for hex grid
			if row%2 == 0 {
				x += (tileWidth * 0.5) / PixelsPerMM
			}
			
			centerX := x + (tileWidth * 0.5) / PixelsPerMM
			centerY := y + (tileHeight * 0.5) / PixelsPerMM

			// Get tile at this position
			coord := g.Map.DisplayToHex(row, col)
			tile := g.Map.TileAtCube(coord)

			// Set color based on terrain type
			if tile != nil {
				switch tile.TileType {
				case 1: // Grass
					ctx.SetFillColor(color.RGBA{R: 34, G: 139, B: 34, A: 255})
				case 2: // Desert
					ctx.SetFillColor(color.RGBA{R: 238, G: 203, B: 173, A: 255})
				case 3: // Water
					ctx.SetFillColor(color.RGBA{R: 65, G: 105, B: 225, A: 255})
				case 4: // Mountain
					ctx.SetFillColor(color.RGBA{R: 139, G: 137, B: 137, A: 255})
				case 5: // Rock
					ctx.SetFillColor(color.RGBA{R: 105, G: 105, B: 105, A: 255})
				default:
					ctx.SetFillColor(color.RGBA{R: 200, G: 200, B: 200, A: 255})
				}
			} else {
				// Empty tile
				ctx.SetFillColor(color.RGBA{R: 220, G: 220, B: 220, A: 255})
			}

			// Create hexagon path
			hexPath := createHexagonPath(centerX, centerY, hexRadius)
			
			// Fill the hexagon
			ctx.DrawPath(0, 0, hexPath)
			ctx.Fill()

			// Draw border
			ctx.SetStrokeColor(color.RGBA{R: 0, G: 0, B: 0, A: 128})
			ctx.SetStrokeWidth(0.5 / PixelsPerMM)
			ctx.DrawPath(0, 0, hexPath)
			ctx.Stroke()
		}
	}

	return nil
}

// createHexagonPath creates a hexagon path centered at (cx, cy) with given radius
func createHexagonPath(cx, cy, radius float64) *canvas.Path {
	path := &canvas.Path{}
	
	// Create hexagon with 6 sides
	for i := 0; i < 6; i++ {
		// Angle for each vertex (60 degrees apart)
		angle := float64(i) * 60.0 * 3.14159 / 180.0
		x := cx + radius*cos(angle)
		y := cy + radius*sin(angle)
		
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
	
	return path
}

// createHexPoints creates points for a hexagon centered at (cx, cy) with given radius
func createHexPoints(cx, cy, radius float64) []Point {
	points := make([]Point, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * 60.0 * 3.14159 / 180.0 // Convert to radians
		x := cx + radius*cos(angle)
		y := cy + radius*sin(angle)
		points[i] = Point{X: x, Y: y}
	}
	return points
}

// createCirclePoints creates points for a circle approximation
func createCirclePoints(cx, cy, radius float64, segments int) []Point {
	points := make([]Point, segments)
	for i := 0; i < segments; i++ {
		angle := float64(i) * 360.0 / float64(segments) * 3.14159 / 180.0
		x := cx + radius*cos(angle)
		y := cy + radius*sin(angle)
		points[i] = Point{X: x, Y: y}
	}
	return points
}

// Simple math helpers (since we can't import math in WASM easily)
func cos(angle float64) float64 {
	// Simple cosine approximation using Taylor series
	// cos(x) ≈ 1 - x²/2! + x⁴/4! - x⁶/6!
	x := angle
	for x > 3.14159*2 {
		x -= 3.14159 * 2
	}
	for x < 0 {
		x += 3.14159 * 2
	}

	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

func sin(angle float64) float64 {
	// sin(x) = cos(x - π/2)
	return cos(angle - 3.14159/2)
}
