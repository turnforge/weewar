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

	// Username fields - for login alias (stored in UsernameStore)
	UsernameNeeded bool // True if user hasn't set a username yet

	// Nickname fields (Nickname is in SampleProfilePage base)
	NicknameNeeded    bool   // True if user needs to set their nickname
	SuggestedNickname string // Random suggestion for nickname

	// Password availability
	HasLocalAuth bool // True if user has an email identity (can potentially have password)
	HasPassword  bool // True if user has a local channel with password set
}

func (p *ProfilePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	// Handle POST requests for profile updates
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

		// Check if username/nickname need to be set (for highlighting)
		p.UsernameNeeded = p.Username == ""
		p.NicknameNeeded = p.Nickname == ""
		// Always provide a suggested nickname for the placeholder
		p.SuggestedNickname = GenerateRandomNickname()

		if p.Email != "" {
			identity, _, identityErr := ctx.AuthService.GetIdentity("email", p.Email, false)
			if identityErr == nil && identity != nil {
				p.EmailVerified = identity.Verified
				// User has local auth capability if they have an email identity
				p.HasLocalAuth = true

				// Check if user has a local channel (password set)
				identityKey := oa.IdentityKey("email", p.Email)
				channel, _, _ := ctx.AuthService.GetChannel("local", identityKey, false)
				p.HasPassword = channel != nil
			}
		}
	}

	return
}

// handlePost handles POST requests for profile updates (nickname and username)
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
		Username string `json:"username"`
		Action   string `json:"action"` // "nickname" or "username"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return nil, true
	}

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

	profile := user.Profile()

	// Handle username update
	if req.Action == "username" {
		return p.handleUsernameUpdate(ctx, userID, req.Username, user, profile, w)
	}

	// Default: handle nickname update
	return p.handleNicknameUpdate(ctx, userID, req.Nickname, user, profile, w)
}

// handleNicknameUpdate updates the user's nickname (display name)
func (p *ProfilePage) handleNicknameUpdate(ctx *LilBattleApp, userID, nickname string, user oa.User, profile map[string]any, w http.ResponseWriter) (error, bool) {
	nickname = strings.TrimSpace(nickname)
	if len(nickname) < 2 || len(nickname) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nickname must be between 2 and 30 characters"})
		return nil, true
	}

	profile["nickname"] = nickname

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

// handleUsernameUpdate updates the user's username (login alias)
// Username is stored both in profile and UsernameStore for login lookup
func (p *ProfilePage) handleUsernameUpdate(ctx *LilBattleApp, userID, newUsername string, user oa.User, profile map[string]any, w http.ResponseWriter) (error, bool) {
	newUsername = strings.TrimSpace(strings.ToLower(newUsername))

	// Validate username format (alphanumeric, underscores, hyphens, 3-20 chars)
	if len(newUsername) < 3 || len(newUsername) > 20 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username must be between 3 and 20 characters"})
		return nil, true
	}

	// Simple validation: alphanumeric, underscores, hyphens only
	for _, c := range newUsername {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Username can only contain letters, numbers, underscores, and hyphens"})
			return nil, true
		}
	}

	// Get current username (if any)
	oldUsername, _ := profile["username"].(string)

	// If username hasn't changed, nothing to do
	if oldUsername == newUsername {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"success": true, "username": newUsername})
		return nil, true
	}

	// Update UsernameStore
	if ctx.UsernameStore != nil {
		if oldUsername != "" {
			// Changing username: use ChangeUsername for atomic operation
			if err := ctx.UsernameStore.ChangeUsername(oldUsername, newUsername, userID); err != nil {
				log.Printf("Error changing username from %s to %s for user %s: %v", oldUsername, newUsername, userID, err)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Username is already taken"})
				return nil, true
			}
		} else {
			// Setting username for first time
			if err := ctx.UsernameStore.ReserveUsername(newUsername, userID); err != nil {
				log.Printf("Error reserving username %s for user %s: %v", newUsername, userID, err)
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Username is already taken"})
				return nil, true
			}
		}
	}

	// Update profile
	profile["username"] = newUsername

	if err := ctx.AuthService.SaveUser(user); err != nil {
		// Try to rollback UsernameStore change
		if ctx.UsernameStore != nil {
			if oldUsername != "" {
				ctx.UsernameStore.ChangeUsername(newUsername, oldUsername, userID)
			} else {
				ctx.UsernameStore.ReleaseUsername(newUsername)
			}
		}
		log.Printf("Error updating username for user %s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save username"})
		return nil, true
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"success": true, "username": newUsername})
	return nil, true
}
