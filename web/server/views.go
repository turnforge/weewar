package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	"github.com/turnforge/weewar/services/fsbe"
)

const TEMPLATES_FOLDER = "./web/templates"
const ServeGzippedResources = true

// You may have a builder/bundler creating an output folder.  Set that path here.  It can be absolute or relative to
// where the executable will be running from
const DIST_FOLDER = "./web/dist"
const STATIC_FOLDER = "./web/static"

// RootViewsHandler is a thin wrapper for page routing.
// It references both the WeewarApp (pure context) and goal.App (templates/rendering).
type RootViewsHandler struct {
	WeewarApp *WeewarApp              // Pure app context
	GoalApp   *goal.App[*WeewarApp]   // goal.App wrapper for templates/rendering
	mux       *http.ServeMux
}

// NewRootViewsHandler creates a new RootViewsHandler.
func NewRootViewsHandler(weewarApp *WeewarApp, goalApp *goal.App[*WeewarApp]) *RootViewsHandler {
	out := &RootViewsHandler{
		WeewarApp: weewarApp,
		GoalApp:   goalApp,
		mux:       http.NewServeMux(),
	}

	// Add any additional template functions specific to views
	goalApp.Templates.AddFuncs(template.FuncMap{
		"UserInfo": func(userId string) map[string]any {
			return map[string]any{
				"FullName":  "XXXX YYY",
				"Name":      "XXXX",
				"AvatarUrl": "/avatar/url",
			}
		},
		"AsHtmlAttribs": func(m map[string]string) template.HTML {
			return `a = 'b' c = 'd'`
		},
		"contains": func(slice []int32, item int32) bool {
			for _, v := range slice {
				if v == item {
					return true
				}
			}
			return false
		},
	})

	out.setupRoutes()
	return out
}

func (b *RootViewsHandler) HandleError(err error, w io.Writer) {
	if err != nil {
		fmt.Fprint(w, "Error rendering: ", err.Error())
	}
}

func (n *RootViewsHandler) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("DEBUG: RootViewsHandler received request: %s %s", r.Method, r.URL.Path)
		n.mux.ServeHTTP(w, r)
	})
}

// setupRoutes sets up all view routes, pages, etc
func (n *RootViewsHandler) setupRoutes() {
	log.Println("DEBUG: Setting up routes...")

	// View fragments for htmx partial responses
	n.mux.Handle("/views/", http.StripPrefix("/views", n.setupViewsMux()))

	// Static files
	if ServeGzippedResources {
		n.mux.Handle("/static/", http.StripPrefix("/static", goal.GzipFileServer(http.Dir(STATIC_FOLDER))))
	} else {
		n.mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(STATIC_FOLDER))))
	}

	// Serve screenshots from filestore
	screenshotsDir := fsbe.FILES_STORAGE_DIR
	if screenshotsDir == "" {
		screenshotsDir = fsbe.DevDataPath("storage/files")
	}
	if ServeGzippedResources {
		n.mux.Handle("/screenshots/", http.StripPrefix("/screenshots", goal.GzipFileServer(http.Dir(screenshotsDir+"/screenshots"))))
	} else {
		n.mux.Handle("/screenshots/", http.StripPrefix("/screenshots", http.FileServer(http.Dir(screenshotsDir+"/screenshots"))))
	}

	// Resource-specific endpoints - pass WeewarApp and GoalApp to groups
	gamesGroup := &GamesGroup{weewarApp: n.WeewarApp, goalApp: n.GoalApp}
	gamesMux := gamesGroup.RegisterRoutes(n.GoalApp)
	n.mux.Handle("/games/", http.StripPrefix("/games", gamesMux))

	worldsGroup := &WorldsGroup{weewarApp: n.WeewarApp, goalApp: n.GoalApp}
	worldsMux := worldsGroup.RegisterRoutes(n.GoalApp)
	n.mux.Handle("/worlds/", http.StripPrefix("/worlds", worldsMux))

	rulesGroup := &RulesGroup{}
	rulesMux := rulesGroup.RegisterRoutes(n.GoalApp)
	n.mux.Handle("/rules/", http.StripPrefix("/rules", rulesMux))

	// Handle no-trailing-slash redirects
	n.mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/games/", http.StatusMovedPermanently)
	})
	n.mux.HandleFunc("/worlds", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/worlds/", http.StatusMovedPermanently)
	})

	// Standalone pages using goal.Register
	goal.Register[*GenericPage](n.GoalApp, n.mux, "/about", goal.WithTemplate("AboutPage"))
	goal.Register[*GenericPage](n.GoalApp, n.mux, "/contact", goal.WithTemplate("ContactUsPage"))
	goal.Register[*LoginPage](n.GoalApp, n.mux, "/login")
	goal.Register[*ProfilePage](n.GoalApp, n.mux, "/profile")
	goal.Register[*PrivacyPolicy](n.GoalApp, n.mux, "/privacy/")
	goal.Register[*TermsOfService](n.GoalApp, n.mux, "/terms/")
	goal.Register[*HomePage](n.GoalApp, n.mux, "/")
	n.mux.Handle("/{invalidbits}/", http.NotFoundHandler())
}

func (n *RootViewsHandler) setupViewsMux() *http.ServeMux {
	mux := http.NewServeMux()
	// Register view fragment handlers here for htmx partial responses
	return mux
}
