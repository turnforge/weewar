package server

import (
	"net/http"
	"os"

	goal "github.com/panyam/goapplib"
)

// LoginPage extends goapplib.SampleLoginPage with app-specific features.
type LoginPage struct {
	goal.SampleLoginPage[*LilBattleApp]
	Header             Header
	EnableTwitterLogin bool
}

// RegisterPage extends goapplib.SampleRegisterPage with app-specific features.
type RegisterPage struct {
	goal.SampleRegisterPage[*LilBattleApp]
	Header Header
}

func (p *LoginPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
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
	// Enable Twitter login if credentials are configured
	p.EnableTwitterLogin = os.Getenv("OAUTH2_TWITTER_CLIENT_ID") != ""
	return
}

func (p *RegisterPage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	return goal.LoadAll(r, w, app, &p.SampleRegisterPage, &p.Header)
}
