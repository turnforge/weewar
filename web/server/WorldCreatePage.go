package server

import (
	"context"
	"fmt"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type WorldCreatePage struct {
	BasePage
	Header Header

	// Form fields (pre-filled on conflict retry)
	SuggestedId  string
	ErrorMessage string
	WorldName    string
}

func (p *WorldCreatePage) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	p.Title = "Create World"
	p.ActiveTab = "worlds"
	p.Header.Load(r, w, app)

	ctx := app.Context
	loggedInUserId := ctx.AuthMiddleware.GetLoggedInUserId(r)

	// Require login to access the create world page
	if loggedInUserId == "" {
		qs := r.URL.RawQuery
		if len(qs) > 0 {
			qs = "?" + qs
		}
		http.Redirect(w, r, fmt.Sprintf("/login?callbackURL=%s", fmt.Sprintf("/worlds/create%s", qs)), http.StatusSeeOther)
		return nil, true
	}

	if r.Method == http.MethodPost {
		// Handle form submission
		if err := r.ParseForm(); err != nil {
			p.ErrorMessage = "Failed to parse form"
			return nil, false
		}

		worldId := r.FormValue("worldId")
		worldName := r.FormValue("worldName")
		if worldName == "" {
			worldName = "Untitled World"
		}

		// Call CreateWorld RPC
		client := ctx.ClientMgr.GetWorldsSvcClient()
		createReq := &protos.CreateWorldRequest{
			World: &protos.World{
				Id:          worldId,
				Name:        worldName,
				Description: "",
				CreatorId:   loggedInUserId,
				Tags:        []string{},
			},
		}

		resp, err := client.CreateWorld(context.Background(), createReq)
		if err != nil {
			p.ErrorMessage = "Failed to create world: " + err.Error()
			p.WorldName = worldName
			p.SuggestedId = worldId
			return nil, false
		}

		// Check for field_errors (ID conflict)
		if len(resp.FieldErrors) > 0 {
			if suggestedId, ok := resp.FieldErrors["id"]; ok {
				p.SuggestedId = suggestedId
				p.ErrorMessage = fmt.Sprintf("ID '%s' already exists. Suggested: %s", worldId, suggestedId)
			} else {
				p.ErrorMessage = "Validation error"
			}
			p.WorldName = worldName
			return nil, false
		}

		// Success - redirect to editor
		editURL := fmt.Sprintf("/worlds/%s/edit", resp.World.Id)
		http.Redirect(w, r, editURL, http.StatusFound)
		return nil, true
	}

	// GET - show form (possibly with query params from external retry)
	p.SuggestedId = r.URL.Query().Get("suggestedId")
	p.ErrorMessage = r.URL.Query().Get("error")
	p.WorldName = r.URL.Query().Get("worldName")
	if p.WorldName == "" {
		p.WorldName = "Untitled World"
	}

	return nil, false
}
