package server

import "net/http"

type GameListingPage struct {
	BasePage
	Header Header

	// Add any other components here to reflect what you want to show in your home page
	// Note that you would also update your HomePage templates to reflect these
	GameListView GameListView
}

func (m *GameListingPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	m.Title = "Games"
	m.Header.Load(r, w, vc)
	err, finished = m.GameListView.Load(r, w, vc)
	if err != nil || finished {
		return
	}
	return
}

func (m *GameListingPage) Copy() View {
	return &GameDetailPage{}
}
