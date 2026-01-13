package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	goal "github.com/panyam/goapplib"
	oa "github.com/panyam/oneauth"
)

// ProfilePage extends goapplib.SampleProfilePage with app-specific features.
type ProfilePage struct {
	goal.SampleProfilePage[*LilBattleApp]
	Header Header

	// App-specific user information
	User oa.User

	// Nickname fields (Nickname is in SampleProfilePage base)
	NicknameNeeded    bool   // True if user needs to set their nickname
	SuggestedNickname string // Random suggestion for nickname
}

func (p *ProfilePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	// Handle POST requests for nickname updates
	if r.Method == http.MethodPost {
		return p.handlePost(r, w, app)
	}

	err, finished = goal.LoadAll(r, w, app, &p.SampleProfilePage, &p.Header)
	if err != nil || finished {
		return
	}

	ctx := app.Context
	p.UserID = ctx.AuthMiddleware.GetLoggedInUserId(r)
	if p.UserID == "" {
		http.Redirect(w, r, "/login?callbackURL=/profile", http.StatusFound)
		return nil, true
	}

	if ctx.AuthService != nil {
		p.User, err = ctx.AuthService.GetUserById(p.UserID)
		if err != nil || p.User == nil {
			http.Redirect(w, r, "/login?callbackURL=/profile", http.StatusFound)
			return nil, true
		}

		p.Profile = p.User.Profile()

		if email, ok := p.Profile["email"].(string); ok {
			p.Email = email
		}
		if username, ok := p.Profile["username"].(string); ok {
			p.Username = username
		}
		if nickname, ok := p.Profile["nickname"].(string); ok {
			p.Nickname = nickname
		}

		// Check if nickname needs to be set (for highlighting)
		p.NicknameNeeded = p.Nickname == ""
		// Always provide a suggested nickname for the placeholder
		p.SuggestedNickname = GenerateRandomNickname()

		if p.Email != "" {
			identity, _, identityErr := ctx.AuthService.GetIdentity("email", p.Email, false)
			if identityErr == nil && identity != nil {
				p.EmailVerified = identity.Verified
			}
		}
	}

	return
}

// handlePost handles POST requests for profile updates (e.g., nickname)
func (p *ProfilePage) handlePost(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	ctx := app.Context
	userID := ctx.AuthMiddleware.GetLoggedInUserId(r)

	w.Header().Set("Content-Type", "application/json")

	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not authenticated"})
		return nil, true
	}

	// Parse request body
	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return nil, true
	}

	// Validate nickname
	nickname := strings.TrimSpace(req.Nickname)
	if len(nickname) < 2 || len(nickname) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nickname must be between 2 and 30 characters"})
		return nil, true
	}

	// Get user and update nickname
	if ctx.AuthService == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Auth service not available"})
		return nil, true
	}

	user, err := ctx.AuthService.GetUserById(userID)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return nil, true
	}

	// Update the nickname in profile
	// Profile() returns a reference to the user's internal profile map
	profile := user.Profile()
	profile["nickname"] = nickname

	// Save the user with updated profile
	if err := ctx.AuthService.SaveUser(user); err != nil {
		log.Printf("Error updating nickname for user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save nickname"})
		return nil, true
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"success": true, "nickname": nickname})
	return nil, true
}
