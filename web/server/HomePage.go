package server

import (
	"net/http"

	goal "github.com/panyam/goapplib"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type GenericPage struct {
	BasePage
	Header Header
}

func (v *GenericPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	return
}

type HomePage struct {
	BasePage
	Header Header

	// Dashboard data
	RecentGames  []*v1.Game
	RecentWorlds []*v1.World
	TotalGames   int32
	TotalWorlds  int32
}

func (p *HomePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	ctx := app.Context
	// Redirect to the first visible tab
	if !ctx.HideWorlds {
		http.Redirect(w, r, "/worlds/", http.StatusFound)
		return nil, true
	} else if !ctx.HideGames {
		http.Redirect(w, r, "/games/", http.StatusFound)
		return nil, true
	} else {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return nil, true
	}
}

type PrivacyPolicy struct {
	BasePage
	Header Header
}

func (p *PrivacyPolicy) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	p.DisableSplashScreen = true
	return p.Header.Load(r, w, app)
}

type TermsOfService struct {
	BasePage
	Header Header
}

func (t *TermsOfService) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	t.DisableSplashScreen = true
	return t.Header.Load(r, w, app)
}
