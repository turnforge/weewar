package themes

import (
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// BaseTheme provides common functionality for all themes
// Mirrors BaseTheme.ts but focused on metadata, not asset loading
type BaseTheme struct {
	manifest     *v1.ThemeManifest
	cityTerrains map[int32]bool // Terrains that use player colors (from RulesEngine)
}

// NewBaseTheme creates a new BaseTheme from a pre-loaded manifest
// cityTerrains is a map of terrain IDs that use player colors (from RulesEngine.TerrainTypes)
func NewBaseTheme(manifest *v1.ThemeManifest, cityTerrains map[int32]bool) *BaseTheme {
	return &BaseTheme{
		manifest:     manifest,
		cityTerrains: cityTerrains,
	}
}

// Manifest returns the underlying ThemeManifest for subclasses
func (b *BaseTheme) Manifest() *v1.ThemeManifest {
	return b.manifest
}

// CityTerrains returns the city terrains map for subclasses
func (b *BaseTheme) CityTerrains() map[int32]bool {
	return b.cityTerrains
}

func (b *BaseTheme) GetUnitName(unitId int32) string {
	if mapping, ok := b.manifest.Units[unitId]; ok {
		return mapping.Name
	}
	return ""
}

func (b *BaseTheme) GetTerrainName(terrainId int32) string {
	if mapping, ok := b.manifest.Terrains[terrainId]; ok {
		return mapping.Name
	}
	return ""
}

func (b *BaseTheme) GetUnitDescription(unitId int32) string {
	if mapping, ok := b.manifest.Units[unitId]; ok {
		return mapping.Description
	}
	return ""
}

func (b *BaseTheme) GetTerrainDescription(terrainId int32) string {
	if mapping, ok := b.manifest.Terrains[terrainId]; ok {
		return mapping.Description
	}
	return ""
}

func (b *BaseTheme) GetUnitPath(unitId int32) string {
	if mapping, ok := b.manifest.Units[unitId]; ok {
		// Return relative path from theme base
		return mapping.Image
	}
	return ""
}

func (b *BaseTheme) GetTilePath(terrainId int32) string {
	if mapping, ok := b.manifest.Terrains[terrainId]; ok {
		// Return relative path from theme base
		return mapping.Image
	}
	return ""
}

func (b *BaseTheme) GetThemeInfo() *v1.ThemeInfo {
	return b.manifest.ThemeInfo
}

func (b *BaseTheme) GetAvailableUnits() []int32 {
	units := make([]int32, 0, len(b.manifest.Units))
	for id := range b.manifest.Units {
		units = append(units, id)
	}
	return units
}

func (b *BaseTheme) GetAvailableTerrains() []int32 {
	terrains := make([]int32, 0, len(b.manifest.Terrains))
	for id := range b.manifest.Terrains {
		terrains = append(terrains, id)
	}
	return terrains
}

func (b *BaseTheme) HasUnit(unitId int32) bool {
	_, ok := b.manifest.Units[unitId]
	return ok
}

func (b *BaseTheme) HasTerrain(terrainId int32) bool {
	_, ok := b.manifest.Terrains[terrainId]
	return ok
}

func (b *BaseTheme) GetEffectivePlayer(terrainId, playerId int32) int32 {
	if b.cityTerrains[terrainId] {
		return playerId
	}
	return 0
}

func (b *BaseTheme) GetPlayerColor(playerId int32) *v1.PlayerColor {
	if color, ok := b.manifest.PlayerColors[playerId]; ok {
		return color
	}
	// Fallback to neutral if not found
	if color, ok := b.manifest.PlayerColors[0]; ok {
		return color
	}
	return nil
}
