package server

import (
	"log"
	"net/http"
	"net/url"

	goal "github.com/panyam/goapplib"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NotFoundPage is the view data for 404 pages
type NotFoundPage struct {
	BasePage
	Header Header
	Path   string
}

func (p *NotFoundPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (error, bool) {
	p.Title = "Not Found"
	p.Header.Load(r, w, app)
	p.Path = r.URL.Path
	return nil, false
}

// ForbiddenPage is the view data for 403 pages
type ForbiddenPage struct {
	BasePage
	Header  Header
	Message string
}

func (p *ForbiddenPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (error, bool) {
	p.Title = "Forbidden"
	p.Header.Load(r, w, app)
	return nil, false
}

// HandleGRPCError checks if err is a gRPC error and handles it appropriately.
// For Unauthenticated errors, it redirects to /login with a callback URL.
// For NotFound errors, it renders the 404 page.
// For PermissionDenied errors, it renders the 403 page.
// Returns (error, finished) - if finished is true, the response has been written.
func HandleGRPCError(err error, w http.ResponseWriter, r *http.Request, app *goal.App[*LilBattleApp]) (error, bool) {
	if err == nil {
		return nil, false
	}

	// Check if it's a gRPC status error
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, return as-is
		return err, false
	}

	switch st.Code() {
	case codes.Unauthenticated:
		// Redirect to login with callback URL
		callbackURL := r.URL.RequestURI()
		loginURL := "/login?callbackURL=" + url.QueryEscape(callbackURL)
		http.Redirect(w, r, loginURL, http.StatusFound)
		return nil, true

	case codes.PermissionDenied:
		// Render 403 page
		page := &ForbiddenPage{Message: st.Message()}
		page.Load(r, w, app)
		w.WriteHeader(http.StatusForbidden)
		if renderErr := app.RenderTemplate(w, "ForbiddenPage", "ForbiddenPage", page); renderErr != nil {
			log.Printf("Error rendering 403 page: %v", renderErr)
			http.Error(w, "Forbidden: "+st.Message(), http.StatusForbidden)
		}
		return nil, true

	case codes.NotFound:
		// Render 404 page
		page := &NotFoundPage{}
		page.Load(r, w, app)
		w.WriteHeader(http.StatusNotFound)
		if renderErr := app.RenderTemplate(w, "NotFoundPage", "NotFoundPage", page); renderErr != nil {
			log.Printf("Error rendering 404 page: %v", renderErr)
			http.Error(w, "Not Found", http.StatusNotFound)
		}
		return nil, true

	default:
		// Return the original error
		return err, false
	}
}

func ThemeFromRequest(req *http.Request) string {
	// Theme fallback priority: URL param > cookie > default
	queryParams := req.URL.Query()
	theme := queryParams.Get("theme")
	if theme == "" {
		if cookie, err := req.Cookie("assetTheme"); err == nil {
			theme = cookie.Value
		}
	}
	if theme == "" {
		theme = "fantasy" // Default theme
	}
	return theme
}
