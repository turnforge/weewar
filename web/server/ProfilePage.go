package server

import (
	"net/http"

	goal "github.com/panyam/goapplib"
	oa "github.com/panyam/oneauth"
)

// ProfilePage extends goapplib.SampleProfilePage with app-specific features.
type ProfilePage struct {
	goal.SampleProfilePage[*WeewarApp]
	Header Header

	// App-specific user information
	User oa.User
}

func (p *ProfilePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*WeewarApp]) (err error, finished bool) {
	err, finished = goal.LoadAll(r, w, app, &p.SampleProfilePage, &p.Header)
	if err != nil || finished {
		return
	}

	ctx := app.Context
	p.UserID = ctx.AuthMiddleware.GetLoggedInUserId(r)
	if p.UserID == "" {
		http.Redirect(w, r, "/login?callbackURL=/profile", http.StatusFound)
		return nil, true
	}

	if ctx.AuthService != nil {
		p.User, err = ctx.AuthService.GetUserById(p.UserID)
		if err != nil || p.User == nil {
			http.Redirect(w, r, "/login?callbackURL=/profile", http.StatusFound)
			return nil, true
		}

		p.Profile = p.User.Profile()

		if email, ok := p.Profile["email"].(string); ok {
			p.Email = email
		}
		if username, ok := p.Profile["username"].(string); ok {
			p.Username = username
		}

		if p.Email != "" {
			identity, _, identityErr := ctx.AuthService.GetIdentity("email", p.Email, false)
			if identityErr == nil && identity != nil {
				p.EmailVerified = identity.Verified
			}
		}
	}

	return
}
