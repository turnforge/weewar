// Package cmd provides CLI commands for lilbattle.
// This file implements profile-based credential management.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/panyam/oneauth/client"
)

// Profile represents a named configuration profile
type Profile struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"` // Stored only for local FS profiles
}

// ProfileCredentials stores authentication tokens for a profile
type ProfileCredentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	UserEmail    string    `json:"user_email,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// IsExpired checks if the credentials have expired
func (c *ProfileCredentials) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// ToServerCredential converts to oneauth ServerCredential for API compatibility
func (c *ProfileCredentials) ToServerCredential() *client.ServerCredential {
	return &client.ServerCredential{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		UserID:       c.UserID,
		UserEmail:    c.UserEmail,
		ExpiresAt:    c.ExpiresAt,
		CreatedAt:    c.CreatedAt,
	}
}

// FromServerCredential creates ProfileCredentials from oneauth ServerCredential
func FromServerCredential(cred *client.ServerCredential) *ProfileCredentials {
	return &ProfileCredentials{
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
		UserID:       cred.UserID,
		UserEmail:    cred.UserEmail,
		ExpiresAt:    cred.ExpiresAt,
		CreatedAt:    cred.CreatedAt,
	}
}

// GlobalConfig stores global CLI configuration
type GlobalConfig struct {
	CurrentProfile string `json:"current_profile,omitempty"`
}

// ProfileStore manages profile storage
type ProfileStore struct {
	baseDir string
}

// Global profile store instance
var profileStore *ProfileStore

// getProfileStore returns the singleton profile store
func getProfileStore() (*ProfileStore, error) {
	if profileStore == nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir := filepath.Join(homeDir, ".config", "lilbattle")
		profileStore = &ProfileStore{baseDir: baseDir}
	}
	return profileStore, nil
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, 0700)
}

// profileDir returns the directory for a specific profile
func (s *ProfileStore) profileDir(name string) string {
	return filepath.Join(s.baseDir, "profiles", name)
}

// profilePath returns the path to the profile.json file
func (s *ProfileStore) profilePath(name string) string {
	return filepath.Join(s.profileDir(name), "profile.json")
}

// credentialsPath returns the path to the credentials.json file
func (s *ProfileStore) credentialsPath(name string) string {
	return filepath.Join(s.profileDir(name), "credentials.json")
}

// configPath returns the path to the global config file
func (s *ProfileStore) configPath() string {
	return filepath.Join(s.baseDir, "config.json")
}

// LoadGlobalConfig loads the global configuration
func (s *ProfileStore) LoadGlobalConfig() (*GlobalConfig, error) {
	data, err := os.ReadFile(s.configPath())
	if os.IsNotExist(err) {
		return &GlobalConfig{}, nil
	}
	if err != nil {
		return nil, err
	}

	var config GlobalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveGlobalConfig saves the global configuration
func (s *ProfileStore) SaveGlobalConfig(config *GlobalConfig) error {
	if err := ensureDir(s.baseDir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath(), data, 0600)
}

// GetCurrentProfile returns the name of the currently selected profile
func (s *ProfileStore) GetCurrentProfile() (string, error) {
	config, err := s.LoadGlobalConfig()
	if err != nil {
		return "", err
	}
	return config.CurrentProfile, nil
}

// SetCurrentProfile sets the currently selected profile
func (s *ProfileStore) SetCurrentProfile(name string) error {
	// Verify profile exists
	if _, err := s.LoadProfile(name); err != nil {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	config, err := s.LoadGlobalConfig()
	if err != nil {
		return err
	}
	config.CurrentProfile = name
	return s.SaveGlobalConfig(config)
}

// ListProfiles returns all profile names
func (s *ProfileStore) ListProfiles() ([]string, error) {
	profilesDir := filepath.Join(s.baseDir, "profiles")
	entries, err := os.ReadDir(profilesDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if profile.json exists
			profilePath := filepath.Join(profilesDir, entry.Name(), "profile.json")
			if _, err := os.Stat(profilePath); err == nil {
				profiles = append(profiles, entry.Name())
			}
		}
	}
	return profiles, nil
}

// LoadProfile loads a profile by name
func (s *ProfileStore) LoadProfile(name string) (*Profile, error) {
	data, err := os.ReadFile(s.profilePath(name))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("profile '%s' does not exist", name)
	}
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	profile.Name = name // Ensure name matches directory
	return &profile, nil
}

// SaveProfile saves a profile
func (s *ProfileStore) SaveProfile(profile *Profile) error {
	dir := s.profileDir(profile.Name)
	if err := ensureDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.profilePath(profile.Name), data, 0600)
}

// LoadCredentials loads credentials for a profile
func (s *ProfileStore) LoadCredentials(name string) (*ProfileCredentials, error) {
	data, err := os.ReadFile(s.credentialsPath(name))
	if os.IsNotExist(err) {
		return nil, nil // No credentials yet
	}
	if err != nil {
		return nil, err
	}

	var creds ProfileCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// SaveCredentials saves credentials for a profile
func (s *ProfileStore) SaveCredentials(name string, creds *ProfileCredentials) error {
	dir := s.profileDir(name)
	if err := ensureDir(dir); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.credentialsPath(name), data, 0600)
}

// DeleteProfile removes a profile and all its data
func (s *ProfileStore) DeleteProfile(name string) error {
	dir := s.profileDir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Clear current profile if it's being deleted
	config, _ := s.LoadGlobalConfig()
	if config != nil && config.CurrentProfile == name {
		config.CurrentProfile = ""
		s.SaveGlobalConfig(config)
	}

	return os.RemoveAll(dir)
}

// GetActiveProfile returns the active profile based on precedence:
// 1. --profile flag
// 2. Current selected profile
func GetActiveProfile(profileFlag string) (*Profile, error) {
	store, err := getProfileStore()
	if err != nil {
		return nil, err
	}

	var profileName string

	if profileFlag != "" {
		profileName = profileFlag
	} else {
		profileName, err = store.GetCurrentProfile()
		if err != nil {
			return nil, err
		}
	}

	if profileName == "" {
		return nil, nil // No profile active
	}

	return store.LoadProfile(profileName)
}

// GetActiveCredentials returns credentials for the active profile
func GetActiveCredentials(profileFlag string) (*ProfileCredentials, error) {
	store, err := getProfileStore()
	if err != nil {
		return nil, err
	}

	var profileName string

	if profileFlag != "" {
		profileName = profileFlag
	} else {
		profileName, err = store.GetCurrentProfile()
		if err != nil {
			return nil, err
		}
	}

	if profileName == "" {
		return nil, nil
	}

	return store.LoadCredentials(profileName)
}

// GetTokenForProfile returns the access token for a profile, or empty if not found/expired
func GetTokenForProfile(profileFlag string) string {
	creds, err := GetActiveCredentials(profileFlag)
	if err != nil || creds == nil {
		return ""
	}

	if creds.IsExpired() {
		return ""
	}

	return creds.AccessToken
}
