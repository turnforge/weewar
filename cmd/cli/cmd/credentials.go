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
func GetTokenForServer(serverURL string) string {
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
