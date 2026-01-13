package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/turnforge/lilbattle/services/server"
	"github.com/turnforge/lilbattle/utils"
	web "github.com/turnforge/lilbattle/web/server"
)

var (
	grpcAddress    = flag.String("grpcAddress", DefaultServiceAddress(), "Address where the gRPC endpoint is running")
	gatewayAddress = flag.String("gatewayAddress", DefaultGatewayAddress(), "Address where the http grpc gateway endpoint is running")
)

type Backend struct {
	GrpcAddress    string
	GatewayAddress string
	App            *utils.App
}

// Sample main file for starting the backend

func main() {
	parseFlags()

	backend := Backend{GrpcAddress: *grpcAddress, GatewayAddress: *gatewayAddress}
	backend.SetupApp()
	backend.Start()
}

func DefaultGatewayAddress() string {
	gateway_addr := os.Getenv("LILBATTLE_INDEXER_WEB_PORT")
	if gateway_addr != "" {
		return gateway_addr
	}
	return ":6060"
}

func DefaultServiceAddress() string {
	port := os.Getenv("LILBATTLE_INDEXER_GRPC_PORT")
	if port != "" {
		return port
	}
	return ":7070"
}

func parseFlags() {
	envfile := ".env"
	log.Println("Environment: ", os.Getenv("LILBATTLE_INDEXER_ENV"))
	if os.Getenv("WEEAR_INDEXER_ENV") == "dev" {
		envfile = ".env.dev"
		logger := slog.New(utils.NewPrettyHandler(os.Stdout, utils.PrettyHandlerOptions{
			SlogOpts: slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}))
		slog.SetDefault(logger)
	}
	log.Println("loading env file: ", envfile)
	err := godotenv.Load(envfile)
	if err != nil {
		log.Fatal("Error loading .env file", envfile, err)
	}
	flag.Parse()
}

func (b *Backend) Start() {
	b.App.Start()
	b.App.Done(nil)
}

func (b *Backend) SetupApp() *utils.App {
	// this is the bit you wol
	app := &utils.App{Ctx: context.Background()}
	log.Println("Grpc, Address: ", grpcAddress)
	log.Println("gateway, Address: ", gatewayAddress)
	app.AddServer(&server.Server{Address: b.GrpcAddress})

	isDevMode := os.Getenv("WEEAR_INDEXER_ENV") == "dev"
	app.AddServer(&web.IndexerAppServer{
		GrpcAddress:   b.GrpcAddress,
		Address:       b.GatewayAddress,
		AllowLocalDev: isDevMode,
	})
	b.App = app
	return app
}
