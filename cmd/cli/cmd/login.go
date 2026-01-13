package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	loginToken string // For --token flag
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <server>",
	Short: "Authenticate to a LilBattle server",
	Long: `Authenticate to a LilBattle server and store credentials locally.

The server URL should be the base URL of the server (e.g., http://localhost:8080).

Authentication methods:
  - Interactive: Prompts for email and password
  - Token: Use --token flag to provide a pre-generated API token

Credentials are stored in ~/.config/lilbattle/credentials.json with
restricted permissions (readable only by owner).

Examples:
  ww login http://localhost:8080
  ww login https://lilbattle.example.com
  ww login http://localhost:8080 --token eyJhbGc...`,
	Args: cobra.ExactArgs(1),
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginToken, "token", "", "API token (skip interactive login)")
}

// OAuth2TokenRequest is the request body for /auth/cli/token (OAuth2 password grant)
type OAuth2TokenRequest struct {
	GrantType string `json:"grant_type"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
}

// OAuth2TokenResponse is the response from /auth/cli/token
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

func runLogin(cmd *cobra.Command, args []string) error {
	serverURL := args[0]

	// Normalize server URL
	baseURL, err := extractServerBase(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	formatter := NewOutputFormatter()

	// Check if already logged in
	store, err := LoadCredentialStore()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	existingCred, _ := store.GetCredential(baseURL)
	if existingCred != nil && !existingCred.IsExpired() {
		if !formatter.JSON {
			fmt.Printf("Already logged in to %s as %s\n", baseURL, existingCred.UserEmail)
			fmt.Print("Do you want to re-authenticate? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				return nil
			}
		}
	}

	var cred *ServerCredential

	if loginToken != "" {
		// Token-based login
		cred, err = loginWithToken(baseURL, loginToken)
	} else {
		// Interactive login
		cred, err = loginInteractive(baseURL)
	}

	if err != nil {
		return err
	}

	// Store the credential
	if err := store.SetCredential(baseURL, cred); err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	if formatter.JSON {
		return formatter.PrintJSON(map[string]any{
			"server":     baseURL,
			"user_id":    cred.UserID,
			"user_email": cred.UserEmail,
			"expires_at": cred.ExpiresAt,
		})
	}

	fmt.Printf("Successfully logged in to %s as %s\n", baseURL, cred.UserEmail)
	fmt.Printf("Token expires: %s\n", cred.ExpiresAt.Format(time.RFC3339))
	return nil
}

func loginInteractive(serverURL string) (*ServerCredential, error) {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for email
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	// Prompt for password (hidden)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add newline after password input
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	// Make the token request
	return requestToken(serverURL, email, password)
}

func requestToken(serverURL, email, password string) (*ServerCredential, error) {
	tokenURL := serverURL + "/auth/cli/token"

	reqBody := OAuth2TokenRequest{
		GrantType: "password",
		Username:  email,
		Password:  password,
		Scope:     "read write profile offline",
		ClientID:  "cli",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	resp, err := http.Post(tokenURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp OAuth2TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("invalid response from server: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if tokenResp.ErrorDesc != "" {
			return nil, fmt.Errorf("authentication failed: %s", tokenResp.ErrorDesc)
		}
		if tokenResp.Error != "" {
			return nil, fmt.Errorf("authentication failed: %s", tokenResp.Error)
		}
		return nil, fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}

	// Calculate expiration time from expires_in seconds
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &ServerCredential{
		Token:     tokenResp.AccessToken,
		UserID:    "", // JWT tokens don't include user ID in response, extracted from token claims
		UserEmail: email,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

func loginWithToken(serverURL, token string) (*ServerCredential, error) {
	// For pre-generated tokens (e.g., API keys or JWTs from other sources),
	// store them directly. The token will be validated on first use.
	// We assume a 30-day expiration for manually provided tokens.
	return &ServerCredential{
		Token:     token,
		UserID:    "",
		UserEmail: "token-auth",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}
