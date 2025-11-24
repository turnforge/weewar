package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	protos "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	weewar "github.com/turnforge/weewar/services"
)

var AllowedUnitIDs = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 37, 38, 39, 40, 41, 44}

type StartGamePage struct {
	BasePage
	Header            Header
	World             *protos.World
	WorldData         *protos.WorldData
	WorldId           string
	UnitTypes         []UnitType
	GameConfiguration *protos.GameConfiguration
}

func (p *StartGamePage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	// Get worldId from query parameter (optional)
	p.WorldId = r.URL.Query().Get("worldId")

	// If no worldId provided, redirect to world selection page
	if p.WorldId == "" {
		http.Redirect(w, r, "/worlds/select", http.StatusSeeOther)
		return nil, true
	}

	p.Title = "New Game"
	p.Header.Load(r, w, vc)

	// If a worldId is provided, fetch the world data
	if p.WorldId != "" {
		// Fetch the World using the client manager
		client := vc.ClientMgr.GetWorldsSvcClient()
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
			p.WorldData = resp.WorldData
			p.Title = "New Game - " + p.World.Name
		}
	}

	// Initialize default game configuration
	p.initializeGameConfiguration()

	// Load unit types for unit restrictions UI (after config is initialized)
	p.loadUnitTypes()

	return nil, false
}

// loadUnitTypes populates the UnitTypes field for the unit restrictions UI
func (p *StartGamePage) loadUnitTypes() {
	// Load unit types with icons from rules engine
	p.UnitTypes = []UnitType{}

	// Get all available unit types from the rules engine
	rulesEngine := weewar.DefaultRulesEngine()

	// Use units from rules engine
	for _, unitID := range AllowedUnitIDs {
		unitData, err := rulesEngine.GetUnitData(unitID)
		if unitData != nil && err == nil {
			// Use web-accessible static URL path for the unit asset
			iconDataURL := fmt.Sprintf("/static/assets/v1/Units/%d/0.png", unitID)

			p.UnitTypes = append(p.UnitTypes, UnitType{
				UnitDefinition: unitData,
				IconDataURL:    iconDataURL,
			})
		}
	}
}

// initializeGameConfiguration sets up default game configuration values
func (p *StartGamePage) initializeGameConfiguration() {
	playerColors := []string{"red", "blue", "green", "yellow", "purple", "orange"}

	// Initialize players with defaults
	players := []*protos.GamePlayer{}
	for i := range 2 {
		playerType := "ai"
		if i == 0 {
			playerType = "human"
		}
		players = append(players, &protos.GamePlayer{
			PlayerId:      int32(i + 1),
			PlayerType:    playerType,
			Color:         playerColors[i%len(playerColors)],
			TeamId:        int32(i + 1),
			Name:          fmt.Sprintf("Player %d", i+1),
			IsActive:      true,
			StartingCoins: weewar.DefaultStartingCoins,
			Coins:         weewar.DefaultStartingCoins,
		})
	}

	// Initialize default income configuration
	incomeConfig := &protos.IncomeConfig{
		LandbaseIncome:    weewar.DefaultLandbaseIncome,
		NavalbaseIncome:   weewar.DefaultNavalbaseIncome,
		AirportbaseIncome: weewar.DefaultAirportbaseIncome,
		MissilesiloIncome: weewar.DefaultMissilesiloIncome,
		MinesIncome:       weewar.DefaultMinesIncome,
	}

	// Initialize default settings
	settings := &protos.GameSettings{
		AllowedUnits:  AllowedUnitIDs,
		TurnTimeLimit: 0,
		TeamMode:      "ffa",
		MaxTurns:      0,
	}

	p.GameConfiguration = &protos.GameConfiguration{
		Players:       players,
		Teams:         []*protos.GameTeam{},
		IncomeConfigs: incomeConfig,
		Settings:      settings,
	}
}

func (p *StartGamePage) Copy() View {
	return &StartGamePage{}
}
