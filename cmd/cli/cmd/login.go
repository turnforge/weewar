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

func runLogin(cmd *cobra.Command, args []string) error {
	serverURL := args[0]

	// Normalize server URL
	baseURL, err := extractServerBase(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	formatter := NewOutputFormatter()

	// Get credential store
	store, err := getCredentialStore()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Check if already logged in
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

	var cred *client.ServerCredential

	if loginToken != "" {
		// Token-based login - store directly
		cred = &client.ServerCredential{
			AccessToken: loginToken,
			UserEmail:   "token-auth",
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:   time.Now(),
		}
		if err := store.SetCredential(baseURL, cred); err != nil {
			return fmt.Errorf("failed to store credential: %w", err)
		}
		if err := store.Save(); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}
	} else {
		// Interactive login using AuthClient
		authClient := client.NewAuthClient(baseURL, store)

		reader := bufio.NewReader(os.Stdin)

		// Prompt for email
		fmt.Print("Email: ")
		email, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(email)

		// Prompt for password (hidden)
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Add newline after password input
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password := string(passwordBytes)

		// Login via AuthClient
		cred, err = authClient.Login(email, password, "read write profile offline")
		if err != nil {
			return err
		}
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
