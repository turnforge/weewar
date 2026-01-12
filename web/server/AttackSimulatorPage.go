package server

import (
	"net/http"
	"sort"
	"strconv"

	goal "github.com/panyam/goapplib"
	"github.com/turnforge/lilbattle/lib"
)

type AttackSimulatorPage struct {
	BasePage
	Header Header

	// URL parameters for pre-populating form
	AttackerUnit    string
	AttackerTerrain string
	AttackerHealth  string
	DefenderUnit    string
	DefenderTerrain string
	DefenderHealth  string
	WoundBonus      string
	NumSimulations  string

	// Terrain and unit data for dropdowns
	AllTerrains []TerrainType
	AllUnits    []UnitType
	Theme       string
}

func (p *AttackSimulatorPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	p.DisableSplashScreen = true
	p.Title = "Attack Simulator"
	p.Header.Load(r, w, app)

	// Read URL parameters for pre-populating form
	query := r.URL.Query()
	p.AttackerUnit = getQueryOrDefault(query, "attackerUnit", "4")       // Default: Hovercraft
	p.AttackerTerrain = getQueryOrDefault(query, "attackerTerrain", "1") // Default: Grass
	p.AttackerHealth = getQueryOrDefault(query, "attackerHealth", "10")
	p.DefenderUnit = getQueryOrDefault(query, "defenderUnit", "1")       // Default: Soldier
	p.DefenderTerrain = getQueryOrDefault(query, "defenderTerrain", "1") // Default: Grass
	p.DefenderHealth = getQueryOrDefault(query, "defenderHealth", "10")
	p.WoundBonus = getQueryOrDefault(query, "woundBonus", "0")
	p.NumSimulations = getQueryOrDefault(query, "numSims", "1000")
	p.Theme = getQueryOrDefaultStr(query, "theme", "default")

	// Load terrain and unit data
	p.loadTerrainAndUnitData()

	return nil, false
}

func (p *AttackSimulatorPage) loadTerrainAndUnitData() {
	themeName := p.Theme
	useTheme := themeName != "default"
	tm := GetThemeManager()
	rulesEngine := lib.DefaultRulesEngine()

	// Load all terrains (nature + city)
	p.AllTerrains = []TerrainType{}
	for i := int32(0); i <= 30; i++ {
		terrainData, err := rulesEngine.GetTerrainData(i)
		if err == nil && terrainData != nil && terrainData.Id != 0 { // Skip Clear (ID 0)
			iconDataURL := tm.GetTerrainIconURL(i, useTheme, themeName)
			terrainName := tm.GetTerrainName(i, terrainData.Name, useTheme, themeName)

			terrain := TerrainType{
				TerrainData: TerrainData{
					ID:           terrainData.Id,
					Name:         terrainName,
					BaseMoveCost: 1.0,
					DefenseBonus: 0.0,
				},
				IconDataURL:     iconDataURL,
				HasPlayerColors: false,
			}
			p.AllTerrains = append(p.AllTerrains, terrain)
		}
	}

	// Sort terrains alphabetically by name
	sortAlphabetically := true
	if sortAlphabetically {
		sort.Slice(p.AllTerrains, func(i, j int) bool {
			return p.AllTerrains[i].Name < p.AllTerrains[j].Name
		})
	} else {
		sort.Slice(p.AllTerrains, func(i, j int) bool {
			return p.AllTerrains[i].ID < p.AllTerrains[j].ID
		})
	}

	// Load all units
	p.AllUnits = []UnitType{}
	for _, unitID := range AllowedUnitIDs {
		unitData, err := rulesEngine.GetUnitData(unitID)
		if unitData != nil && err == nil {
			iconDataURL := tm.GetUnitIconURL(unitID, useTheme, themeName)
			themedUnitData := unitData
			themedUnitData.Name = tm.GetUnitName(unitID, unitData.Name, useTheme, themeName)

			p.AllUnits = append(p.AllUnits, UnitType{
				UnitDefinition: themedUnitData,
				IconDataURL:    iconDataURL,
			})
		}
	}

	// Sort units alphabetically by name
	if sortAlphabetically {
		sort.Slice(p.AllUnits, func(i, j int) bool { return p.AllUnits[i].Name < p.AllUnits[j].Name })
	} else {
		sort.Slice(p.AllUnits, func(i, j int) bool { return p.AllUnits[i].Id < p.AllUnits[j].Id })
	}
}

// Helper to get query parameter or return default value (validates as number)
func getQueryOrDefault(query map[string][]string, key string, defaultVal string) string {
	if vals, ok := query[key]; ok && len(vals) > 0 {
		// Validate it's a number if needed
		if _, err := strconv.Atoi(vals[0]); err == nil {
			return vals[0]
		}
	}
	return defaultVal
}

// Helper to get query parameter or return default value (string, no validation)
func getQueryOrDefaultStr(query map[string][]string, key string, defaultVal string) string {
	if vals, ok := query[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return defaultVal
}
