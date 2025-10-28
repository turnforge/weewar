package rendering

import "image"

// =============================================================================
// Drawable Interface - Common interface for Buffer and CanvasBuffer
// =============================================================================

// Drawable represents a drawable surface (either PNG buffer or HTML canvas)
type Drawable interface {
	Clear()
	Size() (width, height float64)
	FillPath(points []Point, fillColor Color)
	StrokePath(points []Point, strokeColor Color, strokeProperties StrokeProperties)
	DrawText(x, y float64, text string, fontSize float64, textColor Color)
	DrawTextWithStyle(x, y float64, text string, fontSize float64, textColor Color, bold bool, backgroundColor Color)
	DrawImage(x, y, width, height float64, img image.Image)
}
