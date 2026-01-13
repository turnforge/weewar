package themes

import (
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// WorldRenderer renders a world (tiles and units) to a specific output format
type WorldRenderer interface {
	// Render produces a composite image of the world
	// Returns the image bytes and the content type (e.g., "image/png", "image/svg+xml")
	Render(tiles map[string]*v1.Tile, units map[string]*v1.Unit, options *lib.RenderOptions) ([]byte, string, error)
}

// CreateWorldRenderer returns the appropriate renderer for a theme
func CreateWorldRenderer(theme Theme) (WorldRenderer, error) {
	info := theme.GetThemeInfo()
	if info == nil {
		return nil, fmt.Errorf("theme has no theme info")
	}

	switch info.AssetType {
	case "png":
		return NewPNGWorldRenderer(theme)
	case "svg":
		return NewSVGWorldRenderer(theme)
	default:
		return nil, fmt.Errorf("unknown asset type: %s", info.AssetType)
	}
}

// computeBounds calculates the bounding box for tiles and units
// Returns minX, minY, width, height in pixel coordinates
func computeBounds(tiles map[string]*v1.Tile, units map[string]*v1.Unit, opts *lib.RenderOptions) (minX, minY, width, height int) {
	bounds := lib.ComputeWorldBounds(tiles, units, opts)
	return bounds.MinX, bounds.MinY, bounds.Width, bounds.Height
}
