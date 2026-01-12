package themes

import (
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// BaseTheme provides common functionality for all themes
// Mirrors BaseTheme.ts but focused on metadata, not asset loading
type BaseTheme struct {
	manifest     *v1.ThemeManifest
	cityTerrains map[int32]bool // Terrains that use player colors (from RulesEngine)
}

// Default player colors - used when manifest doesn't specify playerColors
var defaultPlayerColors = map[int32]*v1.PlayerColor{
	0:  {Primary: "#888888", Secondary: "#666666", Name: "Neutral"},
	1:  {Primary: "#60a5fa", Secondary: "#2563eb", Name: "Blue"},
	2:  {Primary: "#f87171", Secondary: "#dc2626", Name: "Red"},
	3:  {Primary: "#facc15", Secondary: "#ca8a04", Name: "Yellow"},
	4:  {Primary: "#f0f0f0", Secondary: "#888888", Name: "White"},
	5:  {Primary: "#f472b6", Secondary: "#db2777", Name: "Pink"},
	6:  {Primary: "#fb923c", Secondary: "#ea580c", Name: "Orange"},
	7:  {Primary: "#1f2937", Secondary: "#111827", Name: "Black"},
	8:  {Primary: "#2dd4bf", Secondary: "#14b8a6", Name: "Teal"},
	9:  {Primary: "#1e3a8a", Secondary: "#1e40af", Name: "Navy Blue"},
	10: {Primary: "#a16207", Secondary: "#854d0e", Name: "Brown"},
	11: {Primary: "#22d3ee", Secondary: "#0891b2", Name: "Cyan"},
	12: {Primary: "#c084fc", Secondary: "#9333ea", Name: "Purple"},
}

// NewBaseTheme creates a new BaseTheme from a pre-loaded manifest
// cityTerrains is a map of terrain IDs that use player colors (from RulesEngine.TerrainTypes)
func NewBaseTheme(manifest *v1.ThemeManifest, cityTerrains map[int32]bool) *BaseTheme {
	// Populate default player colors if not specified in manifest
	if len(manifest.PlayerColors) == 0 {
		manifest.PlayerColors = make(map[int32]*v1.PlayerColor)
		for k, v := range defaultPlayerColors {
			manifest.PlayerColors[k] = v
		}
	}
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
