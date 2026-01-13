package server

import (
	"context"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type WorldListView struct {
	goal.WithPagination
	goal.WithFiltering

	Worlds     []*protos.World
	ActionMode string // "manage" or "select"
}

func (p *WorldListView) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	// Load pagination and filtering using goal Load methods
	p.WithPagination.Load(r, w, nil)
	p.WithFiltering.Load(r, w, nil)

	// Override defaults for this view
	if p.ViewMode == "" {
		p.ViewMode = "grid"
	}
	if p.Sort == "" {
		p.Sort = "modified_desc"
	}
	if p.ActionMode == "" {
		p.ActionMode = "manage"
	}

	ctx := app.Context
	userId := ctx.AuthMiddleware.GetLoggedInUserId(r)
	client := ctx.ClientMgr.GetWorldsSvcClient()

	req := protos.ListWorldsRequest{
		Pagination: &protos.Pagination{
			PageOffset: int32(p.Offset()),
			PageSize:   int32(p.PageSize),
		},
		OwnerId: userId,
	}
	resp, err := client.ListWorlds(context.Background(), &req)
	if err != nil {
		log.Println("error getting worlds: ", err)
		return HandleGRPCError(err, w, r, app)
	}
	p.Worlds = resp.Items
	p.HasPrevPage = p.CurrentPage > 0
	if resp.Pagination != nil {
		p.HasNextPage = resp.Pagination.HasMore
		p.TotalCount = int(resp.Pagination.TotalResults)
		p.EvalPages()
	}
	return nil, false
}
