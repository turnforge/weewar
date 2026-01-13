package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout <server>",
	Short: "Remove stored credentials for a server",
	Long: `Remove stored credentials for a LilBattle server.

If no server is specified, lists all servers with stored credentials.

Examples:
  ww logout http://localhost:8080
  ww logout https://lilbattle.example.com
  ww logout                              # List all logged-in servers`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	store, err := LoadCredentialStore()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	formatter := NewOutputFormatter()

	// If no server specified, list all servers
	if len(args) == 0 {
		if len(store.Servers) == 0 {
			if formatter.JSON {
				return formatter.PrintJSON(map[string]any{
					"servers": []any{},
				})
			}
			fmt.Println("No servers configured. Use 'ww login <server>' to authenticate.")
			return nil
		}

		if formatter.JSON {
			servers := make([]map[string]any, 0, len(store.Servers))
			for serverURL, cred := range store.Servers {
				servers = append(servers, map[string]any{
					"server":     serverURL,
					"user_email": cred.UserEmail,
					"expires_at": cred.ExpiresAt,
					"expired":    cred.IsExpired(),
				})
			}
			return formatter.PrintJSON(map[string]any{
				"servers": servers,
			})
		}

		fmt.Println("Logged in servers:")
		for serverURL, cred := range store.Servers {
			status := ""
			if cred.IsExpired() {
				status = " (expired)"
			}
			fmt.Printf("  %s - %s%s\n", serverURL, cred.UserEmail, status)
		}
		fmt.Println("\nUse 'ww logout <server>' to remove credentials.")
		return nil
	}

	serverURL := args[0]

	// Normalize server URL
	baseURL, err := extractServerBase(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	// Check if credential exists
	cred, err := store.GetCredential(baseURL)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	if cred == nil {
		if formatter.JSON {
			return formatter.PrintJSON(map[string]any{
				"server":  baseURL,
				"removed": false,
				"message": "not logged in",
			})
		}
		fmt.Printf("Not logged in to %s\n", baseURL)
		return nil
	}

	// Remove the credential
	if err := store.RemoveCredential(baseURL); err != nil {
		return fmt.Errorf("failed to remove credential: %w", err)
	}

	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	if formatter.JSON {
		return formatter.PrintJSON(map[string]any{
			"server":  baseURL,
			"removed": true,
		})
	}

	fmt.Printf("Logged out from %s\n", baseURL)
	return nil
}
