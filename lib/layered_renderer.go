package weewar

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"
)

// LayeredRenderer provides fast prototyping rendering with separate layers using WASM buffers
type LayeredRenderer struct {
	// Canvas target
	canvasID string
	width    int
	height   int

	// Layer buffers (WASM Buffer objects, not DOM canvases)
	terrainBuffer *Buffer // Static terrain tiles
	unitBuffer    *Buffer // Unit sprites
	uiBuffer      *Buffer // Selection/hover/UI

	// Dirty tracking for efficient updates
	dirtyTerrain map[CubeCoord]bool
	dirtyUnits   map[CubeCoord]bool
	dirtyUI      bool

	// Batching system
	batchTimer    *time.Timer
	batchInterval time.Duration
	renderPending bool

	// Tile rendering parameters (should match game.go rendering)
	tileWidth  float64
	tileHeight float64
	yIncrement float64

	// Asset cache (terrain and unit images)
	terrainSprites map[int]image.Image    // Cached terrain images
	unitSprites    map[string]image.Image // Cached unit images

	// Current map reference
	currentMap *Map

	// Asset provider for terrain and unit sprites
	assetProvider AssetProvider
}

// NewLayeredRenderer creates a new layered renderer with default tile dimensions
func NewLayeredRenderer(canvasID string, width, height int) (*LayeredRenderer, error) {
	return NewLayeredRendererWithTileSize(canvasID, width, height, 60.0, 52.0, 39.0)
}

// NewLayeredRendererWithTileSize creates a new layered renderer with specified tile dimensions
func NewLayeredRendererWithTileSize(canvasID string, width, height int, tileWidth, tileHeight, yIncrement float64) (*LayeredRenderer, error) {
	// Create WASM buffers for each layer instead of DOM canvases
	terrainBuffer := NewBuffer(width, height)
	unitBuffer := NewBuffer(width, height)
	uiBuffer := NewBuffer(width, height)

	// Clear all buffers to transparent
	terrainBuffer.Clear()
	unitBuffer.Clear()
	uiBuffer.Clear()

	renderer := &LayeredRenderer{
		canvasID:       canvasID,
		width:          width,
		height:         height,
		terrainBuffer:  terrainBuffer,
		unitBuffer:     unitBuffer,
		uiBuffer:       uiBuffer,
		dirtyTerrain:   make(map[CubeCoord]bool),
		dirtyUnits:     make(map[CubeCoord]bool),
		dirtyUI:        false,
		batchInterval:  30 * time.Millisecond, // 33 FPS for prototyping
		renderPending:  false,
		terrainSprites: make(map[int]image.Image),
		unitSprites:    make(map[string]image.Image),
		tileWidth:      tileWidth,
		tileHeight:     tileHeight,
		yIncrement:     yIncrement,
	}

	return renderer, nil
}

// SetMap updates the current map reference
func (r *LayeredRenderer) SetMap(m *Map) {
	r.currentMap = m
	// Mark all terrain as dirty when map changes
	r.MarkAllTerrainDirty()
	// Also mark all units as dirty since we have a new map
	r.MarkAllUnitsDirty()
}

// SetAssetProvider updates the asset provider for sprite rendering
func (r *LayeredRenderer) SetAssetProvider(provider AssetProvider) {
	fmt.Printf("SetAssetProvider called, clearing sprite caches and marking layers dirty\n")
	r.assetProvider = provider
	// Clear cached sprites since provider changed
	r.terrainSprites = make(map[int]image.Image)
	r.unitSprites = make(map[string]image.Image)
	// Mark all terrain as dirty to re-render with new sprites
	r.MarkAllTerrainDirty()
	// Also mark all units as dirty since unit sprites changed too
	r.MarkAllUnitsDirty()
}

// SetTileDimensions updates the tile rendering dimensions
func (r *LayeredRenderer) SetTileDimensions(tileWidth, tileHeight, yIncrement float64) {
	r.tileWidth = tileWidth
	r.tileHeight = tileHeight
	r.yIncrement = yIncrement
	// Mark all terrain as dirty since dimensions changed
	r.MarkAllTerrainDirty()
	r.MarkAllUnitsDirty()
}

