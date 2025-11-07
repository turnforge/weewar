package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	oa "github.com/panyam/oneauth"
	"github.com/panyam/oneauth/stores"
	"golang.org/x/oauth2"
)

// AuthService implements oa.AuthUserStore and orchestrates authentication
type AuthService struct {
	UserStore     oa.UserStore
	IdentityStore oa.IdentityStore
	ChannelStore  oa.ChannelStore
	TokenStore    oa.TokenStore
	NextID        func() string // Callback for generating user IDs
}

func NewAuthService(storagePath string) *AuthService {
	service := &AuthService{
		UserStore:     stores.NewFSUserStore(storagePath),
		IdentityStore: stores.NewFSIdentityStore(storagePath),
		ChannelStore:  stores.NewFSChannelStore(storagePath),
		TokenStore:    stores.NewFSTokenStore(storagePath),
		NextID:        defaultIDGenerator,
	}
	return service
}

// defaultIDGenerator generates a cryptographically secure random ID
func defaultIDGenerator() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateLocalUser creates a new user with local authentication
func (s *AuthService) CreateLocalUser(creds *oa.Credentials) (oa.User, error) {
	createFunc := oa.NewCreateUserFunc(s.UserStore, s.IdentityStore, s.ChannelStore)
	return createFunc(creds)
}

// ValidateLocalCredentials validates username/password and returns the user
func (s *AuthService) ValidateLocalCredentials(username, password, usernameType string) (oa.User, error) {
	validateFunc := oa.NewCredentialsValidator(s.IdentityStore, s.ChannelStore, s.UserStore)
	return validateFunc(username, password, usernameType)
}

// VerifyEmailByToken verifies an email using a verification token
func (s *AuthService) VerifyEmailByToken(token string) error {
	verifyFunc := oa.NewVerifyEmailFunc(s.IdentityStore, s.TokenStore)
	return verifyFunc(token)
}

// UpdatePassword updates the password for a user identified by email
func (s *AuthService) UpdatePassword(email, newPassword string) error {
	updateFunc := oa.NewUpdatePasswordFunc(s.IdentityStore, s.ChannelStore)
	return updateFunc(email, newPassword)
}

// Implement oa.UserStore interface
func (s *AuthService) CreateUser(userId string, isActive bool, profile map[string]any) (oa.User, error) {
	return s.UserStore.CreateUser(userId, isActive, profile)
}

func (s *AuthService) GetUserById(userId string) (oa.User, error) {
	return s.UserStore.GetUserById(userId)
}

func (s *AuthService) SaveUser(user oa.User) error {
	return s.UserStore.SaveUser(user)
}

// Implement oa.IdentityStore interface
func (s *AuthService) GetIdentity(identityType, identityValue string, createIfMissing bool) (*oa.Identity, bool, error) {
	return s.IdentityStore.GetIdentity(identityType, identityValue, createIfMissing)
}

func (s *AuthService) SaveIdentity(identity *oa.Identity) error {
	return s.IdentityStore.SaveIdentity(identity)
}

func (s *AuthService) SetUserForIdentity(identityType, identityValue string, newUserId string) error {
	return s.IdentityStore.SetUserForIdentity(identityType, identityValue, newUserId)
}

func (s *AuthService) MarkIdentityVerified(identityType, identityValue string) error {
	return s.IdentityStore.MarkIdentityVerified(identityType, identityValue)
}

func (s *AuthService) GetUserIdentities(userId string) ([]*oa.Identity, error) {
	return s.IdentityStore.GetUserIdentities(userId)
}

// Implement oa.ChannelStore interface
func (s *AuthService) GetChannel(provider string, identityKey string, createIfMissing bool) (*oa.Channel, bool, error) {
	return s.ChannelStore.GetChannel(provider, identityKey, createIfMissing)
}

func (s *AuthService) SaveChannel(channel *oa.Channel) error {
	return s.ChannelStore.SaveChannel(channel)
}

func (s *AuthService) GetChannelsByIdentity(identityKey string) ([]*oa.Channel, error) {
	return s.ChannelStore.GetChannelsByIdentity(identityKey)
}

// EnsureAuthUser is the main orchestration method for authentication
// It handles both OAuth and local auth, unifying identities across providers
func (s *AuthService) EnsureAuthUser(authtype, provider string, token *oauth2.Token, userInfo map[string]any) (oa.User, error) {
	// Extract primary identity (email or phone)
	var identityType, identityValue string

	if email, ok := userInfo["email"].(string); ok && email != "" {
		identityType = "email"
		identityValue = email
	} else if phone, ok := userInfo["phone"].(string); ok && phone != "" {
		identityType = "phone"
		identityValue = phone
	} else {
		return nil, fmt.Errorf("no valid identity found in userInfo")
	}

	identityKey := oa.IdentityKey(identityType, identityValue)

	// Get or create identity
	identity, newIdentity, err := s.IdentityStore.GetIdentity(identityType, identityValue, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	var user oa.User

	// If identity doesn't have a user, create one
	if identity.UserID == "" {
		userId := s.NextID()
		profile := userInfo
		if profile == nil {
			profile = make(map[string]any)
		}

		user, err = s.UserStore.CreateUser(userId, true, profile)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		identity.UserID = userId
		if err := s.IdentityStore.SaveIdentity(identity); err != nil {
			return nil, fmt.Errorf("failed to link identity to user: %w", err)
		}

		log.Printf("Created new user %s for identity %s", userId, identityKey)
	} else {
		// Load existing user
		user, err = s.UserStore.GetUserById(identity.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		log.Printf("Found existing user %s for identity %s", identity.UserID, identityKey)
	}

	// Get or create channel
	channel, newChannel, err := s.ChannelStore.GetChannel(provider, identityKey, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// Update channel credentials (e.g., OAuth tokens)
	if token != nil {
		if channel.Credentials == nil {
			channel.Credentials = make(map[string]any)
		}
		channel.Credentials["access_token"] = token.AccessToken
		channel.Credentials["refresh_token"] = token.RefreshToken
		channel.Credentials["token_type"] = token.TokenType
	}

	// Update channel profile
	if userInfo != nil {
		channel.Profile = userInfo
	}

	if err := s.ChannelStore.SaveChannel(channel); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// OAuth providers verify identities automatically
	if authtype == "oauth" && !identity.Verified {
		if err := s.IdentityStore.MarkIdentityVerified(identityType, identityValue); err != nil {
			log.Printf("Warning: failed to mark identity verified: %v", err)
		}
	}

	if newIdentity {
		log.Printf("Created new identity: %s", identityKey)
	}
	if newChannel {
		log.Printf("Created new channel: %s for %s", provider, identityKey)
	}

	return user, nil
}

