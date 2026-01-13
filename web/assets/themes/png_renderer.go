package themes

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// PNGWorldRenderer renders worlds using PNG assets (for default theme)
type PNGWorldRenderer struct {
	theme *DefaultTheme

	// Cache for loaded images
	tileCache  map[string]image.Image
	unitCache  map[string]image.Image
	cacheMutex sync.RWMutex
}

// NewPNGWorldRenderer creates a new PNG-based world renderer
func NewPNGWorldRenderer(theme Theme) (*PNGWorldRenderer, error) {
	defaultTheme, ok := theme.(*DefaultTheme)
	if !ok {
		return nil, fmt.Errorf("PNGWorldRenderer requires a DefaultTheme")
	}

	return &PNGWorldRenderer{
		theme:     defaultTheme,
		tileCache: make(map[string]image.Image),
		unitCache: make(map[string]image.Image),
	}, nil
}

// Render produces a composite PNG image of the world
func (r *PNGWorldRenderer) Render(tiles map[string]*v1.Tile, units map[string]*v1.Unit, options *lib.RenderOptions) ([]byte, string, error) {
	if options == nil {
		options = lib.DefaultRenderOptions()
	}

	if len(tiles) == 0 {
		return nil, "", fmt.Errorf("no tiles to render")
	}

	// Compute bounds
	minX, minY, width, height := computeBounds(tiles, units, options)

	// Create the output image
	outputImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Render tiles first (background layer)
	for _, tile := range tiles {
		if err := r.renderTile(outputImg, tile, minX, minY, options); err != nil {
			// Log but continue - don't fail entire render for one missing tile
			fmt.Printf("Warning: failed to render tile at (%d,%d): %v\n", tile.Q, tile.R, err)
		}
	}

	// Render units on top
	for _, unit := range units {
		if err := r.renderUnit(outputImg, unit, minX, minY, options); err != nil {
			fmt.Printf("Warning: failed to render unit at (%d,%d): %v\n", unit.Q, unit.R, err)
		}
	}

	// Render tile labels if enabled (below tiles, above units)
	if options.ShowTileLabels {
		for _, tile := range tiles {
			r.renderTileLabel(outputImg, tile, minX, minY, options)
		}
	}

	// Render unit labels if enabled (on top of everything)
	if options.ShowUnitLabels {
		for _, unit := range units {
			r.renderUnitLabel(outputImg, unit, minX, minY, options)
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, outputImg); err != nil {
		return nil, "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), "image/png", nil
}

// renderTile draws a single tile onto the output image
func (r *PNGWorldRenderer) renderTile(output *image.RGBA, tile *v1.Tile, offsetX, offsetY int, options *lib.RenderOptions) error {
	// Get tile image
	tileImg, err := r.getTileImage(tile.TileType, tile.Player)
	if err != nil {
		return err
	}

	// Calculate position (adjusted for offset)
	x, y := lib.HexToPixelInt32(tile.Q, tile.R, options)
	x -= offsetX
	y -= offsetY

	// Draw tile at position (x,y is top-left)
	r.drawImageAt(output, tileImg, x, y, options.TileWidth, options.TileHeight)
	return nil
}

// renderUnit draws a single unit onto the output image
func (r *PNGWorldRenderer) renderUnit(output *image.RGBA, unit *v1.Unit, offsetX, offsetY int, options *lib.RenderOptions) error {
	// Get unit image
	unitImg, err := r.getUnitImage(unit.UnitType, unit.Player)
	if err != nil {
		return err
	}

	// Calculate position (adjusted for offset)
	x, y := lib.HexToPixelInt32(unit.Q, unit.R, options)
	x -= offsetX
	y -= offsetY

	// Draw unit slightly smaller than tile (90% size) and centered within the tile
	unitWidth := int(float64(options.TileWidth) * 0.9)
	unitHeight := int(float64(options.TileHeight) * 0.9)
	unitX := x + (options.TileWidth-unitWidth)/2
	unitY := y + (options.TileHeight-unitHeight)/2
	r.drawImageAt(output, unitImg, unitX, unitY, unitWidth, unitHeight)
	return nil
}

// drawImageAt draws an image at the given top-left position with scaling and alpha blending
func (r *PNGWorldRenderer) drawImageAt(output *image.RGBA, src image.Image, x, y, width, height int) {
	srcBounds := src.Bounds()

	// Calculate destination rectangle (x,y is top-left corner)
	dstRect := image.Rect(x, y, x+width, y+height)

	// Use draw.Draw with draw.Over for proper alpha blending
	if srcBounds.Dx() == width && srcBounds.Dy() == height {
		draw.Draw(output, dstRect, src, srcBounds.Min, draw.Over)
	} else {
		// Scale the image first, then draw with alpha blending
		// Create a temporary scaled image
		scaled := image.NewRGBA(image.Rect(0, 0, width, height))

		scaleX := float64(srcBounds.Dx()) / float64(width)
		scaleY := float64(srcBounds.Dy()) / float64(height)

		for dy := range height {
			for dx := range width {
				srcX := int(float64(dx) * scaleX)
				srcY := int(float64(dy) * scaleY)
				if srcX < srcBounds.Dx() && srcY < srcBounds.Dy() {
					c := src.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY)
					scaled.Set(dx, dy, c)
				}
			}
		}

		// Now draw the scaled image with proper alpha blending
		draw.Draw(output, dstRect, scaled, image.Point{}, draw.Over)
	}
}

// getTileImage loads and caches a tile image
func (r *PNGWorldRenderer) getTileImage(tileType, playerId int32) (image.Image, error) {
	// Get path from theme - theme handles player color logic internally
	// Theme returns "/static/assets/themes/default/Tiles/1/0.png"
	webPath := r.theme.GetTileAssetPath(tileType, playerId)
	if webPath == "" {
		return nil, fmt.Errorf("tile %d not found in theme", tileType)
	}

	// Use web path as cache key (already includes effective player)
	cacheKey := webPath

	// Check cache first
	r.cacheMutex.RLock()
	if img, ok := r.tileCache[cacheKey]; ok {
		r.cacheMutex.RUnlock()
		return img, nil
	}
	r.cacheMutex.RUnlock()

	// Convert web path to filesystem path (remove leading "/" and prepend "web")
	path := "web" + webPath

	img, err := r.loadPNG(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load tile %d for player %d from %s: %w", tileType, playerId, path, err)
	}

	// Cache it
	r.cacheMutex.Lock()
	r.tileCache[cacheKey] = img
	r.cacheMutex.Unlock()

	return img, nil
}

// getUnitImage loads and caches a unit image
func (r *PNGWorldRenderer) getUnitImage(unitType, playerId int32) (image.Image, error) {
	cacheKey := fmt.Sprintf("unit_%d_%d", unitType, playerId)

	// Check cache first
	r.cacheMutex.RLock()
	if img, ok := r.unitCache[cacheKey]; ok {
		r.cacheMutex.RUnlock()
		return img, nil
	}
	r.cacheMutex.RUnlock()

	// Get path from theme and convert web path to filesystem path
	// Theme returns "/static/assets/themes/default/Units/1/0.png"
	// We need "web/static/assets/themes/default/Units/1/0.png"
	webPath := r.theme.GetUnitAssetPath(unitType, playerId)
	if webPath == "" {
		return nil, fmt.Errorf("unit %d not found in theme", unitType)
	}
	// Convert web path to filesystem path (remove leading "/" and prepend "web")
	path := "web" + webPath

	img, err := r.loadPNG(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load unit %d for player %d from %s: %w", unitType, playerId, path, err)
	}

	// Cache it
	r.cacheMutex.Lock()
	r.unitCache[cacheKey] = img
	r.cacheMutex.Unlock()

	return img, nil
}

// loadPNG loads a PNG file from the filesystem
func (r *PNGWorldRenderer) loadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	return img, nil
}

