package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexedwards/scs/v2"
	goalservices "github.com/panyam/goapplib/services"
	oa "github.com/panyam/oneauth"
	oa2 "github.com/panyam/oneauth/oauth2"
)

func setupAuthService(session *scs.SessionManager) (*goalservices.AuthService, *oa.OneAuth) {
	// Initialize authentication
	storagePath := os.Getenv("WEEWAR_USER_STORAGE_PATH")
	if storagePath == "" {
		storagePath = filepath.Join(os.Getenv("HOME"), "dev-app-data", "weewar", "storage")
	}
	authService := goalservices.NewAuthService(storagePath)
	oneauth := oa.New("weewar")
	oneauth.Session = session
	oneauth.Middleware.SessionGetter = func(r *http.Request, key string) any {
		return session.GetString(r.Context(), key)
	}
	oneauth.UserStore = authService
	// OAuth providers
	oneauth.AddAuth("/google", oa2.NewGoogleOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/github", oa2.NewGithubOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
	oneauth.AddAuth("/twitter", NewTwitterOAuth2("", "", "", oneauth.SaveUserAndRedirect).Handler())
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
		RequireEmailVerification: false,   // Optional verification
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

	oneauth.AddAuth("/change-password", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error": "Method not allowed"}`))
			return
		}

		// Get logged in user ID
		userId := oneauth.Middleware.GetLoggedInUserId(r)
		if userId == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Not logged in"}`))
			return
		}

		// Get user to find their email
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

		// Parse form data
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
	return authService, oneauth
}
