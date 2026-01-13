package themes_test

import (
	"fmt"
	"testing"

	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/web/assets/themes"
)

// Helper to get cityTerrains for tests
func testCityTerrains() map[int32]bool {
	return lib.DefaultRulesEngine().GetCityTerrains()
}

// Example showing how to create and use themes
func ExampleCreateTheme() {
	cityTerrains := testCityTerrains()

	// Create themes using the registry
	defaultTheme, _ := themes.CreateTheme("default", cityTerrains)
	fantasyTheme, _ := themes.CreateTheme("fantasy", cityTerrains)
	modernTheme, _ := themes.CreateTheme("modern", cityTerrains)

	// Use default theme (PNG-based)
	fmt.Println(defaultTheme.GetUnitName(1))      // "Infantry"
	fmt.Println(defaultTheme.GetThemeInfo().Name) // "Default (PNG)"

	// Use fantasy theme (SVG-based, loaded from embedded mapping.json)
	fmt.Println(fantasyTheme.GetUnitName(1))      // "Peasant"
	fmt.Println(fantasyTheme.GetThemeInfo().Name) // "Medieval Fantasy"

	// Use modern theme (SVG-based, loaded from embedded mapping.json)
	fmt.Println(modernTheme.GetUnitName(1))      // "Infantry"
	fmt.Println(modernTheme.GetThemeInfo().Name) // "Modern Military"

	// Output:
	// Infantry
	// Default (PNG)
	// Peasant
	// Medieval Fantasy
	// Infantry
	// Modern Military
}

// Example showing how to get asset paths
func Example_assetPaths() {
	cityTerrains := testCityTerrains()
	defaultTheme := themes.NewDefaultTheme(cityTerrains)
	fantasyTheme, _ := themes.NewFantasyTheme(cityTerrains)

	// Default theme returns full PNG paths
	unitPath := defaultTheme.GetAssetPathForTemplate("unit", 1, 2)
	fmt.Println(unitPath) // /static/assets/themes/default/Units/1/2.png

	// Fantasy theme returns SVG template paths
	fantasyUnitPath := fantasyTheme.GetUnitAssetPath(1)
	fmt.Println(fantasyUnitPath) // /static/assets/themes/fantasy/Units/Peasant.svg

	// Output:
	// /static/assets/themes/default/Units/1/2.png
	// /static/assets/themes/fantasy/Units/Peasant.svg
}

// Test that all themes can be created
func TestAllThemes(t *testing.T) {
	cityTerrains := testCityTerrains()
	themeNames := []string{"default", "fantasy", "modern"}

	for _, name := range themeNames {
		theme, err := themes.CreateTheme(name, cityTerrains)
		if err != nil {
			t.Errorf("Failed to create theme %s: %v", name, err)
			continue
		}

		info := theme.GetThemeInfo()
		if info == nil {
			t.Errorf("Theme %s returned nil info", name)
			continue
		}

		t.Logf("Created theme: %s (version %s, type: %s)",
			info.Name, info.Version, info.AssetType)

		// Test that we can get unit names
		unitName := theme.GetUnitName(1)
		if unitName == "" {
			t.Errorf("Theme %s returned empty name for unit 1", name)
		}

		// Test that we can get terrain names
		terrainName := theme.GetTerrainName(1)
		if terrainName == "" {
			t.Errorf("Theme %s returned empty name for terrain 1", name)
		}
	}
}

// Test GetEffectivePlayer for terrain rendering
func TestGetEffectivePlayer(t *testing.T) {
	cityTerrains := testCityTerrains()
	theme := themes.NewDefaultTheme(cityTerrains)

	// City tiles should return the actual player ID
	if theme.GetEffectivePlayer(1, 2) != 2 { // Land Base with player 2
		t.Error("Land Base should return actual player ID for city terrain")
	}

	// Nature tiles should always return 0 (neutral)
	if theme.GetEffectivePlayer(5, 2) != 0 { // Grass with player 2
		t.Error("Grass should return 0 (neutral) for nature terrain")
	}

	// Water tiles should always return 0 (neutral)
	if theme.GetEffectivePlayer(10, 1) != 0 { // Water with player 1
		t.Error("Water should return 0 (neutral) for water terrain")
	}
}