// ClearCache clears the image cache
func (r *PNGWorldRenderer) ClearCache() {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	r.tileCache = make(map[string]image.Image)
	r.unitCache = make(map[string]image.Image)
}

// renderUnitLabel draws a label below the unit showing "Shortcut:MP/Health"
func (r *PNGWorldRenderer) renderUnitLabel(output *image.RGBA, unit *v1.Unit, offsetX, offsetY int, options *lib.RenderOptions) {
	// Calculate position (adjusted for offset)
	x, y := lib.HexToPixelInt32(unit.Q, unit.R, options)
	x -= offsetX
	y -= offsetY

	// Format label: "Shortcut:MP/Health" or "MP/Health" if no shortcut
	health := unit.AvailableHealth
	if health == 0 {
		health = 10
	}
	mp := int(unit.DistanceLeft)

	var labelText string
	if unit.Shortcut != "" {
		labelText = fmt.Sprintf("%s:%d/%d", unit.Shortcut, mp, health)
	} else {
		labelText = fmt.Sprintf("%d/%d", mp, health)
	}

	// Use basicfont (7x13 pixels per character)
	face := basicfont.Face7x13

	// Calculate text dimensions
	textWidth := len(labelText) * 7
	textHeight := 13

	// Position: centered horizontally, near bottom of tile
	labelX := x + (options.TileWidth-textWidth)/2
	labelY := y + options.TileHeight - 4 // 4 pixels from bottom

	// Background color: brown with alpha (0x3d2817 from web)
	bgColor := color.RGBA{R: 0x3d, G: 0x28, B: 0x17, A: 0xB3} // ~70% opacity
	padding := 2

	// Draw background rectangle
	bgRect := image.Rect(
		labelX-padding,
		labelY-textHeight-padding,
		labelX+textWidth+padding,
		labelY+padding,
	)
	draw.Draw(output, bgRect, &image.Uniform{bgColor}, image.Point{}, draw.Over)

	// Draw text in white
	r.drawText(output, labelText, labelX, labelY, color.White, face)
}

