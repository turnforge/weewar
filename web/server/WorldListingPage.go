package server

import (
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type WorldListingPage struct {
	BasePage
	Header Header

	WorldListView WorldListView
	ListingData   *goal.EntityListingData[*protos.World]
}

func (m *WorldListingPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	m.DisableSplashScreen = true
	m.Title = "Worlds"
	m.ActiveTab = "worlds"
	m.Header.Load(r, w, app)

	// Load worlds via the existing WorldListView
	if err, finished := m.WorldListView.Load(r, w, app); err != nil || finished {
		return err, finished
	}

	// Build listing data for EntityListing template
	m.ListingData = goal.NewEntityListingData[*protos.World]("My Worlds", "/worlds").
		WithCreate("/worlds/new", "Create New World").
		WithView("/worlds/%s/view").
		WithEdit("/worlds/%s/edit").
		WithDelete("/worlds/%s/delete")
	m.ListingData.Items = m.WorldListView.Worlds
	m.ListingData.ViewMode = m.WorldListView.ViewMode
	m.ListingData.EnableViewToggle = true
	m.ListingData.SearchPlaceholder = "Search worlds..."
	m.ListingData.EmptyTitle = "No worlds found"
	m.ListingData.EmptyMessage = "Get started by creating a new world."

	return nil, false
}
