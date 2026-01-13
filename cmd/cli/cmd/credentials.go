package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ServerCredential holds authentication info for a single server
type ServerCredential struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired returns true if the credential has expired
func (c *ServerCredential) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// CredentialStore manages server credentials
type CredentialStore struct {
	Servers map[string]*ServerCredential `json:"servers"`
}

// NewCredentialStore creates an empty credential store
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		Servers: make(map[string]*ServerCredential),
	}
}

// getCredentialsPath returns the path to the credentials file
func getCredentialsPath() (string, error) {
	// Check for custom config location
	if cfgFile != "" {
		dir := filepath.Dir(cfgFile)
		return filepath.Join(dir, "credentials.json"), nil
	}

	// Default to ~/.config/lilbattle/credentials.json
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine config directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	lilbattleDir := filepath.Join(configDir, "lilbattle")
	return filepath.Join(lilbattleDir, "credentials.json"), nil
}

// LoadCredentialStore loads credentials from disk
func LoadCredentialStore() (*CredentialStore, error) {
	path, err := getCredentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewCredentialStore(), nil
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var store CredentialStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	if store.Servers == nil {
		store.Servers = make(map[string]*ServerCredential)
	}

	return &store, nil
}

// Save writes the credential store to disk
func (s *CredentialStore) Save() error {
	path, err := getCredentialsPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// normalizeServerURL normalizes a server URL for use as a key
func normalizeServerURL(serverURL string) (string, error) {
	// Parse the URL
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	// Use scheme + host as the key (ignore path)
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

// SetCredential stores a credential for a server
func (s *CredentialStore) SetCredential(serverURL string, cred *ServerCredential) error {
	key, err := normalizeServerURL(serverURL)
	if err != nil {
		return err
	}
	s.Servers[key] = cred
	return nil
}

// GetCredential retrieves a credential for a server
func (s *CredentialStore) GetCredential(serverURL string) (*ServerCredential, error) {
	key, err := normalizeServerURL(serverURL)
	if err != nil {
		return nil, err
	}

	cred, ok := s.Servers[key]
	if !ok {
		return nil, nil
	}

	return cred, nil
}

// RemoveCredential removes a credential for a server
func (s *CredentialStore) RemoveCredential(serverURL string) error {
	key, err := normalizeServerURL(serverURL)
	if err != nil {
		return err
	}

	delete(s.Servers, key)
	return nil
}

// GetTokenForServer returns the token for a server, or empty string if not found/expired
func GetTokenForServer(serverURL string) string {
	store, err := LoadCredentialStore()
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

	return cred.Token
}

// extractServerBase extracts the base server URL from a full API URL
// e.g., "http://localhost:8080/api/v1/worlds/Desert" -> "http://localhost:8080"
func extractServerBase(fullURL string) (string, error) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("URL must include scheme and host: %s", fullURL)
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

// extractWorldID extracts the world ID from a worlds API URL
// e.g., "http://localhost:8080/api/v1/worlds/Desert" -> "Desert"
func extractWorldID(worldURL string) (string, error) {
	u, err := url.Parse(worldURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Look for /api/v1/worlds/<id> pattern
	path := u.Path
	if idx := strings.Index(path, "/api/v1/worlds/"); idx >= 0 {
		remainder := path[idx+len("/api/v1/worlds/"):]
		// Remove trailing slash if present
		remainder = strings.TrimSuffix(remainder, "/")
		if remainder == "" {
			return "", fmt.Errorf("no world ID found in URL: %s", worldURL)
		}
		return remainder, nil
	}

	return "", fmt.Errorf("URL does not match /api/v1/worlds/<id> pattern: %s", worldURL)
}
