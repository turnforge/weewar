package server

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	goal "github.com/panyam/goapplib"
	goalservices "github.com/panyam/goapplib/services"
	gotl "github.com/panyam/goutils/template"
	oa "github.com/panyam/oneauth"
	"github.com/turnforge/weewar/services"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type BasePage struct {
	goal.BasePage
}

// WeewarApp is the pure application context.
// It holds all app-specific state without knowing about goapplib.
// Views access this via app.Context in goal.App[*WeewarApp].
type WeewarApp struct {
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

	mux     *http.ServeMux
	BaseUrl string
}

// NewWeewarApp creates a new WeewarApp and its associated goal.App.
// Returns the WeewarApp and the goal.App wrapper.
func NewWeewarApp(clientMgr *services.ClientMgr) (weewarApp *WeewarApp, goalApp *goal.App[*WeewarApp], err error) {
	session := scs.New()
	authService, oneauth := setupAuthService(session)

	// Create WeewarApp (pure app context)
	weewarApp = &WeewarApp{
		Auth:           oneauth,
		AuthMiddleware: &oneauth.Middleware,
		AuthService:    authService,
		Session:        session,
		ClientMgr:      clientMgr,
		HideGames:      os.Getenv("WEEWAR_HIDE_GAMES") == "true",
		HideWorlds:     os.Getenv("WEEWAR_HIDE_WORLDS") == "true",
	}

	// Setup templates with app-specific FuncMap additions
	templates := goal.SetupTemplates(TEMPLATES_FOLDER)
	// Add goutils template functions (Ago, etc.)
	templates.AddFuncs(gotl.DefaultFuncMap())
	templates.AddFuncs(template.FuncMap{
		// Ctx provides access to the WeewarApp context in templates
		"Ctx": func() *WeewarApp { return weewarApp },
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
	goalApp = goal.NewApp(weewarApp, templates)

	// Initialize API
	api := &ApiHandler{AuthMiddleware: &oneauth.Middleware, ClientMgr: clientMgr}
	if err := api.Init(); err != nil {
		return nil, nil, err
	}
	weewarApp.Api = api

	// Create ViewsRoot (now just a thin wrapper referencing weewarApp and goalApp)
	weewarApp.ViewsRoot = NewRootViewsHandler(weewarApp, goalApp)

	return
}

// Handler returns a configured HTTP handler with all routes.
func (a *WeewarApp) Handler() http.Handler {
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
