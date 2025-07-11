package weewar

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	
	xdraw "golang.org/x/image/draw"
)

// Buffer represents a drawable canvas with compositing capabilities
type Buffer struct {
	img    *image.RGBA
	width  int
	height int
}

// NewBuffer creates a new buffer with the specified dimensions
func NewBuffer(width, height int) *Buffer {
	return &Buffer{
		img:    image.NewRGBA(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}
}

// Clear clears the buffer (fills with transparent pixels)
func (b *Buffer) Clear() {
	draw.Draw(b.img, b.img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)
}

// Copy creates a deep copy of the buffer
func (b *Buffer) Copy() *Buffer {
	newBuffer := NewBuffer(b.width, b.height)
	draw.Draw(newBuffer.img, newBuffer.img.Bounds(), b.img, image.Point{}, draw.Src)
	return newBuffer
}

// Size returns the dimensions of the buffer
func (b *Buffer) Size() (width, height float64) {
	return float64(b.width), float64(b.height)
}

// DrawImage draws an image at the specified position with scaling and alpha compositing
func (b *Buffer) DrawImage(x, y, width, height float64, img image.Image) {
	// Calculate destination rectangle
	dstRect := image.Rect(int(x), int(y), int(x+width), int(y+height))
	
	// Clip to buffer bounds
	dstRect = dstRect.Intersect(b.img.Bounds())
	if dstRect.Empty() {
		return // Nothing to draw
	}
	
	// Get source bounds
	srcBounds := img.Bounds()
	
	// Use bilinear scaling for smooth results
	xdraw.BiLinear.Scale(b.img, dstRect, img, srcBounds, draw.Over, nil)
}

// RenderBuffer copies a source buffer onto this buffer with proper clipping
func (dest *Buffer) RenderBuffer(src *Buffer) {
	// Calculate intersection of source and destination
	srcBounds := src.img.Bounds()
	dstBounds := dest.img.Bounds()
	
	// Copy the source buffer onto the destination with alpha blending
	draw.Draw(dest.img, dstBounds, src.img, srcBounds.Min, draw.Over)
}

// Save saves the buffer to a PNG file
func (b *Buffer) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return png.Encode(file, b.img)
}