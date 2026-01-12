package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
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

	// Form fields (can be pre-filled or from error retry)
	GameId       string
	GameName     string
	ErrorMessage string // Error message for ID conflicts etc.
}

func (p *StartGamePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	// Require login to start a game
	ctx := app.Context
	loggedInUserId := ctx.AuthMiddleware.GetLoggedInUserId(r)
	if loggedInUserId == "" {
		qs := r.URL.RawQuery
		if len(qs) > 0 {
			qs = "?" + qs
		}
		http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/games/start%s", qs)), http.StatusSeeOther)
		return nil, true
	}

	// Get worldId from query parameter (optional)
	p.WorldId = r.URL.Query().Get("worldId")

	// If no worldId provided, redirect to world selection page
	if p.WorldId == "" {
		http.Redirect(w, r, "/worlds/select", http.StatusSeeOther)
		return nil, true
	}

	// Read optional gameId, gameName, and error from query params
	p.GameId = r.URL.Query().Get("gameId")
	p.GameName = r.URL.Query().Get("gameName")
	p.ErrorMessage = r.URL.Query().Get("error")
	if p.GameName == "" {
		p.GameName = "New Game"
	}

	p.Title = p.GameName
	p.Header.Load(r, w, app)

	// If a worldId is provided, fetch the world data
	if p.WorldId != "" {
		// Fetch the World using the client manager
		ctx := app.Context
		client := ctx.ClientMgr.GetWorldsSvcClient()
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
	rulesEngine := lib.DefaultRulesEngine()

	// Use units from rules engine
	for _, unitID := range AllowedUnitIDs {
		unitData, err := rulesEngine.GetUnitData(unitID)
		if unitData != nil && err == nil {
			// Use web-accessible static URL path for the unit asset
			iconDataURL := fmt.Sprintf("/static/assets/themes/default/Units/%d/0.png", unitID)

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

	// Initialize default income configuration first (so we can use StartingCoins for players)
	incomeConfig := &protos.IncomeConfig{
		StartingCoins:     lib.DefaultStartingCoins,
		LandbaseIncome:    lib.DefaultLandbaseIncome,
		NavalbaseIncome:   lib.DefaultNavalbaseIncome,
		AirportbaseIncome: lib.DefaultAirportbaseIncome,
		MissilesiloIncome: lib.DefaultMissilesiloIncome,
		MinesIncome:       lib.DefaultMinesIncome,
	}

	// Initialize players with defaults using IncomeConfig.StartingCoins
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
			StartingCoins: incomeConfig.StartingCoins,
		})
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
