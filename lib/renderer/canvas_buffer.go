//go:build js && wasm
// +build js,wasm

package rendering

import (
	"fmt"
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
	canvasElement js.Value               // HTML canvas element
	renderer      *htmlcanvas.HTMLCanvas // HTML canvas renderer
	context       *canvas.Context        // canvas drawing context
	width         int
	height        int
	canvasID      string
}

// NewCanvasDrawable creates a Drawable that renders to the specified HTML canvas
func NewCanvasDrawable(canvasID string) func(width, height int) Drawable {
	return func(width, height int) Drawable {
		return NewCanvasBuffer(canvasID, width, height)
	}
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
		fmt.Printf("DEBUG: DrawImage called with nil image\n")
		return
	}

	// For DrawImage, we want direct coordinate conversion without Y-axis flipping
	// because the image data is already in the correct orientation
	canvasX := x / PixelsPerMM
	canvasY := y / PixelsPerMM

	// fmt.Printf("DEBUG: DrawImage called with buffer coords (%.2f, %.2f) -> canvas coords (%.2f, %.2f), image size: %dx%d\n", x, y, canvasX, canvasY, img.Bounds().Dx(), img.Bounds().Dy())

	// Draw the actual image using the canvas context
	// Note: tdewolff/canvas DrawImage uses image's natural size, scaling handled by resolution
	cb.context.DrawImage(canvasX, canvasY, img, canvas.Resolution(PixelsPerMM))

	// fmt.Printf("DEBUG: DrawImage completed\n")
}

// DrawTextDirect uses the HTML Canvas 2D API directly for text rendering
// This bypasses the tdewolff/canvas font system and uses browser native text rendering
func (cb *CanvasBuffer) DrawTextDirect(x, y float64, text string, fontSize float64, color Color) {
	if text == "" {
		return
	}

	// Get the 2D context from the canvas element
	ctx := cb.canvasElement.Call("getContext", "2d")
	if ctx.IsUndefined() {
		return
	}

	// Set font properties
	fontSpec := fmt.Sprintf("%.0fpx Arial", fontSize)
	ctx.Set("font", fontSpec)
	ctx.Set("textAlign", "center")
	ctx.Set("textBaseline", "middle")

	// Set text color
	colorStr := fmt.Sprintf("rgba(%d, %d, %d, %.2f)", color.R, color.G, color.B, float64(color.A)/255.0)
	ctx.Set("fillStyle", colorStr)

	// Draw the text directly to the canvas
	ctx.Call("fillText", text, x, y)
}
