package server

import (
	oa "github.com/panyam/oneauth"
	"github.com/turnforge/weewar/services"
	svc "github.com/turnforge/weewar/services"
	"golang.org/x/oauth2"
)

func (n *App) GetUserByID(userId string) (oa.User, error) {
	var err error
	if userId == "test1" {
		// Mocking user login
		return &svc.User{
			ID: "test1",
			ProfileInfo: svc.StringMapField{
				Properties: map[string]any{
					"Name": "Test User",
				},
			},
		}, nil
	}
	u, err := n.ClientMgr.GetAuthService().GetUserById(userId)
	return u.(*services.User), err
}

func (n *App) EnsureAuthUser(authtype string, provider string, token *oauth2.Token, userInfo map[string]any) (oa.User, error) {
	var err error
	// Mocking user login
	email := userInfo["email"].(string)
	if email == "test@gmail.com" {
		return &svc.User{
			ID: "test1",
			ProfileInfo: svc.StringMapField{
				Properties: map[string]any{
					"Name": "Test User",
				},
			},
		}, nil
	}
	user, err := n.ClientMgr.GetAuthService().EnsureAuthUser(authtype, provider, token, userInfo)
	return user.(*services.User), err
}

func (n *App) ValidateUsernamePassword(username string, password string) (out oa.User, err error) {
	if username == "test@gmail.com" {
		out = &svc.User{
			ID: "test1",
			ProfileInfo: svc.StringMapField{
				Properties: map[string]any{
					"Name": "Test User",
				},
			},
		}
	}
	return
}
