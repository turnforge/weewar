package server

import (
	"context"
	"log"
	"net/http"

	goal "github.com/panyam/goapplib"
	protos "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

type GameListView struct {
	goal.WithPagination
	goal.WithFiltering

	Games []*protos.Game
}

func (p *GameListView) Load(r *http.Request, w http.ResponseWriter, app *goal.App[*LilBattleApp]) (err error, finished bool) {
	// Load pagination and filtering using goal Load methods
	p.WithPagination.Load(r, w, nil)
	p.WithFiltering.Load(r, w, nil)

	ctx := app.Context
	userId := ctx.AuthMiddleware.GetLoggedInUserId(r)
	client := ctx.ClientMgr.GetGamesSvcClient()

	req := protos.ListGamesRequest{
		Pagination: &protos.Pagination{
			PageOffset: int32(p.Offset()),
			PageSize:   int32(p.PageSize),
		},
		OwnerId: userId,
	}
	resp, err := client.ListGames(context.Background(), &req)
	if err != nil {
		log.Println("error getting games: ", err)
		return HandleGRPCError(err, w, r, app)
	}
	log.Println("Found Games: ", resp.Items)
	p.Games = resp.Items
	p.HasPrevPage = p.CurrentPage > 0
	if resp.Pagination != nil {
		p.HasNextPage = resp.Pagination.HasMore
		p.TotalCount = int(resp.Pagination.TotalResults)
		p.EvalPages()
	}
	return nil, false
}