// MarkTerrainDirty marks a specific tile as needing terrain update
func (r *LayeredRenderer) MarkTerrainDirty(coord CubeCoord) {
	r.dirtyTerrain[coord] = true
	r.scheduleRender()
}

// MarkUnitDirty marks a specific position as needing unit update
func (r *LayeredRenderer) MarkUnitDirty(coord CubeCoord) {
	r.dirtyUnits[coord] = true
	r.scheduleRender()
}

// MarkUIDirty marks the UI layer as needing update
func (r *LayeredRenderer) MarkUIDirty() {
	r.dirtyUI = true
	r.scheduleRender()
}

// MarkAllTerrainDirty marks entire terrain layer for rebuild
func (r *LayeredRenderer) MarkAllTerrainDirty() {
	// Clear and mark for full rebuild
	for coord := range r.dirtyTerrain {
		delete(r.dirtyTerrain, coord)
	}
	r.dirtyTerrain[CubeCoord{Q: -999999, R: -999999}] = true // Special marker for "rebuild all"
	r.scheduleRender()
}

// MarkAllUnitsDirty marks all units in the current map as dirty
func (r *LayeredRenderer) MarkAllUnitsDirty() {
	if r.currentMap == nil {
		return
	}

	// Clear existing dirty units
	for coord := range r.dirtyUnits {
		delete(r.dirtyUnits, coord)
	}

	// Mark all tiles with units as dirty
	for coord, tile := range r.currentMap.Tiles {
		if tile != nil && tile.Unit != nil {
			r.dirtyUnits[coord] = true
			fmt.Printf("MarkAllUnitsDirty: Marked unit at coord %v as dirty\n", coord)
		}
	}

	fmt.Printf("MarkAllUnitsDirty: Total %d units marked as dirty\n", len(r.dirtyUnits))
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
	fmt.Printf("LayeredRenderer.ForceRender called - terrain dirty: %d, units dirty: %d, UI dirty: %v\n",
		len(r.dirtyTerrain), len(r.dirtyUnits), r.dirtyUI)

	// Debug: List the dirty units
	if len(r.dirtyUnits) > 0 {
		fmt.Printf("Dirty units: ")
		for coord := range r.dirtyUnits {
			fmt.Printf("%v ", coord)
		}
		fmt.Printf("\n")
	}

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

	// Update terrain layer if dirty
	if len(r.dirtyTerrain) > 0 {
		fmt.Printf("Updating terrain layer with %d dirty tiles\n", len(r.dirtyTerrain))
		r.updateTerrainLayer()
	}

	// Update unit layer if dirty
	if len(r.dirtyUnits) > 0 {
		fmt.Printf("Updating unit layer with %d dirty positions\n", len(r.dirtyUnits))
		r.updateUnitLayer()
	}

	// Update UI layer if dirty
	if r.dirtyUI {
		fmt.Printf("Updating UI layer\n")
		r.updateUILayer()
		r.dirtyUI = false
	}

	// Composite all layers to main canvas
	fmt.Printf("Compositing layers to main canvas\n")
	r.composite()
	fmt.Printf("DEBUG: performRender() completed successfully\n")
}

// updateTerrainLayer renders dirty terrain tiles
func (r *LayeredRenderer) updateTerrainLayer() {
	// Check if full rebuild is needed
	_, fullRebuild := r.dirtyTerrain[CubeCoord{Q: -999999, R: -999999}]
	_, renderAllVisible := r.dirtyTerrain[CubeCoord{Q: -999998, R: -999998}]

	if fullRebuild || renderAllVisible {
		// Clear entire terrain buffer
		r.terrainBuffer.Clear()

		// Render all tiles in the current map
		if r.currentMap != nil {
			fmt.Printf("Rendering all %d tiles in map\n", len(r.currentMap.Tiles))
			for coord := range r.currentMap.Tiles {
				r.renderTerrainTile(coord)
			}
		}

		// Clear both markers
		delete(r.dirtyTerrain, CubeCoord{Q: -999999, R: -999999})
		delete(r.dirtyTerrain, CubeCoord{Q: -999998, R: -999998})
	} else {
		// Render individual dirty tiles
		for coord := range r.dirtyTerrain {
			r.renderTerrainTile(coord)
			delete(r.dirtyTerrain, coord)
		}
	}
}

