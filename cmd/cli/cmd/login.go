package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/panyam/oneauth/client"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	loginHost     string
	loginEmail    string
	loginPassword string
	loginToken    string
	loginReset    bool
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <profile>",
	Short: "Authenticate to a LilBattle server using a named profile",
	Long: `Create or update a named profile and authenticate to a LilBattle server.

A profile stores your server URL, email, and credentials for easy switching
between different servers or accounts.

Missing required fields (host, email, password) will be prompted for.
If a profile already exists, only provided flags will update existing values.

Authentication methods:
  - Interactive: Prompts for password (hidden input)
  - Token: Use --token flag to provide a pre-generated API token

Examples:
  ww login prod --host https://lilbattle.com --email user@example.com
  ww login local --host http://localhost:8080
  ww login prod --token eyJhbGc...
  ww login prod --reset                    # Force re-authentication

The profile is automatically selected as active after successful login.`,
	Args: cobra.ExactArgs(1),
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginHost, "host", "", "server URL (e.g., https://lilbattle.com)")
	loginCmd.Flags().StringVar(&loginEmail, "email", "", "email address for authentication")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "password (will prompt if not provided)")
	loginCmd.Flags().StringVar(&loginToken, "token", "", "API token (skip password authentication)")
	loginCmd.Flags().BoolVar(&loginReset, "reset", false, "force re-authentication even if valid credentials exist")
}

func runLogin(cmd *cobra.Command, args []string) error {
	profileNameArg := args[0]

	store, err := getProfileStore()
	if err != nil {
		return fmt.Errorf("failed to initialize profile store: %w", err)
	}

	formatter := NewOutputFormatter()
	reader := bufio.NewReader(os.Stdin)

	// Load existing profile if it exists
	existingProfile, _ := store.LoadProfile(profileNameArg)
	existingCreds, _ := store.LoadCredentials(profileNameArg)

	// Build profile from existing + flags
	profile := &Profile{Name: profileNameArg}
	if existingProfile != nil {
		profile = existingProfile
	}

	// Override with provided flags
	if loginHost != "" {
		profile.Host = loginHost
	}
	if loginEmail != "" {
		profile.Email = loginEmail
	}
	if loginPassword != "" {
		profile.Password = loginPassword
	}

	// Prompt for missing required fields
	if profile.Host == "" {
		fmt.Print("Host URL: ")
		host, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read host: %w", err)
		}
		profile.Host = strings.TrimSpace(host)
	}

	// Normalize host URL
	baseURL, err := extractServerBase(profile.Host)
	if err != nil {
		return fmt.Errorf("invalid host URL: %w", err)
	}
	profile.Host = baseURL

	// Check if we need to authenticate
	needsAuth := loginReset || loginToken != "" || existingCreds == nil || existingCreds.IsExpired()

	// If not forcing auth and have valid credentials, just update profile
	if !needsAuth && existingCreds != nil && !existingCreds.IsExpired() {
		if !formatter.JSON {
			fmt.Printf("Profile '%s' already has valid credentials.\n", profileNameArg)
			fmt.Print("Do you want to re-authenticate? [y/N]: ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				// Just save any profile updates
				if err := store.SaveProfile(profile); err != nil {
					return fmt.Errorf("failed to save profile: %w", err)
				}
				fmt.Println("Profile updated.")
				return nil
			}
		}
		needsAuth = true
	}

	var creds *ProfileCredentials

	if loginToken != "" {
		// Token-based login
		creds = &ProfileCredentials{
			AccessToken: loginToken,
			UserEmail:   "token-auth",
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:   time.Now(),
		}
	} else {
		// Interactive login - need email and password
		if profile.Email == "" {
			fmt.Print("Email: ")
			email, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read email: %w", err)
			}
			profile.Email = strings.TrimSpace(email)
		}

		// Get password - check profile first, then prompt
		password := profile.Password
		if password == "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			password = string(passwordBytes)

			// Ask if they want to save the password
			if !formatter.JSON {
				fmt.Print("Save password to profile? [y/N]: ")
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response == "y" || response == "yes" {
					profile.Password = password
				}
			}
		}

		// Create a temporary credential store for the auth client
		tempStore := &memCredStore{}
		authClient := client.NewAuthClient(profile.Host, tempStore)

		serverCred, err := authClient.Login(profile.Email, password, "read write profile offline")
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		creds = FromServerCredential(serverCred)
	}

	// Save profile and credentials
	if err := store.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	if err := store.SaveCredentials(profileNameArg, creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Auto-select this profile as the current one
	if err := store.SetCurrentProfile(profileNameArg); err != nil {
		return fmt.Errorf("failed to set current profile: %w", err)
	}

	if formatter.JSON {
		return formatter.PrintJSON(map[string]any{
			"profile":    profileNameArg,
			"host":       profile.Host,
			"email":      profile.Email,
			"user_id":    creds.UserID,
			"user_email": creds.UserEmail,
			"expires_at": creds.ExpiresAt,
		})
	}

	fmt.Printf("Successfully logged in to profile '%s' (now active)\n", profileNameArg)
	fmt.Printf("  Host: %s\n", profile.Host)
	fmt.Printf("  Email: %s\n", creds.UserEmail)
	fmt.Printf("  Expires: %s\n", creds.ExpiresAt.Format(time.RFC3339))

	return nil
}

// memCredStore is a minimal in-memory credential store for the auth client
type memCredStore struct {
	creds map[string]*client.ServerCredential
}

func (s *memCredStore) GetCredential(serverURL string) (*client.ServerCredential, error) {
	if s.creds == nil {
		return nil, nil
	}
	return s.creds[serverURL], nil
}

func (s *memCredStore) SetCredential(serverURL string, cred *client.ServerCredential) error {
	if s.creds == nil {
		s.creds = make(map[string]*client.ServerCredential)
	}
	s.creds[serverURL] = cred
	return nil
}

func (s *memCredStore) RemoveCredential(serverURL string) error {
	if s.creds != nil {
		delete(s.creds, serverURL)
	}
	return nil
}

func (s *memCredStore) ListServers() ([]string, error) {
	servers := make([]string, 0, len(s.creds))
	for k := range s.creds {
		servers = append(servers, k)
	}
	return servers, nil
}

func (s *memCredStore) Save() error {
	return nil
}
