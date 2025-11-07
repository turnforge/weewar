package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexedwards/scs/v2"
	oa "github.com/panyam/oneauth"
	oa2 "github.com/panyam/oneauth/oauth2"
	"github.com/panyam/templar"
	"github.com/panyam/turnengine/games/weewar/services"
	"github.com/panyam/turnengine/games/weewar/services/server"
)

// You can all this anything - but App is just a convention for all "top level" routes and handlers
type App struct {
	Api         *ApiHandler
	Auth        *oa.OneAuth         // One auth gives us alots of things out of the box
	AuthService *services.AuthService
	Session     *scs.SessionManager // Session and auth go together
	ClientMgr   *server.ClientMgr

	// Instead of giving each resource its own dedicated handler we are having a top level "Views"
	// handler.  This is responsible for handling all views/pages/static resources.  In the Views
	// you'd setup the various routes for your project.  The idea is with Views router we can start
	// bundling common "View COntext" related items from a single place
	ViewsRoot *RootViewsHandler

	mux     *http.ServeMux
	BaseUrl string
}

func NewApp(ClientMgr *server.ClientMgr) (app *App, err error) {
	session := scs.New()
	// session.Store = NewMemoryStore(0)

	// Initialize authentication
	storagePath := os.Getenv("WEEWAR_USER_STORAGE_PATH")
	if storagePath == "" {
		storagePath = filepath.Join(os.Getenv("HOME"), "dev-app-data", "weewar", "storage")
	}
	authService := services.NewAuthService(storagePath)

	oneauth := oa.New("weewar")
	oneauth.Session = session
	oneauth.Middleware.SessionGetter = func(r *http.Request, key string) any {
		return session.GetString(r.Context(), key)
	}
	oneauth.UserStore = authService

	// OAuth providers
	oneauth.AddAuth("/google", oa2.NewGoogleOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/github", oa2.NewGithubOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())

	// Get base URL for verification/reset links
	baseURL := os.Getenv("WEEWAR_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Local authentication
	localAuth := &oa.LocalAuth{
		ValidateCredentials:      authService.ValidateLocalCredentials,
		CreateUser:               authService.CreateLocalUser,
		ValidateSignup:           nil, // Use default validator
		EmailSender:              &oa.ConsoleEmailSender{},
		TokenStore:               authService.TokenStore,
		BaseURL:                  baseURL,
		RequireEmailVerification: false, // Optional verification
		UsernameField:            "email", // For login: accept email as username
		HandleUser:               oneauth.SaveUserAndRedirect,
		VerifyEmail:              authService.VerifyEmailByToken,
		UpdatePassword:           authService.UpdatePassword,
	}

	oneauth.AddAuth("/login", localAuth)
	oneauth.AddAuth("/signup", http.HandlerFunc(localAuth.HandleSignup))
	oneauth.AddAuth("/verify-email", http.HandlerFunc(localAuth.HandleVerifyEmail))
	oneauth.AddAuth("/forgot-password", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			localAuth.HandleForgotPasswordForm(w, r)
		} else {
			localAuth.HandleForgotPassword(w, r)
		}
	}))
	oneauth.AddAuth("/reset-password", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			localAuth.HandleResetPasswordForm(w, r)
		} else {
			localAuth.HandleResetPassword(w, r)
		}
	}))

	app = &App{
		Session:     session,
		Auth:        oneauth,
		AuthService: authService,
		Api:         NewApiHandler(&oneauth.Middleware, ClientMgr),
		ViewsRoot:   NewRootViewsHandler(&oneauth.Middleware, ClientMgr),
	}

	return
}

// GetRouter returns a configured HTTP router with all Canvas API routes
func (a *App) Handler() http.Handler {
	r := http.NewServeMux()

	// here is where we go and setup all the routes for the various prefixes

	// Auth routes
	r.Handle("/auth/", http.StripPrefix("/auth", a.Auth.Handler()))

	// API routes
	r.Handle("/api/", http.StripPrefix("/api", a.Api.Handler()))

	// Fileappitem API endpoints
	// Serve examples directory for WASM demos
	r.Handle("/examples/", http.StripPrefix("/examples", http.FileServer(http.Dir("./examples/"))))

	r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("DEBUG: App handler received request: %s %s", r.Method, r.URL.Path)
		a.ViewsRoot.Handler().ServeHTTP(w, r)
	}))

	sessionHandler := a.Session.LoadAndSave(r)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("DEBUG: Session middleware handling request: %s %s", r.Method, r.URL.Path)
		sessionHandler.ServeHTTP(w, r)
	})
}

// SetupTemplates initializes the Templar template group
func SetupTemplates(templatesDir string) (*templar.TemplateGroup, error) {
	// Create a new template group
	group := templar.NewTemplateGroup()

	// Set up the file appitem loader with multiple paths
	group.Loader = templar.NewFileSystemLoader(
		templatesDir,
		templatesDir+"/shared",
		templatesDir+"/components",
	)

	// Preload common templates to ensure they're available
	commonTemplates := []string{
		"base.html",
		"appitems/listing.html",
		"appitems/details.html",
	}

	for _, tmpl := range commonTemplates {
		// Use defer to catch panics from MustLoad
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Template not found (will create): %s", tmpl)
				}
			}()
			group.MustLoad(tmpl, "")
		}()
	}

	return group, nil
}
