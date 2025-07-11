package weewar

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png" // For PNG decoding
	"os"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	xdraw "golang.org/x/image/draw"
)

// Buffer represents a drawable canvas with compositing capabilities
type Buffer struct {
	img    *image.RGBA
	width  int
	height int
}

// Point represents a 2D point
type Point struct {
	X, Y float64
}

// Color represents a color with alpha channel
type Color struct {
	R, G, B, A uint8
}

// StrokeProperties defines stroke rendering properties
type StrokeProperties struct {
	Width       float64
	LineCap     string // "butt", "round", "square"
	LineJoin    string // "miter", "round", "bevel"
	DashPattern []float64
	DashOffset  float64
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

// FillPath fills a given path with the given color (with alpha channel compositing)
func (b *Buffer) FillPath(points []Point, fillColor Color) {
	if len(points) < 2 {
		return // Need at least 2 points to create a path
	}

	// Create canvas with buffer dimensions (in mm, so scale down)
	c := canvas.New(float64(b.width)/3.78, float64(b.height)/3.78) // ~96 DPI conversion
	ctx := canvas.NewContext(c)

	// Set fill color
	rgba := color.RGBA{R: fillColor.R, G: fillColor.G, B: fillColor.B, A: fillColor.A}
	ctx.SetFillColor(rgba)

	// Build path (scale coordinates to mm)
	ctx.MoveTo(points[0].X/3.78, points[0].Y/3.78)
	for i := 1; i < len(points); i++ {
		ctx.LineTo(points[i].X/3.78, points[i].Y/3.78)
	}
	ctx.Close()

	// Fill the path
	ctx.Fill()

	// Render canvas to a temporary file and then load it
	tempFile := "/tmp/temp_fill.png"
	err := renderers.Write(tempFile, c, canvas.DPMM(3.78))
	if err != nil {
		return // Skip if rendering fails
	}

	// Load the temporary image
	file, err := os.Open(tempFile)
	if err != nil {
		return
	}
	defer file.Close()
	defer os.Remove(tempFile)

	tempImg, _, err := image.Decode(file)
	if err != nil {
		return
	}

	// Composite the temporary image onto the buffer
	draw.Draw(b.img, b.img.Bounds(), tempImg, image.Point{}, draw.Over)
}

// StrokePath strokes a given path with a given color and stroke properties
func (b *Buffer) StrokePath(points []Point, strokeColor Color, strokeProperties StrokeProperties) {
	if len(points) < 2 {
		return // Need at least 2 points to create a path
	}

	// Create canvas with buffer dimensions (in mm, so scale down)
	c := canvas.New(float64(b.width)/3.78, float64(b.height)/3.78) // ~96 DPI conversion
	ctx := canvas.NewContext(c)

	// Set stroke color
	rgba := color.RGBA{R: strokeColor.R, G: strokeColor.G, B: strokeColor.B, A: strokeColor.A}
	ctx.SetStrokeColor(rgba)

	// Set stroke width (scale to mm)
	ctx.SetStrokeWidth(strokeProperties.Width / 3.78)

	// Set line cap
	switch strokeProperties.LineCap {
	case "round":
		ctx.SetStrokeCapper(canvas.RoundCapper{})
	case "square":
		ctx.SetStrokeCapper(canvas.SquareCapper{})
	default: // "butt" or unspecified
		ctx.SetStrokeCapper(canvas.ButtCapper{})
	}

	// Set line join
	switch strokeProperties.LineJoin {
	case "round":
		ctx.SetStrokeJoiner(canvas.RoundJoiner{})
	case "bevel":
		ctx.SetStrokeJoiner(canvas.BevelJoiner{})
	default: // "miter" or unspecified
		ctx.SetStrokeJoiner(canvas.MiterJoiner{})
	}

	// Set dash pattern if specified (scale to mm)
	if len(strokeProperties.DashPattern) > 0 {
		scaledDashes := make([]float64, len(strokeProperties.DashPattern))
		for i, dash := range strokeProperties.DashPattern {
			scaledDashes[i] = dash / 3.78
		}
		ctx.SetDashes(strokeProperties.DashOffset/3.78, scaledDashes...)
	}

	// Build path (scale coordinates to mm)
	ctx.MoveTo(points[0].X/3.78, points[0].Y/3.78)
	for i := 1; i < len(points); i++ {
		ctx.LineTo(points[i].X/3.78, points[i].Y/3.78)
	}

	// Stroke the path
	ctx.Stroke()

	// Render canvas to a temporary file and then load it
	tempFile := "/tmp/temp_stroke.png"
	err := renderers.Write(tempFile, c, canvas.DPMM(3.78))
	if err != nil {
		return // Skip if rendering fails
	}

	// Load the temporary image
	file, err := os.Open(tempFile)
	if err != nil {
		return
	}
	defer file.Close()
	defer os.Remove(tempFile)

	tempImg, _, err := image.Decode(file)
	if err != nil {
		return
	}

	// Composite the temporary image onto the buffer
	draw.Draw(b.img, b.img.Bounds(), tempImg, image.Point{}, draw.Over)
}

// DrawText renders text at the specified position with the given font size and color
func (b *Buffer) DrawText(x, y float64, text string, fontSize float64, textColor Color) {
	b.DrawTextWithStyle(x, y, text, fontSize, textColor, false, Color{})
}

// DrawTextWithStyle renders text with optional bold and background
func (b *Buffer) DrawTextWithStyle(x, y float64, text string, fontSize float64, textColor Color, bold bool, backgroundColor Color) {
	if text == "" {
		return
	}

	// Create canvas with buffer dimensions (in mm, so scale down)
	c := canvas.New(float64(b.width)/3.78, float64(b.height)/3.78) // ~96 DPI conversion
	ctx := canvas.NewContext(c)

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

	// 3.78 = 96 DPI ÷ 25.4 mm/inch
	face := fontFamily.Face(fontSize/3.78, rgba, fontWeight, canvas.FontNormal) // Scale font size to mm

	// Calculate text metrics for background
	textLine := canvas.NewTextLine(face, text, canvas.Left)
	textWidth := textLine.Bounds().W()
	textHeight := textLine.Bounds().H()

	// Draw background rectangle if specified
	if backgroundColor.A > 0 {
		canvasHeight := float64(b.height) / 3.78
		canvasX := x / 3.78
		canvasY := canvasHeight - (y / 3.78) // Flip Y coordinate

		// Add padding around text
		padding := 2.0
		bgX := canvasX - padding
		// bgY := (canvasY - padding) //  - (textHeight * 2)
		bgY := canvasY + 1
		bgWidth := textWidth + (padding * 2)
		bgHeight := (textHeight + (padding * 2)) / 2

		ctx.SetFillColor(color.RGBA{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B, A: backgroundColor.A})
		ctx.DrawPath(bgX, bgY, canvas.Rectangle(bgWidth, bgHeight))
		ctx.Fill()
	}

	// Draw the text (coordinates already calculated above)
	canvasHeight := float64(b.height) / 3.78
	canvasX := x / 3.78
	canvasY := canvasHeight - (y / 3.78) // Flip Y coordinate
	ctx.DrawText(canvasX, canvasY, textLine)

	// Render canvas to a temporary file and then load it
	tempFile := "/tmp/temp_text.png"
	err := renderers.Write(tempFile, c, canvas.DPMM(3.78))
	if err != nil {
		return // Skip if rendering fails
	}

	// Load the temporary image
	file, err := os.Open(tempFile)
	if err != nil {
		return
	}
	defer file.Close()
	defer os.Remove(tempFile)

	tempImg, _, err := image.Decode(file)
	if err != nil {
		return
	}

	// Composite the temporary image onto the buffer
	draw.Draw(b.img, b.img.Bounds(), tempImg, image.Point{}, draw.Over)
}
