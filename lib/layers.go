package weewar

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// Layer represents a single rendering layer (terrain, units, UI, etc.)
type Layer interface {
	// Core rendering
	Render(world *World, options LayerRenderOptions)

	// Dirty tracking for efficient updates
	MarkDirty(coord CubeCoord)
	MarkAllDirty()
	ClearDirty()
	IsDirty() bool

	// Lifecycle management
	SetViewPort(x, y, width, height int)
	SetAssetProvider(provider AssetProvider)

	// Layer identification
	GetName() string
}

// LayerRenderOptions contains rendering parameters for layers
type LayerRenderOptions struct {
	// Hex grid parameters
	TileWidth  float64
	TileHeight float64
	YIncrement float64

	// Visual options
	ShowGrid        bool
	ShowCoordinates bool
}

// LayerScheduler interface for layers to request renders
type LayerScheduler interface {
	ScheduleRender()
}

// BaseLayer provides common functionality for all layers
type BaseLayer struct {
	name                string
	x, y, width, height int
	buffer              *Buffer

	// Dirty tracking
	dirtyCoords map[CubeCoord]bool
	allDirty    bool

	// Asset provider
	assetProvider AssetProvider

	// Renderer reference for scheduling
	scheduler LayerScheduler
}

// NewBaseLayer creates a new base layer
func NewBaseLayer(name string, width, height int, scheduler LayerScheduler) *BaseLayer {
	return &BaseLayer{
		name:        name,
		width:       width,
		height:      height,
		buffer:      NewBuffer(int(width), int(height)),
		dirtyCoords: make(map[CubeCoord]bool),
		allDirty:    true, // Start with everything dirty
		scheduler:   scheduler,
	}
}

// Common BaseLayer methods
func (bl *BaseLayer) GetName() string {
	return bl.name
}

func (bl *BaseLayer) MarkDirty(coord CubeCoord) {
	bl.dirtyCoords[coord] = true
	if true || bl.scheduler != nil {
		bl.scheduler.ScheduleRender()
	}
}

func (bl *BaseLayer) MarkAllDirty() {
	bl.allDirty = true
	bl.dirtyCoords = make(map[CubeCoord]bool)
	if true || bl.scheduler != nil {
		bl.scheduler.ScheduleRender()
	}
}

func (bl *BaseLayer) ClearDirty() {
	bl.dirtyCoords = make(map[CubeCoord]bool)
	bl.allDirty = false
}

func (bl *BaseLayer) IsDirty() bool {
	return bl.allDirty || len(bl.dirtyCoords) > 0
}

func (bl *BaseLayer) SetViewPort(x, y, width, height int) {
	fmt.Printf("BaseLayer.SetViewPort called on layer '%s' with: x=%d, y=%d, width=%d, height=%d\n", bl.name, x, y, width, height)
	bl.x = x
	bl.y = y
	bl.width = width
	bl.height = height
	bl.buffer = NewBuffer(width, height)
	bl.MarkAllDirty()
}

func (bl *BaseLayer) SetAssetProvider(provider AssetProvider) {
	bl.assetProvider = provider
	bl.MarkAllDirty()
}

func (bl *BaseLayer) GetBuffer() *Buffer {
	return bl.buffer
}

// drawImageToBuffer draws an image to the buffer
func (ul *BaseLayer) drawImageToBuffer(img image.Image, x, y, width, height float64) {
	bufferImg := ul.buffer.GetImageData()

	destRect := image.Rect(
		int(x-width/2), int(y-height/2),
		int(x+width/2), int(y+height/2),
	)

	draw.DrawMask(bufferImg, destRect, img, image.Point{}, nil, image.Point{}, draw.Over)
}

// drawSimpleHexToBuffer draws a colored hexagon
func (tl *BaseLayer) drawSimpleHexToBuffer(x, y float64, hexColor Color, options LayerRenderOptions) {
	bufferImg := tl.buffer.GetImageData()

	// Draw ellipse approximation
	radiusX := int(options.TileWidth / 2)
	radiusY := int(options.TileHeight / 2)
	centerX, centerY := int(x), int(y)
	width, height := int(tl.width), int(tl.height)

	for dy := -radiusY; dy <= radiusY; dy++ {
		for dx := -radiusX; dx <= radiusX; dx++ {
			if float64(dx*dx)/float64(radiusX*radiusX)+float64(dy*dy)/float64(radiusY*radiusY) <= 1.0 {
				px, py := centerX+dx, centerY+dy
				if px >= 0 && py >= 0 && px < width && py < height {
					rgba := color.RGBA{R: hexColor.R, G: hexColor.G, B: hexColor.B, A: hexColor.A}
					bufferImg.Set(px, py, rgba)
				}
			}
		}
	}
}

// parseHexColor converts hex color string to Color
func (tl *BaseLayer) parseHexColor(hexColor string) Color {
	if len(hexColor) > 0 && hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	if len(hexColor) != 6 {
		return Color{R: 34, G: 139, B: 34, A: 255}
	}

	var red, green, blue uint8
	fmt.Sscanf(hexColor[0:2], "%02x", &red)
	fmt.Sscanf(hexColor[2:4], "%02x", &green)
	fmt.Sscanf(hexColor[4:6], "%02x", &blue)

	return Color{R: red, G: green, B: blue, A: 255}
}
