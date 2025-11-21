package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	v1s "github.com/turnforge/weewar/gen/go/weewar/v1/services"
	"github.com/turnforge/weewar/services/fsbe"
	"github.com/turnforge/weewar/services/gormbe"
	"github.com/turnforge/weewar/services/server"
	"github.com/turnforge/weewar/utils"
	web "github.com/turnforge/weewar/web/server"
	"google.golang.org/grpc"
)

const DEFAULT_DB_ENDPOINT = "postgres://postgres:password@localhost:5432/weewardb"

var (
	grpcAddress    = flag.String("grpcAddress", DefaultServiceAddress(), "Address where the gRPC endpoint is running")
	gatewayAddress = flag.String("gatewayAddress", DefaultGatewayAddress(), "Address where the http grpc gateway endpoint is running")
	db_endpoint    = flag.String("db_endpoint", "", fmt.Sprintf("Endpoint of DB where all data is persisted.  Default value: WEEWAR_DB_ENDPOINT environment variable or %s", DEFAULT_DB_ENDPOINT))
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
	gateway_addr := os.Getenv("WEEWAR_WEB_PORT")
	if gateway_addr != "" {
		return gateway_addr
	}
	return ":8080"
}

func DefaultServiceAddress() string {
	port := os.Getenv("WEEWAR_GRPC_PORT")
	if port != "" {
		return port
	}
	return ":9090"
}

func parseFlags() {
	envfile := ".env"
	log.Println("Environment: ", os.Getenv("WEEWAR_ENV"))
	if os.Getenv("WEEWAR_ENV") == "dev" {
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
		log.Fatal("Error loading .env file: ", envfile, err)
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
	grpcServer := &server.Server{Address: b.GrpcAddress}
	grpcServer.RegisterCallback = func(server *grpc.Server) error {
		gamesService := fsbe.NewFSGamesService("")
		v1s.RegisterGamesServiceServer(server, gamesService)

		db := gormbe.OpenWeewarDB(*db_endpoint, DEFAULT_DB_ENDPOINT)
		// v1s.RegisterWorldsServiceServer(server, fsbe.NewFSWorldsService(""))
		v1s.RegisterWorldsServiceServer(server, gormbe.NewWorldsService(db))

		// TODO - use diferent kinds of db based on setup
		v1s.RegisterIndexerServiceServer(server, gormbe.NewIndexerService(db))
		return nil
	}
	app.AddServer(grpcServer)

	isDevMode := os.Getenv("WEEWAR_ENV") == "dev"
	app.AddServer(&web.WebAppServer{
		WebAppServer: utils.WebAppServer{
			GrpcAddress:   b.GrpcAddress,
			Address:       b.GatewayAddress,
			AllowLocalDev: isDevMode,
		},
	})
	b.App = app
	return app
}
