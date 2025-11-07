package server

import (
	"html/template"
	"net/http"
)

type HeaderMenuItem struct {
	Id            string
	Title         string
	Link          string
	Class         string
	LowWidthClass string
	Raw           string
	OnClick       string
	Attributes    map[string]string
}

type Header struct {
	AppName             string
	PageData            any
	HeaderLogo          func() template.HTML
	LogoTitle           string
	Styles              map[string]any
	Width               string
	HeaderStyleLink     string
	MenuStyleLink       string
	FixedHeader         bool
	ShowHomeButton      bool
	ShowComposeButton   bool
	ShowLoginButton     bool
	ShowLogoutButton    bool
	HomeButtonImage     string
	CenterMenuItems     []HeaderMenuItem
	SpecialMenuItems    []HeaderMenuItem
	RightMenuItems      []HeaderMenuItem
	HideCenterMenuItems bool
	IsLoggedIn          bool
	LoggedInUserId      string
	Username            string
}

func (h *Header) SetupDefaults() {
	h.AppName = "Notations"
	h.HeaderStyleLink = "/static/css/Header.css"
	h.MenuStyleLink = "/static/css/Menu.css"
	h.Width = "max-w-7xl"
	h.ShowHomeButton = false
	h.ShowComposeButton = true
	h.ShowLoginButton = true
	h.ShowLogoutButton = true
	h.HomeButtonImage = "/static/icons/homebutton.jpg"
	h.LogoTitle = "Notations"
	h.HeaderStyleLink = "/static/css/Header.css"
	h.MenuStyleLink = "/static/css/Menu.css"
	h.HideCenterMenuItems = false
	h.SpecialMenuItems = []HeaderMenuItem{
		{Title: "Compose", Link: "/compose"},
	}
	h.RightMenuItems = []HeaderMenuItem{
		{Title: "My Music", Link: "/mymusic"},
		// {Title: "Tutorial", Link: "/tutorial"},
	}
	h.Styles = map[string]any{
		"FixedHeightHeader":          false,
		"HeaderHeightIfFixed":        "80px",
		"MinWidthForFullWidthMenu":   "24em",
		"MinWidthForHamburgerMenu":   "48em",
		"MinWidthForCompressingLogo": "24em",
	}
}

func (v *Header) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	v.SetupDefaults()
	v.LoggedInUserId = vc.AuthMiddleware.GetLoggedInUserId(r)
	v.IsLoggedIn = v.LoggedInUserId != ""

	// Load username if logged in
	if v.IsLoggedIn && vc.AuthService != nil {
		user, userErr := vc.AuthService.GetUserById(v.LoggedInUserId)
		if userErr == nil && user != nil {
			profile := user.Profile()
			if username, ok := profile["username"].(string); ok {
				v.Username = username
			}
		}
	}

	return
}
