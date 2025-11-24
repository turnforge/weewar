package server

import (
	"context"

	"github.com/turnforge/weewar/services"
	"github.com/turnforge/weewar/utils"
)

type WebAppServer struct {
	utils.WebAppServer
}

// func (s *WebAppServer) Start(ctx context.Context, mux http.Handler, gw_addr string, srvErr chan error, stopChan chan bool) {
func (s *WebAppServer) Start(ctx context.Context, srvErr chan error, stopChan chan bool) error {
	cm := services.NewClientMgr(s.GrpcAddress)
	app, _ := NewApp(cm)
	return s.StartWithHandler(ctx, app.Handler(), srvErr, stopChan)
}

type IndexerAppServer struct {
	utils.WebAppServer
}

/*
// func (s *WebAppServer) Start(ctx context.Context, mux http.Handler, gw_addr string, srvErr chan error, stopChan chan bool) {
func (s *IndexerAppServer) Start(ctx context.Context, srvErr chan error, stopChan chan bool) error {
	cm := server.NewClientMgr(s.GrpcAddress)
	app, _ := NewIndexerApp(cm)
	return s.StartWithHandler(ctx, app.Handler(), srvErr, stopChan)
}
*/
