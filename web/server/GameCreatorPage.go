package server

import (
	"fmt"
	"net/http"

	protos "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type GameCreatorPage struct {
	Header        Header
	Game          *protos.Game
	Errors        map[string]string
	AllowCustomId bool
}

func (g *GameCreatorPage) Copy() View { return &GameCreatorPage{} }

func (v *GameCreatorPage) SetupDefaults() {
	v.Header.Width = "w-full"
	v.Header.PageData = v
	v.Header.FixedHeader = true
	v.Header.ShowHomeButton = true
	v.Header.ShowLogoutButton = false
	v.Header.ShowComposeButton = false
	v.Header.Styles = map[string]any{
		"FixedHeightHeader":          true,
		"HeaderHeightIfFixed":        "70px",
		"MinWidthForFullWidthMenu":   "24em",
		"MinWidthForHamburgerMenu":   "48em",
		"MinWidthForCompressingLogo": "24em",

		"EditorHeaderHeight":  "50px",
		"ToolbarButtonHeight": "30px",
		"HeaderHeight":        "90px",
		"StatusBarHeight":     "30px",
		"HeaderBarHeight":     "90px",
	}
}

func (v *GameCreatorPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.Header.Load(r, w, vc)
	v.SetupDefaults()
	loggedInUserId := vc.AuthMiddleware.GetLoggedInUserId(r)

	if loggedInUserId == "" {
		// For now enforce login even on new
		qs := r.URL.RawQuery
		if len(qs) > 0 {
			qs = "?" + qs
		}
		http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/notations/new%s", qs)), http.StatusSeeOther)
		return nil, true
	}
	v.Game = &protos.Game{}
	if v.Game.Name == "" {
		v.Game.Name = "Untitled Game"
	}
	return
}