// renderTerrainTile renders a single terrain tile using cached sprites
func (r *LayeredRenderer) renderTerrainTile(coord CubeCoord) {
	if r.currentMap == nil {
		return
	}

	// Get tile from current map
	tile := r.currentMap.TileAtCube(coord)

	// Calculate pixel position from hex coordinate
	x, y := r.hexToPixel(coord)

	if tile != nil {
		// Try to use real terrain sprite if asset provider is available
		if r.assetProvider != nil && r.assetProvider.HasTileAsset(tile.TileType) {
			// fmt.Printf("Rendering terrain sprite for tile type %d at (%f, %f)\n", tile.TileType, x, y)
			r.renderTerrainSprite(coord, tile.TileType, x, y)
		} else {
			// Fallback to colored hexagon using buffer operations
			if r.assetProvider == nil {
				fmt.Printf("AssetProvider is nil, using colored hex for tile type %d at (%f, %f)\n", tile.TileType, x, y)
			} else {
				fmt.Printf("AssetProvider.HasTileAsset(%d) returned false, using colored hex at (%f, %f)\n", tile.TileType, x, y)
			}
			color := r.getTerrainColor(tile.TileType)
			r.drawSimpleHexToBuffer(r.terrainBuffer, x, y, color)
		}
	}
}

// renderTerrainSprite renders a terrain sprite at the given position
func (r *LayeredRenderer) renderTerrainSprite(coord CubeCoord, tileType int, x, y float64) {
	// Check if we have a cached sprite for this tile type
	cachedSprite, exists := r.terrainSprites[tileType]
	if !exists {
		// Load and cache the terrain sprite
		img, err := r.assetProvider.GetTileImage(tileType)
		if err != nil {
			fmt.Printf("Failed to load terrain sprite for type %d: %v\n", tileType, err)
			// Fallback to colored hex
			color := r.getTerrainColor(tileType)
			r.drawSimpleHexToBuffer(r.terrainBuffer, x, y, color)
			return
		}

		// Debug: Check image properties
		bounds := img.Bounds()
		fmt.Printf("Loaded terrain sprite for type %d: size %dx%d, bounds %v\n", tileType, bounds.Dx(), bounds.Dy(), bounds)

		// Cache the image directly
		r.terrainSprites[tileType] = img
		cachedSprite = img
	}

	// Draw the sprite to the terrain buffer with proper alpha blending
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in drawImageToBuffer: %v\n", r)
		}
	}()
	r.drawImageToBuffer(r.terrainBuffer, cachedSprite, x, y, r.tileWidth, r.tileHeight)
}

// getTerrainColor returns the color for a terrain type
func (r *LayeredRenderer) getTerrainColor(terrainType int) string {
	switch terrainType {
	case 1: // Grass
		return "#228B22"
	case 2: // Desert
		return "#EECBAD"
	case 3: // Water
		return "#4169E1"
	case 4: // Mountain
		return "#8B8989"
	case 5: // Rock
		return "#696969"
	default:
		return "#C8C8C8"
	}
}

// updateUnitLayer renders dirty unit positions
func (r *LayeredRenderer) updateUnitLayer() {
	fmt.Printf("updateUnitLayer called with %d dirty units\n", len(r.dirtyUnits))
	// Clear dirty areas and redraw units using buffer operations
	for coord := range r.dirtyUnits {
		fmt.Printf("Processing dirty unit at coord %v\n", coord)
		// Clear the specific hex area in unitBuffer first
		r.clearHexArea(r.unitBuffer, coord)

		// Get unit at this position from current map
		if r.currentMap != nil {
			tile := r.currentMap.TileAtCube(coord)
			if tile != nil && tile.Unit != nil {
				// Render unit sprite to unitBuffer
				r.renderUnitSprite(coord, tile.Unit)
			}
		}

		delete(r.dirtyUnits, coord)
	}
}

