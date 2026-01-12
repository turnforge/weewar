package themes

import (
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/web/assets"
)

// FantasyTheme implements the Theme interface for the Medieval Fantasy theme
// Mirrors fantasy.ts and extends BaseTheme
type FantasyTheme struct {
	*BaseTheme
}

// NewFantasyTheme creates a new Fantasy theme instance
// Loads the mapping.json from embedded files
func NewFantasyTheme(cityTerrains map[int32]bool) (*FantasyTheme, error) {
	manifest, err := assets.LoadThemeManifest("fantasy")
	if err != nil {
		return nil, fmt.Errorf("failed to load fantasy theme: %w", err)
	}

	// Ensure themeInfo is populated (if not in mapping.json, set defaults)
	if manifest.ThemeInfo == nil {
		manifest.ThemeInfo = &v1.ThemeInfo{
			Name:                "Medieval Fantasy",
			Version:             "1.0.0",
			BasePath:            "/static/assets/themes/fantasy",
			AssetType:           "svg",
			NeedsPostProcessing: true,
		}
	}

	return &FantasyTheme{
		BaseTheme: NewBaseTheme(manifest, cityTerrains),
	}, nil
}

// GetUnitAssetPath returns the full path to a unit SVG template
// For SVG themes, this returns the path to the template file
func (f *FantasyTheme) GetUnitAssetPath(unitId int32) string {
	if path := f.GetUnitPath(unitId); path != "" {
		return fmt.Sprintf("%s/%s", f.manifest.ThemeInfo.BasePath, path)
	}
	return ""
}

// GetTileAssetPath returns the full path to a terrain SVG template
func (f *FantasyTheme) GetTileAssetPath(terrainId int32) string {
	if path := f.GetTilePath(terrainId); path != "" {
		return fmt.Sprintf("%s/%s", f.manifest.ThemeInfo.BasePath, path)
	}
	return ""
}
