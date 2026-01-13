package server

import (
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type GameListingPage struct {
	BasePage
	Header Header

	GameListView GameListView
	ListingData  *goal.EntityListingData[*protos.Game]
}

func (m *GameListingPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	m.Title = "Games"
	m.ActiveTab = "games"
	m.DisableSplashScreen = true
	m.Header.Load(r, w, app)

	// Load games via the existing GameListView
	if err, finished := m.GameListView.Load(r, w, app); err != nil || finished {
		return err, finished
	}

	// Build listing data for EntityListing template
	m.ListingData = goal.NewEntityListingData[*protos.Game]("My Games", "/games").
		WithCreate("/games/new", "Start New Game").
		WithView("/games/%s/view").
		WithEdit("/games/%s/edit").
		WithDelete("/games/%s/delete")
	m.ListingData.Items = m.GameListView.Games
	m.ListingData.ViewMode = m.GameListView.ViewMode
	m.ListingData.EnableViewToggle = true
	m.ListingData.SearchPlaceholder = "Search games..."
	m.ListingData.EmptyTitle = "No games yet?"
	m.ListingData.EmptyMessage = "Get started on your first game."

	return nil, false
}