// updateUILayer renders UI elements (selection, hover, etc.)
func (r *LayeredRenderer) updateUILayer() {
	// Clear entire UI buffer
	r.uiBuffer.Clear()

	// TODO: Render current selection highlight to uiBuffer
	// TODO: Render hover highlight to uiBuffer
	// TODO: Render range indicators to uiBuffer, etc.
}

// composite just marks that layers need to be blitted
func (r *LayeredRenderer) composite() {
	// No complex compositing - just signal that buffers are ready for blitting
}

// GetTerrainBuffer returns the terrain buffer for external blitting
func (r *LayeredRenderer) GetTerrainBuffer() *Buffer {
	return r.terrainBuffer
}

// GetUnitBuffer returns the unit buffer for external blitting
func (r *LayeredRenderer) GetUnitBuffer() *Buffer {
	return r.unitBuffer
}

// GetUIBuffer returns the UI buffer for external blitting
func (r *LayeredRenderer) GetUIBuffer() *Buffer {
	return r.uiBuffer
}

// blendBuffers blends src buffer onto dst buffer with alpha blending
func (r *LayeredRenderer) blendBuffers(dst, src *Buffer) {
	dstImg := dst.GetImageData()
	srcImg := src.GetImageData()

	// Use Go's image/draw for proper alpha blending
	draw.Draw(dstImg, dstImg.Bounds(), srcImg, image.Point{}, draw.Over)
}

