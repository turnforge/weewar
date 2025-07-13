//go:build !js || !wasm
// +build !js !wasm

package weewar

// =============================================================================
// CanvasRenderer Stub - Non-WASM Builds
// =============================================================================

// CanvasRenderer stub for non-WASM builds
// This ensures compilation works on all platforms while CanvasRenderer is only available for WASM
type CanvasRenderer struct {
	BaseRenderer
}

// NewCanvasRenderer creates a stub canvas renderer (non-functional on non-WASM platforms)
func NewCanvasRenderer() *CanvasRenderer {
	return &CanvasRenderer{}
}

// RenderWorld stub implementation (does nothing on non-WASM platforms)
func (cr *CanvasRenderer) RenderWorld(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// No-op for non-WASM builds
}

// RenderTerrain stub implementation
func (cr *CanvasRenderer) RenderTerrain(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// No-op for non-WASM builds
}

// RenderUnits stub implementation
func (cr *CanvasRenderer) RenderUnits(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// No-op for non-WASM builds
}

// RenderHighlights stub implementation
func (cr *CanvasRenderer) RenderHighlights(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// No-op for non-WASM builds
}

// RenderUI stub implementation
func (cr *CanvasRenderer) RenderUI(world *World, viewState *ViewState, drawable Drawable, options WorldRenderOptions) {
	// No-op for non-WASM builds
}