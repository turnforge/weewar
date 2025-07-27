//go:build !js || !wasm
// +build !js !wasm

package rendering

import "image"

// Canvas functionality is disabled for non-WASM builds
// CanvasBuffer placeholder for non-WASM builds
type CanvasBuffer struct {
	canvasID string // Stub field for compatibility
}

// NewCanvasBuffer stub for non-WASM builds
func NewCanvasBuffer(canvasID string, width, height int) *CanvasBuffer {
	return nil // Canvas not supported in non-WASM builds
}

// Drawable methods for CanvasBuffer stub (should never be called)
func (cb *CanvasBuffer) Clear()                                   {}
func (cb *CanvasBuffer) Size() (width, height float64)            { return 0, 0 }
func (cb *CanvasBuffer) FillPath(points []Point, fillColor Color) {}
func (cb *CanvasBuffer) StrokePath(points []Point, strokeColor Color, strokeProperties StrokeProperties) {
}
func (cb *CanvasBuffer) DrawText(x, y float64, text string, fontSize float64, textColor Color) {}
func (cb *CanvasBuffer) DrawTextWithStyle(x, y float64, text string, fontSize float64, textColor Color, bold bool, backgroundColor Color) {
}
func (cb *CanvasBuffer) DrawImage(x, y, width, height float64, img image.Image) {}
