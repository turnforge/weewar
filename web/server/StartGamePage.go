package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
)

var DefaultRulesEngine *weewar.RulesEngine

func init() {
	var err error
	// TODO - only for dev
	DefaultRulesEngine, err = weewar.LoadRulesEngineFromFile(weewar.DevDataPath("data/rules-data.json"))
	if err != nil {
		panic(fmt.Sprintf("Failed to load rules engine: %v", err))
	}
}

type StartGamePage struct {
	BasePage
	Header    Header
	World     *protos.World
	WorldId   string
	UnitTypes []UnitType
}

func (p *StartGamePage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	// Get worldId from query parameter (optional)
	p.WorldId = r.URL.Query().Get("worldId")

	p.Title = "New Game"
	p.Header.Load(r, w, vc)

	// If a worldId is provided, fetch the world data
	if p.WorldId != "" {
		// Fetch the World using the client manager
		client, err := vc.ClientMgr.GetWorldsSvcClient()
		if err != nil {
			log.Printf("Error getting Worlds client: %v", err)
			// Don't fail the page, just log the error
			p.WorldId = ""
		} else {
			req := &protos.GetWorldRequest{
				Id: p.WorldId,
			}

			resp, err := client.GetWorld(context.Background(), req)
			if err != nil {
				log.Printf("Error fetching World %s: %v", p.WorldId, err)
				// Don't fail the page, just clear the worldId
				p.WorldId = ""
			} else if resp.World != nil {
				p.World = resp.World
				p.Title = "New Game - " + p.World.Name
			}
		}
	}

	// Load unit types for unit restrictions UI
	p.loadUnitTypes()

	return nil, false
}

// loadUnitTypes populates the UnitTypes field for the unit restrictions UI
func (p *StartGamePage) loadUnitTypes() {
	// Load unit types with icons from rules engine
	p.UnitTypes = []UnitType{}

	// Get all available unit types from the rules engine
	rulesEngine := DefaultRulesEngine

	// If rules engine is not populated, fall back to GetUnitData function which uses the unitDataWorld
	unitIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29}

	// If rules engine has units loaded, use those; otherwise use the static list
	if rulesEngine.GetLoadedUnitsCount() > 0 {
		// Use units from rules engine
		for unitID := range rulesEngine.Units {
			unitData := rulesEngine.Units[unitID]
			if unitData != nil {
				// Use web-accessible static URL path for the unit asset
				iconDataURL := fmt.Sprintf("/static/assets/v1/Units/%d/0.png", unitID)

				p.UnitTypes = append(p.UnitTypes, UnitType{
					ID:          unitData.Id,
					Name:        unitData.Name,
					IconDataURL: iconDataURL,
				})
			}
		}
	} else {
		// Fall back to static unit data world
		for _, unitID := range unitIDs {
			unitData := weewar.GetUnitData(unitID)
			if unitData != nil {
				// Use web-accessible static URL path for the unit asset
				iconDataURL := fmt.Sprintf("/static/assets/v1/Units/%d/0.png", unitID)

				p.UnitTypes = append(p.UnitTypes, UnitType{
					ID:          int32(unitData.ID),
					Name:        unitData.Name,
					IconDataURL: iconDataURL,
				})
			}
		}
	}
}

func (p *StartGamePage) Copy() View {
	return &StartGamePage{}
}
