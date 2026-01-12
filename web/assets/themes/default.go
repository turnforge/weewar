package themes

import (
	_ "embed"
	"fmt"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed default/mapping.json
var defaultMappingJSON []byte

// DefaultTheme implements the Theme interface for PNG-based default assets
// Extends BaseTheme with PNG-specific asset path methods
type DefaultTheme struct {
	*BaseTheme
}

// NewDefaultTheme creates a new default theme instance by parsing embedded mapping.json
// cityTerrains is a map of terrain IDs that use player colors (from RulesEngine.TerrainTypes)
func NewDefaultTheme(cityTerrains map[int32]bool) *DefaultTheme {
	manifest := &v1.ThemeManifest{}
	if err := protojson.Unmarshal(defaultMappingJSON, manifest); err != nil {
		panic(fmt.Sprintf("failed to parse embedded default theme mapping: %v", err))
	}

	return &DefaultTheme{
		BaseTheme: NewBaseTheme(manifest, cityTerrains),
	}
}

// GetUnitAssetPath returns the full path to a specific unit+player PNG file
func (d *DefaultTheme) GetUnitAssetPath(unitId, playerId int32) string {
	if entry, ok := d.Manifest().Units[unitId]; ok {
		return fmt.Sprintf("%s/%s/%d.png", d.Manifest().ThemeInfo.BasePath, entry.Image, playerId)
	}
	return ""
}

// GetTileAssetPath returns the full path to a specific terrain+player PNG file
func (d *DefaultTheme) GetTileAssetPath(terrainId, playerId int32) string {
	if entry, ok := d.Manifest().Terrains[terrainId]; ok {
		// Only city terrains use player colors; all others use player 0 (neutral)
		effectivePlayer := int32(0)
		if d.CityTerrains()[terrainId] {
			effectivePlayer = playerId
		}
		return fmt.Sprintf("%s/%s/%d.png", d.Manifest().ThemeInfo.BasePath, entry.Image, effectivePlayer)
	}
	return ""
}

// GetAssetPathForTemplate is a helper for templates to get either unit or tile paths
func (d *DefaultTheme) GetAssetPathForTemplate(assetType string, assetId, playerId int32) string {
	switch assetType {
	case "unit":
		return d.GetUnitAssetPath(assetId, playerId)
	case "tile", "terrain":
		return d.GetTileAssetPath(assetId, playerId)
	default:
		return ""
	}
}
