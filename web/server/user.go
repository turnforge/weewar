package server

import (
	"os"

	oa "github.com/panyam/oneauth"
	svc "github.com/turnforge/lilbattle/services"
	"golang.org/x/oauth2"
)

// testAuthEnabled returns true if test authentication is enabled via env var.
// This should only be enabled in development/testing environments.
func testAuthEnabled() bool {
	return os.Getenv("ENABLE_TEST_AUTH") == "true"
}

// testUser returns the mock test user for development/testing.
func testUser() *svc.User {
	return &svc.User{
		ID: "test1",
		ProfileInfo: svc.StringMapField{
			Properties: map[string]any{
				"Name": "Test User",
			},
		},
	}
}

func (n *LilBattleApp) GetUserByID(userId string) (oa.User, error) {
	var err error
	// Test user bypass - only if ENABLE_TEST_AUTH is set
	if testAuthEnabled() && userId == "test1" {
		return testUser(), nil
	}
	u, err := n.ClientMgr.GetAuthService().GetUserById(userId)
	return u.(*svc.User), err
}

func (n *LilBattleApp) EnsureAuthUser(authtype string, provider string, token *oauth2.Token, userInfo map[string]any) (oa.User, error) {
	var err error
	// Test user bypass - only if ENABLE_TEST_AUTH is set
	if testAuthEnabled() {
		if email, ok := userInfo["email"].(string); ok && email == "test@gmail.com" {
			return testUser(), nil
		}
	}

	// Assign a random nickname if not already set
	if _, hasNickname := userInfo["nickname"]; !hasNickname {
		userInfo["nickname"] = GenerateRandomNickname()
	}

	user, err := n.ClientMgr.GetAuthService().EnsureAuthUser(authtype, provider, token, userInfo)
	return user.(*svc.User), err
}

func (n *LilBattleApp) ValidateUsernamePassword(username string, password string) (out oa.User, err error) {
	// Test user bypass - only if ENABLE_TEST_AUTH is set
	if testAuthEnabled() && username == "test@gmail.com" {
		out = testUser()
		return
	}
	// For production, delegate to auth service
	usernameType := oa.DetectUsernameType(username)
	user, err := n.ClientMgr.GetAuthService().ValidateLocalCredentials(username, password, usernameType)
	if err != nil {
		return nil, err
	}
	out = user
	return
}
