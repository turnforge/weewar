// Package cmd provides CLI commands for lilbattle.
// This file wraps oneauth/client for credential management.
package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/panyam/oneauth/client"
	"github.com/panyam/oneauth/client/stores/fs"
)

// Re-export types from oneauth/client for backward compatibility
type ServerCredential = client.ServerCredential
type CredentialStore = client.CredentialStore

// Global credential store instance
var credentialStore client.CredentialStore

// getCredentialStore returns the singleton credential store
func getCredentialStore() (client.CredentialStore, error) {
	if credentialStore == nil {
		store, err := fs.NewFSCredentialStore("", "lilbattle")
		if err != nil {
			return nil, err
		}
		credentialStore = store
	}
	return credentialStore, nil
}

// LoadCredentialStore loads credentials from disk (backward compat)
func LoadCredentialStore() (client.CredentialStore, error) {
	return getCredentialStore()
}

// GetTokenForServer returns the token for a server, or empty string if not found/expired
// First checks profiles for a matching host, then falls back to legacy credentials
func GetTokenForServer(serverURL string) string {
	// First, check profiles for a matching host
	profileStore, err := getProfileStore()
	if err == nil {
		profiles, _ := profileStore.ListProfiles()
		for _, name := range profiles {
			profile, _ := profileStore.LoadProfile(name)
			if profile != nil && profile.Host == serverURL {
				creds, _ := profileStore.LoadCredentials(name)
				if creds != nil && !creds.IsExpired() {
					return creds.AccessToken
				}
			}
		}
	}

	// Fall back to legacy credential store
	store, err := getCredentialStore()
	if err != nil {
		return ""
	}

	cred, err := store.GetCredential(serverURL)
	if err != nil || cred == nil {
		return ""
	}

	if cred.IsExpired() {
		return ""
	}

	return cred.AccessToken
}

// extractServerBase extracts the base server URL from a full API URL
// e.g., "http://localhost:8080/api/v1/worlds/Desert" -> "http://localhost:8080"
func extractServerBase(fullURL string) (string, error) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" || u.Host == "" {
		return "", err
	}

	return u.Scheme + "://" + u.Host, nil
}

// extractWorldID extracts the world ID from a worlds API URL
// e.g., "http://localhost:8080/api/v1/worlds/Desert" -> "Desert"
func extractWorldID(worldURL string) (string, error) {
	u, err := url.Parse(worldURL)
	if err != nil {
		return "", err
	}

	// Look for /api/v1/worlds/<id> pattern
	path := u.Path
	const prefix = "/api/v1/worlds/"
	idx := strings.Index(path, prefix)
	if idx >= 0 {
		remainder := path[idx+len(prefix):]
		// Remove trailing slash if present
		remainder = strings.TrimSuffix(remainder, "/")
		if remainder == "" {
			return "", fmt.Errorf("no world ID found in URL: %s", worldURL)
		}
		return remainder, nil
	}

	return "", fmt.Errorf("URL does not match /api/v1/worlds/<id> pattern: %s", worldURL)
}

// GetAPIEndpoint returns the API endpoint URL for a given host.
// The Connect API is mounted at /api on the server, so this appends /api if not present.
func GetAPIEndpoint(host string) string {
	if strings.HasSuffix(host, "/api") || strings.HasSuffix(host, "/api/") {
		return host
	}
	return strings.TrimSuffix(host, "/") + "/api"
}

// WorldSpec represents a parsed world specification (profile:worldId or full URL)
type WorldSpec struct {
	Host        string
	WorldID     string
	ProfileName string // Set if parsed from profile:worldId format
	Token       string // Auth token for this server
}

// APIEndpoint returns the full API endpoint URL for this world spec
func (w *WorldSpec) APIEndpoint() string {
	return GetAPIEndpoint(w.Host)
}

// parseWorldSpec parses a world specification which can be either:
// - Full URL: http://localhost:8080/api/v1/worlds/Desert
// - Profile shorthand: profile:worldId (e.g., fsbe:01bdc3ce)
func parseWorldSpec(spec string) (*WorldSpec, error) {
	// Check if it looks like a URL (has scheme)
	if strings.HasPrefix(spec, "http://") || strings.HasPrefix(spec, "https://") {
		host, err := extractServerBase(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		worldID, err := extractWorldID(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		token := GetTokenForServer(host)
		return &WorldSpec{
			Host:    host,
			WorldID: worldID,
			Token:   token,
		}, nil
	}

	// Try to parse as profile:worldId
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid format: expected 'profile:worldId' or full URL, got: %s", spec)
	}

	profileName := parts[0]
	worldID := parts[1]

	// Look up the profile
	store, err := getProfileStore()
	if err != nil {
		return nil, fmt.Errorf("failed to get profile store: %w", err)
	}

	profile, err := store.LoadProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("profile '%s' not found: %w", profileName, err)
	}

	// Get credentials for this profile
	creds, _ := store.LoadCredentials(profileName)
	var token string
	if creds != nil && !creds.IsExpired() {
		token = creds.AccessToken
	}

	return &WorldSpec{
		Host:        profile.Host,
		WorldID:     worldID,
		ProfileName: profileName,
		Token:       token,
	}, nil
}
