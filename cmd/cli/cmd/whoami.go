package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami [server]",
	Short: "Show current authentication status",
	Long: `Show the currently authenticated user for a server.

If no server is specified, shows authentication status for all configured servers.

Examples:
  ww whoami http://localhost:8080
  ww whoami                              # Show all servers`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	store, err := LoadCredentialStore()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	formatter := NewOutputFormatter()

	// If no server specified, show all servers
	if len(args) == 0 {
		if len(store.Servers) == 0 {
			if formatter.JSON {
				return formatter.PrintJSON(map[string]any{
					"authenticated": false,
					"servers":       []any{},
				})
			}
			fmt.Println("Not logged in to any servers.")
			fmt.Println("Use 'ww login <server>' to authenticate.")
			return nil
		}

		if formatter.JSON {
			servers := make([]map[string]any, 0, len(store.Servers))
			for serverURL, cred := range store.Servers {
				servers = append(servers, map[string]any{
					"server":     serverURL,
					"user_id":    cred.UserID,
					"user_email": cred.UserEmail,
					"expires_at": cred.ExpiresAt,
					"expired":    cred.IsExpired(),
				})
			}
			return formatter.PrintJSON(map[string]any{
				"authenticated": true,
				"servers":       servers,
			})
		}

		fmt.Println("Authentication status:")
		for serverURL, cred := range store.Servers {
			status := "valid"
			if cred.IsExpired() {
				status = "EXPIRED"
			} else {
				remaining := time.Until(cred.ExpiresAt)
				if remaining < 24*time.Hour {
					status = fmt.Sprintf("expires in %s", remaining.Round(time.Minute))
				}
			}
			fmt.Printf("  %s\n", serverURL)
			fmt.Printf("    User: %s (%s)\n", cred.UserEmail, cred.UserID)
			fmt.Printf("    Status: %s\n", status)
		}
		return nil
	}

	serverURL := args[0]

	// Normalize server URL
	baseURL, err := extractServerBase(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	cred, err := store.GetCredential(baseURL)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	if cred == nil {
		if formatter.JSON {
			return formatter.PrintJSON(map[string]any{
				"server":        baseURL,
				"authenticated": false,
			})
		}
		fmt.Printf("Not logged in to %s\n", baseURL)
		fmt.Println("Use 'ww login <server>' to authenticate.")
		return nil
	}

	if formatter.JSON {
		return formatter.PrintJSON(map[string]any{
			"server":        baseURL,
			"authenticated": true,
			"user_id":       cred.UserID,
			"user_email":    cred.UserEmail,
			"expires_at":    cred.ExpiresAt,
			"expired":       cred.IsExpired(),
		})
	}

	fmt.Printf("Server: %s\n", baseURL)
	fmt.Printf("User: %s (%s)\n", cred.UserEmail, cred.UserID)
	if cred.IsExpired() {
		fmt.Printf("Status: EXPIRED (expired %s)\n", time.Since(cred.ExpiresAt).Round(time.Minute))
		fmt.Println("Use 'ww login <server>' to re-authenticate.")
	} else {
		remaining := time.Until(cred.ExpiresAt)
		fmt.Printf("Status: Valid (expires in %s)\n", remaining.Round(time.Minute))
	}

	return nil
}
