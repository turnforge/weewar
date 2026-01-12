package server

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexedwards/scs/v2"
	goal "github.com/panyam/goapplib"
	goalservices "github.com/panyam/goapplib/services"
	gotl "github.com/panyam/goutils/template"
	oa "github.com/panyam/oneauth"
	tmplr "github.com/panyam/templar"
	"github.com/turnforge/lilbattle/services"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type BasePage struct {
	goal.BasePage
}

// LilBattleApp is the pure application context.
// It holds all app-specific state without knowing about goapplib.
// Views access this via app.Context in goal.App[*LilBattleApp].
type LilBattleApp struct {
	// Auth
	Auth           *oa.OneAuth
	AuthMiddleware *oa.Middleware
	AuthService    *goalservices.AuthService
	Session        *scs.SessionManager

	// Services
	Api       *ApiHandler
	ClientMgr *services.ClientMgr

	// Views (thin wrapper for page routing)
	ViewsRoot *RootViewsHandler

	// App config
	HideGames  bool
	HideWorlds bool

	// Ad config - all default to enabled, can be disabled per-placement
	AdsEnabled        bool   // Master switch: WEEWAR_ADS_ENABLED (default: true)
	AdsFooterEnabled  bool   // Footer banner: WEEWAR_ADS_FOOTER (default: true)
	AdsHomeEnabled    bool   // Homepage mid-section: WEEWAR_ADS_HOME (default: true)
	AdsListingEnabled bool   // Listing pages: WEEWAR_ADS_LISTING (default: true)
	AdNetworkId       string // Google AdSense publisher ID: WEEWAR_AD_NETWORK_ID

	mux     *http.ServeMux
	BaseUrl string
}

// NewLilBattleApp creates a new LilBattleApp and its associated goal.App.
// Returns the LilBattleApp and the goal.App wrapper.
func NewLilBattleApp(clientMgr *services.ClientMgr) (lilbattleApp *LilBattleApp, goalApp *goal.App[*LilBattleApp], err error) {
	session := scs.New()
	authService, oneauth := setupAuthService(session)

	// Create LilBattleApp (pure app context)
	lilbattleApp = &LilBattleApp{
		Auth:           oneauth,
		AuthMiddleware: &oneauth.Middleware,
		AuthService:    authService,
		Session:        session,
		ClientMgr:      clientMgr,
		HideGames:      os.Getenv("LILBATTLE_HIDE_GAMES") == "true",
		HideWorlds:     os.Getenv("LILBATTLE_HIDE_WORLDS") == "true",
		// Ads default to enabled, can be disabled per-placement
		AdsEnabled:        os.Getenv("LILBATTLE_ADS_ENABLED") != "false",
		AdsFooterEnabled:  os.Getenv("LILBATTLE_ADS_FOOTER") != "false",
		AdsHomeEnabled:    os.Getenv("LILBATTLE_ADS_HOME") != "false",
		AdsListingEnabled: os.Getenv("LILBATTLE_ADS_LISTING") != "false",
		AdNetworkId:       os.Getenv("LILBATTLE_AD_NETWORK_ID"),
	}

	// Setup templates with SourceLoader for @goapplib/ vendored dependencies
	templates := tmplr.NewTemplateGroup()
	configPath := filepath.Join(TEMPLATES_FOLDER, "templar.yaml")
	sourceLoader, err := tmplr.NewSourceLoaderFromConfig(configPath)
	if err != nil {
		log.Printf("Warning: Could not load templar.yaml: %v. Falling back to basic loader.", err)
		// Fall back to basic file system loader
		templates.Loader = tmplr.NewFileSystemLoader(TEMPLATES_FOLDER)
	} else {
		templates.Loader = sourceLoader
	}
	templates.AddFuncs(goal.DefaultFuncMap())
	// Add goutils template functions (Ago, etc.)
	templates.AddFuncs(gotl.DefaultFuncMap())
	templates.AddFuncs(template.FuncMap{
		// Ctx provides access to the LilBattleApp context in templates
		"Ctx": func() *LilBattleApp { return lilbattleApp },
		// Protobuf-aware ToJson (overrides the generic one from goapplib)
		"ToJson": func(v any) template.JS {
			if v == nil {
				return template.JS("null")
			}
			if msg, ok := v.(proto.Message); ok {
				marshaler := protojson.MarshalOptions{
					UseEnumNumbers: true,
				}
				jsonBytes, err := marshaler.Marshal(msg)
				if err == nil {
					return template.JS(jsonBytes)
				}
				log.Printf("Error marshaling protobuf to JSON: %v", err)
			}
			// Fall back to generic ToJson from goapplib
			return goal.DefaultFuncMap()["ToJson"].(func(any) template.JS)(v)
		},
	})

	// Create goal.App wrapper
	goalApp = goal.NewApp(lilbattleApp, templates)

	// Initialize API
	api := &ApiHandler{AuthMiddleware: &oneauth.Middleware, ClientMgr: clientMgr}
	if err := api.Init(); err != nil {
		return nil, nil, err
	}
	lilbattleApp.Api = api

	// Create ViewsRoot (now just a thin wrapper referencing lilbattleApp and goalApp)
	lilbattleApp.ViewsRoot = NewRootViewsHandler(lilbattleApp, goalApp)

	return
}

// Handler returns a configured HTTP handler with all routes.
func (a *LilBattleApp) Handler() http.Handler {
	r := http.NewServeMux()

	// Rate limiting middleware (from goapplib)
	rateLimiter := goal.NewRateLimitMiddleware(goal.DefaultRateLimitConfig())

	// Security headers middleware
	securityHeaders := NewSecurityHeadersMiddleware()

	// Auth routes (with stricter rate limiting)
	r.Handle("/auth/", rateLimiter.WrapAuth(http.StripPrefix("/auth", a.Auth.Handler())))

	// API routes (with API rate limiting)
	r.Handle("/api/", rateLimiter.WrapAPI(http.StripPrefix("/api", a.Api.Handler())))

	// Serve examples directory for WASM demos
	r.Handle("/examples/", http.StripPrefix("/examples", http.FileServer(http.Dir("./examples/"))))

	r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.ViewsRoot.Handler().ServeHTTP(w, r)
	}))

	sessionHandler := a.Session.LoadAndSave(r)

	// Wrap with security headers (outermost middleware)
	return securityHeaders.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionHandler.ServeHTTP(w, r)
	}))
}
