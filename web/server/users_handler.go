package server

import (
	"encoding/json"
	"net/http"
)

// UsersHandler handles user showcase pages
type UsersHandler struct {
	App *LilBattleApp
}

// NewUsersHandler creates a new users handler
func NewUsersHandler(app *LilBattleApp) *UsersHandler {
	return &UsersHandler{App: app}
}

// Handler returns an HTTP handler for users routes
func (h *UsersHandler) Handler() http.Handler {
	mux := http.NewServeMux()

	// User listing page
	mux.HandleFunc("/users", h.handleUserListing)
	mux.HandleFunc("/users/", h.handleUserListing)

	// User details page
	mux.HandleFunc("/user/", h.handleUserDetails)

	return mux
}

// handleUserListing renders the user listing page
func (h *UsersHandler) handleUserListing(w http.ResponseWriter, r *http.Request) {
	/*
		// Get all users from catalog
		users := h.catalog.ListUsers()

		// Prepare template data
		data := map[string]any{
			"Title":    "User Examples",
			"PageType": "user-listing",
			"Users": users,
			"PageDataJSON": toJSON(map[string]any{
				"pageType": "user-listing",
			}),
		}

		// Load and render template
		templates := h.templateGroup.MustLoad("users/listing.html", "")

		// Render the template
		if err := h.templateGroup.RenderHtmlTemplate(w, templates[0], "", data, nil); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render page: %v", err), http.StatusInternalServerError)
			return
		}
	*/
}

// handleUserDetails renders the user details page
func (h *UsersHandler) handleUserDetails(w http.ResponseWriter, r *http.Request) {
	/*
		// Extract user ID from path
		// Path format: /user/bitly
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		userID := parts[2]

		// Get user from catalog
		user := h.catalog.GetUser(userID)
		if user == nil {
			http.NotFound(w, r)
			return
		}

		// Get mode from query params (default to server mode)
		mode := "server"
		if r.URL.Query().Get("mode") == "wasm" {
			mode = "wasm"
		}

		// Get version (default to user's default version)
		version := r.URL.Query().Get("version")
		if version == "" {
			version = user.DefaultVersion
		}

		// Get SDL and recipe content for the version
		versionData := user.Versions[version]

		// Prepare minimal page data for the client (content will be loaded via API)
		pageData := map[string]any{
			"userId": user.ID,
			"mode":      mode,
		}

		// Prepare template data
		data := map[string]any{
			"Title":        user.Name + " - SDL User",
			"PageType":     "user-details",
			"User":      user,
			"Mode":         mode,
			"PageDataJSON": toJSON(pageData),
		}

		// Load and render template
		templates := h.templateGroup.MustLoad("users/details.html", "")

		// Render the template
		if err := h.templateGroup.RenderHtmlTemplate(w, templates[0], "", data, nil); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render page: %v", err), http.StatusInternalServerError)
			return
		}
	*/
}

// toJSON converts data to JSON string for template use
func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
