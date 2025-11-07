package server

import (
	"net/http"
)

type LoginConfig struct {
	EnableEmailLogin    bool
	EnableGoogleLogin   bool
	EnableGitHubLogin   bool
	EnableMicrosoftLogin bool
	EnableAppleLogin    bool
}

type LoginPage struct {
	BasePage
	Header      Header
	CallbackURL string
	CsrfToken   string
	Config      LoginConfig
}

type RegisterPage struct {
	Header         Header
	CallbackURL    string
	CsrfToken      string
	Name           string
	Email          string
	Password       string
	VerifyPassword string
	Errors         map[string]string
}

func (p *LoginPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.DisableSplashScreen = true
	err, finished = p.Header.Load(r, w, vc)
	p.CallbackURL = r.URL.Query().Get("callbackURL")

	// Initialize login config - these can be overridden by environment variables or config
	p.Config = LoginConfig{
		EnableEmailLogin:     true,
		EnableGoogleLogin:    true,
		EnableGitHubLogin:    true,
		EnableMicrosoftLogin: false, // Can be enabled when needed
		EnableAppleLogin:     false, // Can be enabled when needed
	}
	return
}

func (p *RegisterPage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	err, finished = p.Header.Load(r, w, vc)
	p.CallbackURL = r.URL.Query().Get("callbackURL")
	return
}

func (g *LoginPage) Copy() View    { return &LoginPage{} }
func (g *RegisterPage) Copy() View { return &RegisterPage{} }
