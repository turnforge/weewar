package themes

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

//go:embed default/mapping.json
var defaultMappingJSON []byte

// ThemeMappingEntry represents a single unit or terrain entry in mapping.json
type ThemeMappingEntry struct {
	Old   string `json:"old"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// ThemeInfo from mapping.json
type ThemeInfoJSON struct {
	Name                string `json:"name"`
	Version             string `json:"version"`
	BasePath            string `json:"base_path"`
	AssetType           string `json:"asset_type"`
	NeedsPostProcessing bool   `json:"needs_post_processing"`
}

// ThemeMapping represents the full mapping.json structure
type ThemeMappingJSON struct {
	ThemeInfo      ThemeInfoJSON                `json:"themeInfo"`
	Units          map[string]ThemeMappingEntry `json:"units"`
	Terrains       map[string]ThemeMappingEntry `json:"terrains"`
	NatureTerrains []int32                      `json:"natureTerrains"`
}

// DefaultTheme implements the Theme interface for PNG-based default assets
// Loads data from embedded mapping.json
type DefaultTheme struct {
	themeInfo      ThemeInfoJSON
	units          map[int32]ThemeMappingEntry
	terrains       map[int32]ThemeMappingEntry
	natureTerrains map[int32]bool
}

// NewDefaultTheme creates a new default theme instance by parsing embedded mapping.json
func NewDefaultTheme() *DefaultTheme {
	var mapping ThemeMappingJSON
	if err := json.Unmarshal(defaultMappingJSON, &mapping); err != nil {
		panic(fmt.Sprintf("failed to parse embedded default theme mapping: %v", err))
	}

	// Convert string keys to int32 for units
	units := make(map[int32]ThemeMappingEntry)
	for key, entry := range mapping.Units {
		id, err := strconv.ParseInt(key, 10, 32)
		if err == nil {
			units[int32(id)] = entry
		}
	}

	// Convert string keys to int32 for terrains
	terrains := make(map[int32]ThemeMappingEntry)
	for key, entry := range mapping.Terrains {
		id, err := strconv.ParseInt(key, 10, 32)
		if err == nil {
			terrains[int32(id)] = entry
		}
	}

	// Build nature terrains lookup
	natureTerrains := make(map[int32]bool)
	for _, id := range mapping.NatureTerrains {
		natureTerrains[id] = true
	}

	return &DefaultTheme{
		themeInfo:      mapping.ThemeInfo,
		units:          units,
		terrains:       terrains,
		natureTerrains: natureTerrains,
	}
}

func (d *DefaultTheme) GetUnitName(unitId int32) string {
	if entry, ok := d.units[unitId]; ok {
		return entry.Name
	}
	return ""
}

func (d *DefaultTheme) GetTerrainName(terrainId int32) string {
	if entry, ok := d.terrains[terrainId]; ok {
		return entry.Name
	}
	return ""
}

func (d *DefaultTheme) GetUnitDescription(unitId int32) string {
	// Default theme has no custom descriptions
	return ""
}

func (d *DefaultTheme) GetTerrainDescription(terrainId int32) string {
	// Default theme has no custom descriptions
	return ""
}

// GetUnitPath returns the directory path for a unit's assets
// For PNG themes, this is the directory containing player-colored variants
func (d *DefaultTheme) GetUnitPath(unitId int32) string {
	if entry, ok := d.units[unitId]; ok {
		return fmt.Sprintf("%s/%s", d.themeInfo.BasePath, entry.Image)
	}
	return ""
}

// GetTilePath returns the directory path for a terrain's assets
func (d *DefaultTheme) GetTilePath(terrainId int32) string {
	if entry, ok := d.terrains[terrainId]; ok {
		return fmt.Sprintf("%s/%s", d.themeInfo.BasePath, entry.Image)
	}
	return ""
}

// GetUnitAssetPath returns the full path to a specific unit+player PNG file
func (d *DefaultTheme) GetUnitAssetPath(unitId, playerId int32) string {
	if entry, ok := d.units[unitId]; ok {
		return fmt.Sprintf("%s/%s/%d.png", d.themeInfo.BasePath, entry.Image, playerId)
	}
	return ""
}

// GetTileAssetPath returns the full path to a specific terrain+player PNG file
func (d *DefaultTheme) GetTileAssetPath(terrainId, playerId int32) string {
	if entry, ok := d.terrains[terrainId]; ok {
		// Nature terrains always use player 0 (neutral)
		effectivePlayer := playerId
		if d.natureTerrains[terrainId] {
			effectivePlayer = 0
		}
		return fmt.Sprintf("%s/%s/%d.png", d.themeInfo.BasePath, entry.Image, effectivePlayer)
	}
	return ""
}

func (d *DefaultTheme) IsCityTile(terrainId int32) bool {
	return IsCityTerrain(terrainId)
}

func (d *DefaultTheme) IsNatureTile(terrainId int32) bool {
	return IsNatureTerrain(terrainId)
}

func (d *DefaultTheme) IsBridgeTile(terrainId int32) bool {
	return IsBridgeTerrain(terrainId)
}

func (d *DefaultTheme) GetThemeInfo() *v1.ThemeInfo {
	return &v1.ThemeInfo{
		Name:                d.themeInfo.Name,
		Version:             d.themeInfo.Version,
		BasePath:            d.themeInfo.BasePath,
		AssetType:           d.themeInfo.AssetType,
		NeedsPostProcessing: d.themeInfo.NeedsPostProcessing,
	}
}

func (d *DefaultTheme) GetAvailableUnits() []int32 {
	units := make([]int32, 0, len(d.units))
	for id := range d.units {
		units = append(units, id)
	}
	return units
}

func (d *DefaultTheme) GetAvailableTerrains() []int32 {
	terrains := make([]int32, 0, len(d.terrains))
	for id := range d.terrains {
		terrains = append(terrains, id)
	}
	return terrains
}

func (d *DefaultTheme) HasUnit(unitId int32) bool {
	_, ok := d.units[unitId]
	return ok
}

func (d *DefaultTheme) HasTerrain(terrainId int32) bool {
	_, ok := d.terrains[terrainId]
	return ok
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
