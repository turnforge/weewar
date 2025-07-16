package weewar

import (
	"fmt"
	"image"
	"image/draw"
	"time"
)

// =============================================================================
// LayeredRenderer - Coordinates Multiple Layers
// =============================================================================

// LayeredRenderer coordinates multiple rendering layers
type LayeredRenderer struct {
	// Drawing target
	drawable Drawable
	x        int
	y        int
	width    int
	height   int

	// Rendering layers (in order)
	layers []Layer

	// Output buffer for compositing
	outputBuffer *Buffer

	// Batching system
	batchTimer    *time.Timer
	batchInterval time.Duration
	renderPending bool

	// Rendering parameters
	renderOptions LayerRenderOptions

	// Current world reference
	currentWorld *World
}

// NewLayeredRenderer creates a new layered renderer with default tile dimensions
func NewLayeredRenderer(drawable Drawable, width, height int) (*LayeredRenderer, error) {
	return NewLayeredRendererWithTileSize(drawable, width, height, DefaultTileWidth, DefaultTileHeight, DefaultYIncrement)
}

// NewLayeredRendererWithTileSize creates a new layered renderer with specified tile dimensions
func NewLayeredRendererWithTileSize(drawable Drawable, width, height int, tileWidth, tileHeight, yIncrement float64) (*LayeredRenderer, error) {
	// Create output buffer for compositing
	outputBuffer := NewBuffer(width, height)
	outputBuffer.Clear()

	renderer := &LayeredRenderer{
		drawable:      drawable,
		width:         width,
		height:        height,
		outputBuffer:  outputBuffer,
		batchInterval: 30 * time.Millisecond, // 33 FPS for prototyping
		renderPending: false,
		renderOptions: LayerRenderOptions{
			TileWidth:       tileWidth,
			TileHeight:      tileHeight,
			YIncrement:      yIncrement,
			ScrollX:         0,
			ScrollY:         0,
			ShowGrid:        true,
			ShowCoordinates: true,
		},
	}

	return renderer, nil
}

// SetWorld updates the current world reference
func (r *LayeredRenderer) SetWorld(w *World) {
	r.currentWorld = w
	// Mark all layers as dirty when world changes
	for _, layer := range r.layers {
		layer.MarkAllDirty()
	}
}

// SetAssetProvider updates the asset provider for all layers
func (r *LayeredRenderer) SetAssetProvider(provider AssetProvider) {
	for _, layer := range r.layers {
		layer.SetAssetProvider(provider)
	}
}

// SetTileDimensions updates the tile rendering dimensions
func (r *LayeredRenderer) SetTileDimensions(tileWidth, tileHeight, yIncrement float64) {
	r.renderOptions.TileWidth = tileWidth
	r.renderOptions.TileHeight = tileHeight
	r.renderOptions.YIncrement = yIncrement

	// Mark all layers as dirty since dimensions changed
	for _, layer := range r.layers {
		layer.MarkAllDirty()
	}
}

// SetScroll updates the viewport scroll offset
func (r *LayeredRenderer) SetScroll(scrollX, scrollY float64) {
	r.renderOptions.ScrollX = scrollX
	r.renderOptions.ScrollY = scrollY

	// Mark all layers as dirty since viewport changed
	for _, layer := range r.layers {
		layer.MarkAllDirty()
	}
}

// ScheduleRender allows layers to request a render update
func (r *LayeredRenderer) ScheduleRender() {
	r.scheduleRender()
}

// scheduleRender schedules a batched render update
func (r *LayeredRenderer) scheduleRender() {
	if r.renderPending {
		return // Already scheduled
	}

	r.renderPending = true

	// Cancel existing timer
	if r.batchTimer != nil {
		r.batchTimer.Stop()
	}

	// Schedule new render
	r.batchTimer = time.AfterFunc(r.batchInterval, func() {
		r.performRender()
		r.renderPending = false
	})
}

// ForceRender immediately renders all dirty layers (for synchronous updates)
func (r *LayeredRenderer) ForceRender() {
	fmt.Printf("LayeredRenderer.ForceRender called - terrain dirty: %d, units dirty: %d, UI dirty: %v\n")

	if r.batchTimer != nil {
		r.batchTimer.Stop()
	}
	r.performRender()
	r.renderPending = false
	fmt.Printf("DEBUG: ForceRender() completed successfully\n")
}

// performRender executes the actual rendering of dirty layers
func (r *LayeredRenderer) performRender() {
	fmt.Printf("LayeredRenderer.performRender called\n")

	for _, layer := range r.layers {
		if layer.IsDirty() {
			layer.Render(r.currentWorld, r.renderOptions)
		}
	}

	// Composite all layers to main canvas
	fmt.Printf("Compositing layers to main canvas\n")
	r.composite()
	fmt.Printf("DEBUG: performRender() completed successfully\n")
}

// composite just marks that layers need to be blitted
func (r *LayeredRenderer) composite() {
	// No complex compositing - just signal that buffers are ready for blitting
}

// blendBuffers blends src buffer onto dst buffer with alpha blending
func (r *LayeredRenderer) blendBuffers(dst, src *Buffer) {
	dstImg := dst.GetImageData()
	srcImg := src.GetImageData()

	// Use Go's image/draw for proper alpha blending
	draw.Draw(dstImg, dstImg.Bounds(), srcImg, image.Point{}, draw.Over)
}

// Resize updates the layer buffer sizes
func (r *LayeredRenderer) SetViewPort(x, y, width, height int) error {
	fmt.Printf("LayeredRenderer.SetViewPort called with: x=%d, y=%d, width=%d, height=%d\n", x, y, width, height)
	r.x = x
	r.y = y
	r.width = width
	r.height = height

	// Recreate all layer buffers with new size
	for _, layer := range r.layers {
		fmt.Printf("LayeredRenderer updating layer: %s\n", layer.GetName())
		layer.SetViewPort(x, y, width, height)
	}
	return nil
}
