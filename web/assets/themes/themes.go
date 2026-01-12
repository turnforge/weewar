package themes

import (
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// Theme interface provides metadata about units and terrains in a theme
// This is the "lightweight" interface for template rendering and metadata queries
// Mirrors BaseTheme.ts but focused on data, not asset loading
//
// Note: Terrain classification (IsCityTerrain, IsNatureTerrain, etc.) is handled
// by lib.RulesEngine, not by Theme. Theme is purely for rendering concerns.
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

	// GetEffectivePlayer returns the effective player ID for rendering a terrain.
	// City terrains use the actual playerId for player-colored rendering.
	// Non-city terrains (nature, water, etc.) always return 0 (neutral).
	// This is used by renderers to determine which color/variant to use.
	GetEffectivePlayer(terrainId, playerId int32) int32

	// GetPlayerColor returns the color scheme for a player in this theme.
	// Returns nil if the player ID is not found.
	GetPlayerColor(playerId int32) *v1.PlayerColor
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


