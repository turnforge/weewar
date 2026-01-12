package server

import (
	"context"

	goal "github.com/panyam/goapplib"
	"github.com/turnforge/lilbattle/services"
)

type WebAppServer struct {
	goal.WebAppServer
}

func (s *WebAppServer) Start(ctx context.Context, srvErr chan error, stopChan chan bool) error {
	cm := services.NewClientMgr(s.GrpcAddress)
	lilbattleApp, _, _ := NewLilBattleApp(cm)
	return s.StartWithHandler(ctx, lilbattleApp.Handler(), srvErr, stopChan)
}

type IndexerAppServer struct {
	goal.WebAppServer
}
