package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// ThemeMapping represents the mapping.json structure
type ThemeMapping struct {
	Units    map[string]ThemeMappingEntry `json:"units"`
	Terrains map[string]ThemeMappingEntry `json:"terrains"`
}

type ThemeMappingEntry struct {
	Old   string `json:"old"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// ThemeManager handles theme asset resolution and caching
type ThemeManager struct {
	cache map[string]*ThemeMapping
	mu    sync.RWMutex
}

var themeManager = &ThemeManager{
	cache: make(map[string]*ThemeMapping),
}

// GetThemeManager returns the singleton theme manager
func GetThemeManager() *ThemeManager {
	return themeManager
}

// LoadThemeMapping loads and caches a theme's mapping.json
func (tm *ThemeManager) LoadThemeMapping(themeName string) (*ThemeMapping, error) {
	// Check cache first
	tm.mu.RLock()
	if mapping, exists := tm.cache[themeName]; exists {
		tm.mu.RUnlock()
		return mapping, nil
	}
	tm.mu.RUnlock()

	// Load from file - path relative to the web directory
	// The assets are in web/static/assets/themes/<themeName>/mapping.json (relative to project root)
	mappingPath := filepath.Join("web", "static", "assets", "themes", themeName, "mapping.json")
	data, err := os.ReadFile(mappingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load theme mapping for %s: %w", themeName, err)
	}

	var mapping ThemeMapping
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("failed to parse theme mapping for %s: %w", themeName, err)
	}

	// Cache it
	tm.mu.Lock()
	tm.cache[themeName] = &mapping
	tm.mu.Unlock()

	log.Printf("Loaded theme mapping for %s: %d terrains, %d units",
		themeName, len(mapping.Terrains), len(mapping.Units))

	return &mapping, nil
}

// GetTerrainIconURL returns the icon URL for a terrain ID
func (tm *ThemeManager) GetTerrainIconURL(terrainID int32, useTheme bool, themeName string) string {
	if !useTheme {
		// Use PNG assets
		return fmt.Sprintf("/static/assets/v1/Tiles/%d/0.png", terrainID)
	}

	// Try to get from theme mapping
	mapping, err := tm.LoadThemeMapping(themeName)
	if err != nil {
		log.Printf("Failed to load theme mapping: %v", err)
		// Fallback to PNG
		return fmt.Sprintf("/static/assets/v1/Tiles/%d/0.png", terrainID)
	}

	// Look up the terrain in the mapping
	terrainKey := fmt.Sprintf("%d", terrainID)
	if entry, exists := mapping.Terrains[terrainKey]; exists {
		// For PNG themes (like default), image is a directory - append /0.png for neutral
		// For SVG themes, image is the full file path
		if themeName == "default" {
			return fmt.Sprintf("/static/assets/themes/%s/%s/0.png", themeName, entry.Image)
		}
		return fmt.Sprintf("/static/assets/themes/%s/%s", themeName, entry.Image)
	}

	// Terrain not in theme, fallback to PNG
	log.Printf("Terrain %d not found in theme %s, using PNG fallback", terrainID, themeName)
	return fmt.Sprintf("/static/assets/v1/Tiles/%d/0.png", terrainID)
}

// GetUnitIconURL returns the icon URL for a unit ID
func (tm *ThemeManager) GetUnitIconURL(unitID int32, useTheme bool, themeName string) string {
	if !useTheme {
		// Use PNG assets
		return fmt.Sprintf("/static/assets/v1/Units/%d/0.png", unitID)
	}

	// Try to get from theme mapping
	mapping, err := tm.LoadThemeMapping(themeName)
	if err != nil {
		log.Printf("Failed to load theme mapping: %v", err)
		// Fallback to PNG
		return fmt.Sprintf("/static/assets/v1/Units/%d/0.png", unitID)
	}

	// Look up the unit in the mapping
	unitKey := fmt.Sprintf("%d", unitID)
	if entry, exists := mapping.Units[unitKey]; exists {
		// For PNG themes (like default), image is a directory - append /0.png for neutral
		// For SVG themes, image is the full file path
		if themeName == "default" {
			return fmt.Sprintf("/static/assets/themes/%s/%s/0.png", themeName, entry.Image)
		}
		return fmt.Sprintf("/static/assets/themes/%s/%s", themeName, entry.Image)
	}

	// Unit not in theme, fallback to PNG
	log.Printf("Unit %d not found in theme %s, using PNG fallback", unitID, themeName)
	return fmt.Sprintf("/static/assets/v1/Units/%d/0.png", unitID)
}

// GetTerrainName returns the themed name for a terrain, or the default name if not themed
func (tm *ThemeManager) GetTerrainName(terrainID int32, defaultName string, useTheme bool, themeName string) string {
	if !useTheme {
		return defaultName
	}

	mapping, err := tm.LoadThemeMapping(themeName)
	if err != nil {
		return defaultName
	}

	terrainKey := fmt.Sprintf("%d", terrainID)
	if entry, exists := mapping.Terrains[terrainKey]; exists && entry.Name != "" {
		return entry.Name
	}
	return defaultName
}

// GetUnitName returns the themed name for a unit, or the default name if not themed
func (tm *ThemeManager) GetUnitName(unitID int32, defaultName string, useTheme bool, themeName string) string {
	if !useTheme {
		return defaultName
	}

	mapping, err := tm.LoadThemeMapping(themeName)
	if err != nil {
		return defaultName
	}

	unitKey := fmt.Sprintf("%d", unitID)
	if entry, exists := mapping.Units[unitKey]; exists && entry.Name != "" {
		return entry.Name
	}
	return defaultName
}

// ClearCache clears the theme mapping cache
func (tm *ThemeManager) ClearCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cache = make(map[string]*ThemeMapping)
	log.Println("Theme mapping cache cleared")
}
