package themes

import (
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/web/assets"
)

// ModernTheme implements the Theme interface for the Modern Military theme
// Mirrors modern.ts and extends BaseTheme
type ModernTheme struct {
	*BaseTheme
}

// NewModernTheme creates a new Modern theme instance
// Loads the mapping.json from embedded files
func NewModernTheme(cityTerrains map[int32]bool) (*ModernTheme, error) {
	manifest, err := assets.LoadThemeManifest("modern")
	if err != nil {
		return nil, fmt.Errorf("failed to load modern theme: %w", err)
	}

	// Ensure themeInfo is populated (if not in mapping.json, set defaults)
	if manifest.ThemeInfo == nil {
		manifest.ThemeInfo = &v1.ThemeInfo{
			Name:                "Modern Military",
			Version:             "1.0.0",
			BasePath:            "/static/assets/themes/modern",
			AssetType:           "svg",
			NeedsPostProcessing: true,
		}
	}

	return &ModernTheme{
		BaseTheme: NewBaseTheme(manifest, cityTerrains),
	}, nil
}

// GetUnitAssetPath returns the full path to a unit SVG template
// For SVG themes, this returns the path to the template file
func (m *ModernTheme) GetUnitAssetPath(unitId int32) string {
	if path := m.GetUnitPath(unitId); path != "" {
		return fmt.Sprintf("%s/%s", m.manifest.ThemeInfo.BasePath, path)
	}
	return ""
}

// GetTileAssetPath returns the full path to a terrain SVG template
func (m *ModernTheme) GetTileAssetPath(terrainId int32) string {
	if path := m.GetTilePath(terrainId); path != "" {
		return fmt.Sprintf("%s/%s", m.manifest.ThemeInfo.BasePath, path)
	}
	return ""
}
