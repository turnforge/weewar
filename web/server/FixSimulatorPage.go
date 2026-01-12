package server

import (
	"net/http"
	"sort"

	goal "github.com/panyam/goapplib"
	"github.com/turnforge/weewar/lib"
)

type FixSimulatorPage struct {
	BasePage
	Header Header

	// URL parameters for pre-populating form
	FixingUnit       string
	FixingUnitHealth string
	InjuredUnit      string
	NumSimulations   string

	// Units data for dropdowns
	// FixingUnits only includes units with fix_value > 0
	FixingUnits []UnitType
	// AllUnits includes all units (for injured unit dropdown)
	AllUnits []UnitType
	Theme    string
}

func (p *FixSimulatorPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*WeewarApp]) (err error, finished bool) {
	p.DisableSplashScreen = true
	p.Title = "Fix Simulator"
	p.Header.Load(r, w, app)

	// Read URL parameters for pre-populating form
	query := r.URL.Query()
	p.FixingUnit = getQueryOrDefault(query, "fixingUnit", "39")       // Default: Aircraft Carrier (has fix ability)
	p.FixingUnitHealth = getQueryOrDefault(query, "fixingUnitHealth", "10")
	p.InjuredUnit = getQueryOrDefault(query, "injuredUnit", "17")     // Default: Bomber
	p.NumSimulations = getQueryOrDefault(query, "numSims", "1000")
	p.Theme = getQueryOrDefaultStr(query, "theme", "default")

	// Load unit data
	p.loadUnitData()

	return nil, false
}

func (p *FixSimulatorPage) loadUnitData() {
	themeName := p.Theme
	useTheme := themeName != "default"
	tm := GetThemeManager()
	rulesEngine := lib.DefaultRulesEngine()

	// Load all units, separating fixing units (with fix_value > 0)
	p.FixingUnits = []UnitType{}
	p.AllUnits = []UnitType{}

	for _, unitID := range AllowedUnitIDs {
		unitData, err := rulesEngine.GetUnitData(unitID)
		if unitData != nil && err == nil {
			iconDataURL := tm.GetUnitIconURL(unitID, useTheme, themeName)
			themedUnitData := unitData
			themedUnitData.Name = tm.GetUnitName(unitID, unitData.Name, useTheme, themeName)

			unitType := UnitType{
				UnitDefinition: themedUnitData,
				IconDataURL:    iconDataURL,
			}

			p.AllUnits = append(p.AllUnits, unitType)

			// Only add to FixingUnits if it has fix_value > 0
			if unitData.FixValue > 0 {
				p.FixingUnits = append(p.FixingUnits, unitType)
			}
		}
	}

	// Sort units alphabetically by name
	sortAlphabetically := true
	if sortAlphabetically {
		sort.Slice(p.AllUnits, func(i, j int) bool { return p.AllUnits[i].Name < p.AllUnits[j].Name })
		sort.Slice(p.FixingUnits, func(i, j int) bool { return p.FixingUnits[i].Name < p.FixingUnits[j].Name })
	} else {
		sort.Slice(p.AllUnits, func(i, j int) bool { return p.AllUnits[i].Id < p.AllUnits[j].Id })
		sort.Slice(p.FixingUnits, func(i, j int) bool { return p.FixingUnits[i].Id < p.FixingUnits[j].Id })
	}
}
