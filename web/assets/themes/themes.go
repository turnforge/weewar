package themes

import (
	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// Theme interface provides metadata about units and terrains in a theme
// This is the "lightweight" interface for template rendering and metadata queries
// Mirrors BaseTheme.ts but focused on data, not asset loading
type Theme interface {
	// GetUnitName returns the display name for a unit in this theme
	GetUnitName(unitId int32) string

	// GetTerrainName returns the display name for a terrain in this theme
	GetTerrainName(terrainId int32) string

	// GetUnitDescription returns the description for a unit (if available)
	GetUnitDescription(unitId int32) string

	// GetTerrainDescription returns the description for a terrain (if available)
	GetTerrainDescription(terrainId int32) string

	// GetUnitPath returns the file path for a unit's base asset
	// For SVG themes: returns template path (e.g., "Units/Knight.svg")
	// For PNG themes: returns directory path (e.g., "Units/1")
	GetUnitPath(unitId int32) string

	// GetTilePath returns the file path for a terrain's base asset
	GetTilePath(terrainId int32) string

	// IsCityTile checks if a terrain is a city/building tile (colored by player)
	IsCityTile(terrainId int32) bool

	// IsNatureTile checks if a terrain is a nature tile (neutral only)
	IsNatureTile(terrainId int32) bool

	// IsBridgeTile checks if a terrain is a bridge
	IsBridgeTile(terrainId int32) bool

	// GetThemeInfo returns metadata about the theme
	GetThemeInfo() *v1.ThemeInfo

	// GetAvailableUnits returns all unit IDs available in this theme
	GetAvailableUnits() []int32

	// GetAvailableTerrains returns all terrain IDs available in this theme
	GetAvailableTerrains() []int32

	// HasUnit checks if a unit ID exists in this theme
	HasUnit(unitId int32) bool

	// HasTerrain checks if a terrain ID exists in this theme
	HasTerrain(terrainId int32) bool
}

// ThemeAssets interface handles asset loading and rendering
// This is the "heavy" interface for actual asset operations
// Separated from Theme to allow phased implementation
type ThemeAssets interface {
	// GetUnitAsset returns the asset for a unit (either path or rendered SVG)
	// For PNG themes: returns AssetResult with Type=PATH, Data="/static/assets/themes/default/Units/1/0.png"
	// For SVG themes: returns AssetResult with Type=SVG, Data="<svg>...</svg>" (with player colors)
	GetUnitAsset(unitId, playerId int32) (*v1.AssetResult, error)

	// GetTileAsset returns the asset for a terrain tile
	GetTileAsset(tileId, playerId int32) (*v1.AssetResult, error)

	// LoadUnit loads and processes a unit SVG template with player colors
	// Returns the SVG markup as a string
	LoadUnit(unitId, playerId int32) (string, error)

	// LoadTile loads and processes a terrain SVG template with optional player colors
	LoadTile(terrainId, playerId int32) (string, error)

	// ApplyPlayerColors applies player color transformations to SVG content
	// This is the Go equivalent of BaseTheme.applyPlayerColors()
	ApplyPlayerColors(svgContent string, playerId int32) (string, error)
}

// PlayerColors maps player IDs to their color schemes
// Matches PLAYER_COLORS in BaseTheme.ts
var PlayerColors = map[int32]*v1.PlayerColor{
	0:  {Primary: "#888888", Secondary: "#666666"}, // Neutral/unowned
	1:  {Primary: "#f87171", Secondary: "#dc2626"}, // RED
	2:  {Primary: "#60a5fa", Secondary: "#2563eb"}, // BLUE
	3:  {Primary: "#4ade80", Secondary: "#16a34a"}, // GREEN
	4:  {Primary: "#facc15", Secondary: "#ca8a04"}, // YELLOW
	5:  {Primary: "#fb923c", Secondary: "#ea580c"}, // ORANGE
	6:  {Primary: "#c084fc", Secondary: "#9333ea"}, // PURPLE
	7:  {Primary: "#f472b6", Secondary: "#db2777"}, // PINK
	8:  {Primary: "#22d3ee", Secondary: "#0891b2"}, // CYAN
	9:  {Primary: "#22d3ee", Secondary: "#0891b2"}, // CYAN (duplicate)
	10: {Primary: "#22d3ee", Secondary: "#0891b2"}, // CYAN (duplicate)
	11: {Primary: "#22d3ee", Secondary: "#0891b2"}, // CYAN (duplicate)
	12: {Primary: "#22d3ee", Secondary: "#0891b2"}, // CYAN (duplicate)
}

// Terrain classification constants (matches BaseTheme.ts)
var (
	CityTerrainIDs   = []int32{1, 2, 3, 6, 16, 20, 21, 25}
	NatureTerrainIDs = []int32{4, 5, 7, 8, 9, 10, 12, 14, 15, 23, 26}
	BridgeTerrainIDs = []int32{17, 18, 19}
	RoadTerrainID    = 22
)

// Helper functions for terrain classification
func IsCityTerrain(terrainId int32) bool {
	for _, id := range CityTerrainIDs {
		if id == terrainId {
			return true
		}
	}
	return false
}

func IsNatureTerrain(terrainId int32) bool {
	for _, id := range NatureTerrainIDs {
		if id == terrainId {
			return true
		}
	}
	return false
}

func IsBridgeTerrain(terrainId int32) bool {
	for _, id := range BridgeTerrainIDs {
		if id == terrainId {
			return true
		}
	}
	return false
}
