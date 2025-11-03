package server

import (
	"context"
	"net/http"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1/models"
)

type BasePage struct {
	Title               string
	BodyClass           string
	CustomHeader        bool
	BodyDataAttributes  string
	DisableSplashScreen bool
	SplashTitle         string
	SplashMessage       string
}

type HomePage struct {
	BasePage
	Header Header

	// Dashboard data
	RecentGames  []*v1.Game
	RecentWorlds []*v1.World
	TotalGames   int32
	TotalWorlds  int32
}

func (p *HomePage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.Title = "Home"
	p.DisableSplashScreen = true
	p.Header.Load(r, w, vc)

	// Fetch recent games (limit to 6)
	gamesClient, err := vc.ClientMgr.GetGamesSvcClient()
	if err == nil {
		gamesResp, err := gamesClient.ListGames(context.Background(), &v1.ListGamesRequest{
			Pagination: &v1.Pagination{
				PageSize: 6,
			},
		})
		if err == nil && gamesResp != nil {
			p.RecentGames = gamesResp.Items
			p.TotalGames = gamesResp.Pagination.TotalResults
		}
	}

	// Fetch recent worlds (limit to 6)
	worldsClient, err := vc.ClientMgr.GetWorldsSvcClient()
	if err == nil {
		worldsResp, err := worldsClient.ListWorlds(context.Background(), &v1.ListWorldsRequest{
			Pagination: &v1.Pagination{
				PageSize: 6,
			},
		})
		if err == nil && worldsResp != nil {
			p.RecentWorlds = worldsResp.Items
			p.TotalWorlds = worldsResp.Pagination.TotalResults
		}
	}

	return
}

type PrivacyPolicy struct {
	Header Header
}

func (p *PrivacyPolicy) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	return p.Header.Load(r, w, vc)
}

type TermsOfService struct {
	Header Header
}

func (p *TermsOfService) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	return p.Header.Load(r, w, vc)
}

func (g *TermsOfService) Copy() View { return &TermsOfService{} }
func (g *PrivacyPolicy) Copy() View  { return &PrivacyPolicy{} }
func (g *HomePage) Copy() View       { return &HomePage{} }