// drawImageToBuffer draws an image to a buffer with proper alpha blending
func (r *LayeredRenderer) drawImageToBuffer(buffer *Buffer, img image.Image, x, y, width, height float64) {
	// Get the buffer's underlying image
	bufferImg := buffer.GetImageData()

	// Calculate destination rectangle (centered on x,y)
	destRect := image.Rect(
		int(x-width/2),
		int(y-height/2),
		int(x+width/2),
		int(y+height/2),
	)

	// fmt.Printf("DrawImageToBuffer: Drawing image at (%f,%f) size %fx%f, destRect %v, img bounds %v\n", x, y, width, height, destRect, img.Bounds())

	// Resize source image to match destination size if needed
	srcBounds := img.Bounds()
	if srcBounds.Dx() != int(width) || srcBounds.Dy() != int(height) {
		// fmt.Printf("Resizing image from %dx%d to %dx%d\n", srcBounds.Dx(), srcBounds.Dy(), int(width), int(height))
		// Create a resized version
		resized := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		// Simple nearest-neighbor scaling for now
		if int(width) > 0 && int(height) > 0 && srcBounds.Dx() > 0 && srcBounds.Dy() > 0 {
			for dy := 0; dy < int(height); dy++ {
				for dx := 0; dx < int(width); dx++ {
					srcX := dx * srcBounds.Dx() / int(width)
					srcY := dy * srcBounds.Dy() / int(height)
					resized.Set(dx, dy, img.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
				}
			}
		}
		img = resized
		// fmt.Printf("Image resizing completed\n")
	}

	// Draw the image using Go's image/draw with alpha blending
	draw.DrawMask(bufferImg, destRect, img, image.Point{}, nil, image.Point{}, draw.Over)
	// fmt.Printf("Image drawing completed successfully\n")
}

// drawSimpleHexToBuffer draws a simple colored hexagon to a buffer
func (r *LayeredRenderer) drawSimpleHexToBuffer(buffer *Buffer, x, y float64, colorStr string) {
	// Convert hex color string to Color struct
	hexColor := r.parseHexColor(colorStr)

	// Get the buffer's underlying image
	bufferImg := buffer.GetImageData()

	// Draw a simple filled ellipse as a placeholder for hexagon
	// Use tile dimensions to determine the shape
	radiusX := int(r.tileWidth / 2)
	radiusY := int(r.tileHeight / 2)
	centerX, centerY := int(x), int(y)

	fmt.Printf("Drawing simple hex at (%d, %d) with radii %dx%d, color %s (%d,%d,%d,%d)\n",
		centerX, centerY, radiusX, radiusY, colorStr, hexColor.R, hexColor.G, hexColor.B, hexColor.A)

	for dy := -radiusY; dy <= radiusY; dy++ {
		for dx := -radiusX; dx <= radiusX; dx++ {
			// Ellipse equation: (x/a)² + (y/b)² <= 1
			if float64(dx*dx)/float64(radiusX*radiusX) + float64(dy*dy)/float64(radiusY*radiusY) <= 1.0 {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && py >= 0 && px < r.width && py < r.height {
					// Convert our Color to color.RGBA
					rgba := color.RGBA{R: hexColor.R, G: hexColor.G, B: hexColor.B, A: hexColor.A}
					bufferImg.Set(px, py, rgba)
				}
			}
		}
	}
}

// parseHexColor converts a hex color string like "#228B22" to Color
func (r *LayeredRenderer) parseHexColor(hexColor string) Color {
	// Remove # if present
	if len(hexColor) > 0 && hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Default to green if parsing fails
	if len(hexColor) != 6 {
		return Color{R: 34, G: 139, B: 34, A: 255}
	}

	// Parse RGB components
	var red, green, blue uint8
	fmt.Sscanf(hexColor[0:2], "%02x", &red)
	fmt.Sscanf(hexColor[2:4], "%02x", &green)
	fmt.Sscanf(hexColor[4:6], "%02x", &blue)

	return Color{R: red, G: green, B: blue, A: 255}
}

// hexToPixel converts hex coordinates to pixel coordinates using the same logic as game.go
func (r *LayeredRenderer) hexToPixel(coord CubeCoord) (float64, float64) {
	// Use the same conversion as game.go XYForTile - convert to display coordinates first
	row := coord.R
	col := coord.Q + (coord.R+(coord.R&1))/2

	// Use the exact same calculation as game.go XYForTile
	x := float64(col)*r.tileWidth + r.tileWidth/2
	
	// Apply offset for alternating rows (hex grid staggering)
	isEvenRow := (row % 2) == 0
	// Assuming odd rows are offset (EvenRowsOffset() returns false)
	if !isEvenRow {
		x += r.tileWidth / 2
	}
	
	y := float64(row)*r.yIncrement + r.tileHeight/2

	return x, y
}

// clearHexArea clears a hexagonal area in the buffer at the given coordinate
func (r *LayeredRenderer) clearHexArea(buffer *Buffer, coord CubeCoord) {
	// Calculate pixel position
	x, y := r.hexToPixel(coord)

	// Get the buffer's underlying image
	bufferImg := buffer.GetImageData()

	// Clear an elliptical area (approximate hex area using tile dimensions)
	radiusX := int(r.tileWidth / 2)
	radiusY := int(r.tileHeight / 2)
	centerX, centerY := int(x), int(y)

	transparentColor := color.RGBA{R: 0, G: 0, B: 0, A: 0}

	for dy := -radiusY; dy <= radiusY; dy++ {
		for dx := -radiusX; dx <= radiusX; dx++ {
			// Ellipse equation: (x/a)² + (y/b)² <= 1
			if float64(dx*dx)/float64(radiusX*radiusX) + float64(dy*dy)/float64(radiusY*radiusY) <= 1.0 {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && py >= 0 && px < r.width && py < r.height {
					bufferImg.Set(px, py, transparentColor)
				}
			}
		}
	}
}

// renderUnitSprite renders a unit sprite at the given coordinate
func (r *LayeredRenderer) renderUnitSprite(coord CubeCoord, unit *Unit) {
	// Calculate pixel position
	x, y := r.hexToPixel(coord)

	// Try to use real unit sprite if asset provider is available
	if r.assetProvider != nil && r.assetProvider.HasUnitAsset(unit.UnitType, unit.PlayerID) {
		// Check if we have a cached sprite for this unit type and player
		spriteKey := fmt.Sprintf("%d_%d", unit.UnitType, unit.PlayerID)
		cachedSprite, exists := r.unitSprites[spriteKey]
		if !exists {
			// Load and cache the unit sprite
			img, err := r.assetProvider.GetUnitImage(unit.UnitType, unit.PlayerID)
			if err != nil {
				fmt.Printf("Failed to load unit sprite for type %d, player %d: %v\n", unit.UnitType, unit.PlayerID, err)
				// Fallback to simple colored circle
				r.drawSimpleUnitToBuffer(r.unitBuffer, x, y, unit.PlayerID)
				return
			}

			// Cache the image
			r.unitSprites[spriteKey] = img
			cachedSprite = img
		}

		// Draw the sprite to the unit buffer
		fmt.Printf("Drawing unit sprite at position (%f, %f) with tileDimensions %fx%f, sprite bounds: %v\n",
			x, y, r.tileWidth, r.tileHeight, cachedSprite.Bounds())
		r.drawImageToBuffer(r.unitBuffer, cachedSprite, x, y, r.tileWidth, r.tileHeight)
	} else {
		// Fallback to simple colored circle
		fmt.Printf("Asset provider doesn't have unit asset, falling back to simple circle\n")
		r.drawSimpleUnitToBuffer(r.unitBuffer, x, y, unit.PlayerID)
	}
}

// drawSimpleUnitToBuffer draws a simple colored ellipse to represent a unit
func (r *LayeredRenderer) drawSimpleUnitToBuffer(buffer *Buffer, x, y float64, playerID int) {
	// Get player color
	var unitColor Color
	switch playerID {
	case 0:
		unitColor = Color{R: 255, G: 0, B: 0, A: 255} // Red
	case 1:
		unitColor = Color{R: 0, G: 0, B: 255, A: 255} // Blue
	default:
		unitColor = Color{R: 128, G: 128, B: 128, A: 255} // Gray
	}

	// Get the buffer's underlying image
	bufferImg := buffer.GetImageData()

	// Draw a smaller ellipse for units (60% of tile dimensions)
	radiusX := int(r.tileWidth * 0.3)  // 60% of half-width = 30% of full width
	radiusY := int(r.tileHeight * 0.3) // 60% of half-height = 30% of full height
	centerX, centerY := int(x), int(y)

	fmt.Printf("Drawing simple unit ellipse at (%d, %d) with radii %dx%d, player %d color (%d,%d,%d)\n",
		centerX, centerY, radiusX, radiusY, playerID, unitColor.R, unitColor.G, unitColor.B)

	for dy := -radiusY; dy <= radiusY; dy++ {
		for dx := -radiusX; dx <= radiusX; dx++ {
			// Ellipse equation: (x/a)² + (y/b)² <= 1
			if float64(dx*dx)/float64(radiusX*radiusX) + float64(dy*dy)/float64(radiusY*radiusY) <= 1.0 {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && py >= 0 && px < r.width && py < r.height {
					rgba := color.RGBA{R: unitColor.R, G: unitColor.G, B: unitColor.B, A: unitColor.A}
					bufferImg.Set(px, py, rgba)
				}
			}
		}
	}
}

// Resize updates the layer buffer sizes
func (r *LayeredRenderer) Resize(width, height int) error {
	r.width = width
	r.height = height

	// Recreate all layer buffers with new size
	r.terrainBuffer = NewBuffer(width, height)
	r.unitBuffer = NewBuffer(width, height)
	r.uiBuffer = NewBuffer(width, height)

	// Clear all buffers to transparent
	r.terrainBuffer.Clear()
	r.unitBuffer.Clear()
	r.uiBuffer.Clear()

	// Mark everything as dirty for redraw
	r.MarkAllTerrainDirty()
	r.MarkUIDirty()

	return nil
}
