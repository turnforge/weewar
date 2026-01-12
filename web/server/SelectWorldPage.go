package server

import (
	"fmt"
	"net/http"

	goal "github.com/panyam/goapplib"
)

type SelectWorldPage struct {
	BasePage
	Header        Header
	WorldListView WorldListView
}

func (m *SelectWorldPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	// Require login to select a world for game creation
	ctx := app.Context
	loggedInUserId := ctx.AuthMiddleware.GetLoggedInUserId(r)
	if loggedInUserId == "" {
		qs := r.URL.RawQuery
		if len(qs) > 0 {
			qs = "?" + qs
		}
		http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/worlds/select%s", qs)), http.StatusSeeOther)
		return nil, true
	}

	m.Title = "Select a World"
	m.DisableSplashScreen = true
	m.Header.Load(r, w, app)
	m.WorldListView.ActionMode = "select"
	return m.WorldListView.Load(r, w, app)
}
