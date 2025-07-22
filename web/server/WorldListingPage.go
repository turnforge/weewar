package server

import "net/http"

type WorldListingPage struct {
	BasePage
	Header Header

	// Add any other components here to reflect what you want to show in your home page
	// Note that you would also update your HomePage templates to reflect these
	WorldListView WorldListView
}

func (m *WorldListingPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	m.Title = "Worlds"
	m.Header.Load(r, w, vc)
	err, finished = m.WorldListView.Load(r, w, vc)
	if err != nil || finished {
		return
	}
	return
}

func (m *WorldListingPage) Copy() View {
	return &WorldListingPage{}
}
