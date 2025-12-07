package server

import (
	"net/http"

	goal "github.com/panyam/goapplib"
)

// LoginPage extends goapplib.SampleLoginPage with app-specific features.
type LoginPage struct {
	goal.SampleLoginPage[*WeewarApp]
	Header Header
}

// RegisterPage extends goapplib.SampleRegisterPage with app-specific features.
type RegisterPage struct {
	goal.SampleRegisterPage[*WeewarApp]
	Header Header
}

func (p *LoginPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*WeewarApp]) (err error, finished bool) {
	err, finished = goal.LoadAll(r, w, app, &p.SampleLoginPage, &p.Header)
	if err != nil || finished {
		return
	}

	p.Config = goal.LoginConfig{
		EnableEmailLogin:     true,
		EnableGoogleLogin:    true,
		EnableGitHubLogin:    true,
		EnableMicrosoftLogin: false,
		EnableAppleLogin:     false,
	}
	return
}

func (p *RegisterPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*WeewarApp]) (err error, finished bool) {
	return goal.LoadAll(r, w, app, &p.SampleRegisterPage, &p.Header)
}
