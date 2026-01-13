package assets

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// Embed the theme mapping files at compile time
//
//go:embed themes/fantasy/mapping.json
//go:embed themes/modern/mapping.json
var EmbeddedMappings embed.FS

// LoadThemeManifest loads a theme manifest from embedded files
func LoadThemeManifest(themeName string) (*v1.ThemeManifest, error) {
	// Construct the path to the embedded mapping.json
	path := fmt.Sprintf("themes/%s/mapping.json", themeName)

	data, err := EmbeddedMappings.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded theme %s: %w", themeName, err)
	}

	return ParseThemeManifest(data)
}

// ParseThemeManifest converts JSON bytes into a ThemeManifest proto
func ParseThemeManifest(data []byte) (*v1.ThemeManifest, error) {
	// Parse the JSON into a temporary structure
	// Note: JSON has string keys, proto uses int32 keys
	var rawManifest struct {
		ThemeInfo *v1.ThemeInfo                 `json:"themeInfo,omitempty"`
		Units     map[string]*v1.UnitMapping    `json:"units"`
		Terrains  map[string]*v1.TerrainMapping `json:"terrains"`
	}

	if err := json.Unmarshal(data, &rawManifest); err != nil {
		return nil, fmt.Errorf("failed to parse theme manifest: %w", err)
	}

	// Convert string keys to int32
	manifest := &v1.ThemeManifest{
		ThemeInfo: rawManifest.ThemeInfo,
		Units:     make(map[int32]*v1.UnitMapping),
		Terrains:  make(map[int32]*v1.TerrainMapping),
	}

	for k, v := range rawManifest.Units {
		id, _ := strconv.ParseInt(k, 10, 32)
		manifest.Units[int32(id)] = v
	}

	for k, v := range rawManifest.Terrains {
		id, _ := strconv.ParseInt(k, 10, 32)
		manifest.Terrains[int32(id)] = v
	}

	return manifest, nil
}