// renderTileLabel draws a label at the top of the tile showing tile shortcut
func (r *PNGWorldRenderer) renderTileLabel(output *image.RGBA, tile *v1.Tile, offsetX, offsetY int, options *lib.RenderOptions) {
	// Only render if tile has a shortcut
	if tile.Shortcut == "" {
		return
	}

	// Calculate position (adjusted for offset)
	x, y := lib.HexToPixelInt32(tile.Q, tile.R, options)
	x -= offsetX
	y -= offsetY

	labelText := tile.Shortcut

	// Use basicfont (7x13 pixels per character)
	face := basicfont.Face7x13

	// Calculate text dimensions
	textWidth := len(labelText) * 7
	textHeight := 13

	// Position: centered horizontally, near top of tile
	labelX := x + (options.TileWidth-textWidth)/2
	labelY := y + textHeight + 4 // 4 pixels from top

	// Background color: dark teal with alpha (different from unit label brown)
	bgColor := color.RGBA{R: 0x17, G: 0x3d, B: 0x3d, A: 0xB3} // ~70% opacity
	padding := 2

	// Draw background rectangle
	bgRect := image.Rect(
		labelX-padding,
		labelY-textHeight-padding,
		labelX+textWidth+padding,
		labelY+padding,
	)
	draw.Draw(output, bgRect, &image.Uniform{bgColor}, image.Point{}, draw.Over)

	// Draw text in white
	r.drawText(output, labelText, labelX, labelY, color.White, face)
}

// drawText draws text at the given position
func (r *PNGWorldRenderer) drawText(img *image.RGBA, text string, x, y int, col color.Color, face font.Face) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}
