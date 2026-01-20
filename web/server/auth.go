package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	goalservices "github.com/panyam/goapplib/services"
	oa "github.com/panyam/oneauth"
	oa2 "github.com/panyam/oneauth/oauth2"
	oafs "github.com/panyam/oneauth/stores/fs"
)

func setupAuthService(session *scs.SessionManager) (*goalservices.AuthService, oa.UsernameStore, *oa.OneAuth) {
	// Initialize authentication
	storagePath := os.Getenv("LILBATTLE_USER_STORAGE_PATH")
	if storagePath == "" {
		storagePath = filepath.Join(os.Getenv("HOME"), "dev-app-data", "lilbattle", "storage")
	}
	authService := goalservices.NewAuthService(storagePath)

	// Create UsernameStore for username → userID mapping
	usernameStore := oafs.NewFSUsernameStore(storagePath)

	oneauth := oa.New("lilbattle")
	oneauth.Session = session
	oneauth.Middleware.SessionGetter = func(r *http.Request, key string) any {
		return session.GetString(r.Context(), key)
	}
	oneauth.UserStore = authService

	// OAuth providers - credentials loaded from environment
	oneauth.AddAuth("/google", oa2.NewGoogleOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/github", oa2.NewGithubOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/twitter", NewTwitterOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())

	// Get base URL for verification/reset links
	baseURL := os.Getenv("LILBATTLE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Create credentials validator that supports email OR username login
	// - If input contains "@", treats as email
	// - Otherwise, looks up username in UsernameStore to find userID
	validateCredentials := oa.NewCredentialsValidatorWithUsername(
		authService.IdentityStore,
		authService.ChannelStore,
		authService.UserStore,
		usernameStore,
	)

	// Local authentication (username/password)
	localAuth := &oa.LocalAuth{
		ValidateCredentials:      validateCredentials,
		CreateUser:               authService.CreateLocalUser,
		ValidateSignup:           nil, // Policy handles validation now
		EmailSender:              &oa.ConsoleEmailSender{},
		TokenStore:               authService.TokenStore,
		BaseURL:                  baseURL,
		RequireEmailVerification: false,   // Optional verification
		UsernameField:            "email", // Form field name (auto-detection happens after parsing)
		HandleUser:               oneauth.SaveUserAndRedirect,
		VerifyEmail:              authService.VerifyEmailByToken,
		UpdatePassword:           authService.UpdatePassword,
		UsernameStore:            usernameStore,

		// Signup policy: email required, username NOT collected at signup
		SignupPolicy: &oa.SignupPolicy{
			RequireUsername:       false, // Username added later via profile
			RequireEmail:          true,
			RequirePassword:       true,
			EnforceUsernameUnique: false, // Not enforcing at signup since not collected
			EnforceEmailUnique:    true,
			MinPasswordLength:     8,
		},

		// URLs for redirect-based error handling
		LoginURL:  "/login",
		SignupURL: "/login", // Same page, different tab

		// Redirect-based error handling with flash messages
		OnSignupError: func(err *oa.AuthError, w http.ResponseWriter, r *http.Request) bool {
			session.Put(r.Context(), "auth_error", err.Message)
			session.Put(r.Context(), "auth_error_field", err.Field)
			session.Put(r.Context(), "auth_mode", "signup")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return true
		},
		OnLoginError: func(err *oa.AuthError, w http.ResponseWriter, r *http.Request) bool {
			session.Put(r.Context(), "auth_error", err.Message)
			session.Put(r.Context(), "auth_error_field", err.Field)
			session.Put(r.Context(), "auth_mode", "login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return true
		},
	}

	oneauth.AddAuth("/login", localAuth)
	oneauth.AddAuth("/signup", http.HandlerFunc(localAuth.HandleSignup))

	// API/CLI token-based authentication
	jwtSecret := os.Getenv("JWT_CLI_SECRET")
	if jwtSecret == "" {
		jwtSecret = "lilbattle-dev-secret-change-in-production" // Dev fallback
	}
	refreshTokenStore := oafs.NewFSRefreshTokenStore(storagePath)
	apiAuth := &oa.APIAuth{
		RefreshTokenStore:   refreshTokenStore,
		JWTSecretKey:        jwtSecret,
		JWTIssuer:           "lilbattle",
		JWTAudience:         "cli",
		AccessTokenExpiry:   30 * 24 * time.Hour, // 30 days for CLI tokens
		RefreshTokenExpiry:  90 * 24 * time.Hour, // 90 days
		ValidateCredentials: authService.ValidateLocalCredentials,
	}
	oneauth.AddAuth("/cli/token", apiAuth)

	// Wire up APIAuth's JWT validation to the Middleware so that
	// GetLoggedInUserId can validate Bearer tokens (for API/CLI clients)
	oneauth.Middleware.VerifyToken = apiAuth.VerifyTokenFunc()

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

	// Resend verification email
	oneauth.AddAuth("/resend-verification", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		if email == "" {
			http.Redirect(w, r, "/profile?verification_error=Email is required", http.StatusFound)
			return
		}

		// Get the identity to find the user ID
		identity, _, err := authService.IdentityStore.GetIdentity("email", email, false)
		if err != nil || identity == nil {
			// For security, don't reveal if email exists - just say success
			http.Redirect(w, r, "/profile?verification_sent=true", http.StatusFound)
			return
		}

		// Create verification token
		token, err := authService.TokenStore.CreateToken(
			identity.UserID,
			email,
			oa.TokenTypeEmailVerification,
			oa.TokenExpiryEmailVerification,
		)
		if err != nil {
			log.Printf("Error creating verification token: %v", err)
			http.Redirect(w, r, "/profile?verification_error=Failed to create verification token", http.StatusFound)
			return
		}

		// Send verification email
		verificationLink := baseURL + "/auth/verify-email?token=" + token.Token
		if err := localAuth.EmailSender.SendVerificationEmail(email, verificationLink); err != nil {
			log.Printf("Error sending verification email: %v", err)
			http.Redirect(w, r, "/profile?verification_error=Failed to send verification email", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/profile?verification_sent=true", http.StatusFound)
	}))

	// Change password endpoint (for users who already have a password)
	oneauth.AddAuth("/change-password", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error": "Method not allowed"}`))
			return
		}

		userId := oneauth.Middleware.GetLoggedInUserId(r)
		if userId == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Not logged in"}`))
			return
		}

		user, err := authService.GetUserById(userId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "User not found"}`))
			return
		}

		profile := user.Profile()
		email, ok := profile["email"].(string)
		if !ok || email == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "No email associated with account"}`))
			return
		}

		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid form data"}`))
			return
		}

		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")

		if currentPassword == "" || newPassword == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Current password and new password are required"}`))
			return
		}

		// Verify current password
		_, err = authService.ValidateLocalCredentials(email, currentPassword, "email")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Current password is incorrect"}`))
			return
		}

		// Update password
		if err := authService.UpdatePassword(email, newPassword); err != nil {
			log.Printf("Error updating password: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to update password"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))

	// Set password endpoint (for OAuth-only users setting password for the first time)
	oneauth.AddAuth("/set-password", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error": "Method not allowed"}`))
			return
		}

		userId := oneauth.Middleware.GetLoggedInUserId(r)
		if userId == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Not logged in"}`))
			return
		}

		user, err := authService.GetUserById(userId)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "User not found"}`))
			return
		}

		profile := user.Profile()
		email, ok := profile["email"].(string)
		if !ok || email == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "No email associated with account"}`))
			return
		}

		// Check if user already has a password - should use change-password instead
		identityKey := oa.IdentityKey("email", email)
		existingChannel, _, _ := authService.GetChannel("local", identityKey, false)
		if existingChannel != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Password already set. Use change-password endpoint instead."}`))
			return
		}

		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid form data"}`))
			return
		}

		newPassword := r.FormValue("new_password")
		if newPassword == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Password is required"}`))
			return
		}

		// Create local channel with password
		config := oa.EnsureAuthUserConfig{
			UserStore:     authService.UserStore,
			IdentityStore: authService.IdentityStore,
			ChannelStore:  authService.ChannelStore,
			UsernameStore: usernameStore,
		}

		// Get username from profile if set
		username, _ := profile["username"].(string)

		if err := oa.LinkLocalCredentials(config, userId, username, newPassword, email); err != nil {
			log.Printf("Error setting password for user %s: %v", userId, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to set password"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))

	return authService, usernameStore, oneauth
}
