package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	goal "github.com/panyam/goapplib"
	v1s "github.com/turnforge/weewar/gen/go/weewar/v1/services"
	"github.com/turnforge/weewar/services"
	"github.com/turnforge/weewar/services/fsbe"
	"github.com/turnforge/weewar/services/gormbe"
	"github.com/turnforge/weewar/services/r2"
	"github.com/turnforge/weewar/services/server"
	"github.com/turnforge/weewar/utils"
	web "github.com/turnforge/weewar/web/server"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

const DEFAULT_DB_ENDPOINT = "postgres://postgres:password@localhost:5432/weewardb"

var (
	grpcAddress       = flag.String("grpcAddress", DefaultServiceAddress(), "Address where the gRPC endpoint is running")
	gatewayAddress    = flag.String("gatewayAddress", DefaultGatewayAddress(), "Address where the http grpc gateway endpoint is running")
	db_endpoint       = flag.String("db_endpoint", "", fmt.Sprintf("Endpoint of DB where all data is persisted.  Default value: WEEWAR_DB_ENDPOINT environment variable or %s", DEFAULT_DB_ENDPOINT))
	worlds_service_be = flag.String("worlds_service_be", "pg", "Storage for worlds service - 'local', 'pg', 'datastore'.  ")
	games_service_be  = flag.String("games_service_be", "pg", "Storage for games service - 'local', 'pg', 'datastore'.  ")
	filestore_be      = flag.String("filestore_be", "local", "Storage for filestore - 'r2' or 'local', 'pg', 'datastore'.  ")
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
	clientMgr := services.NewClientMgr(b.GrpcAddress)
	grpcServer.RegisterCallback = func(server *grpc.Server) error {
		var gamesService v1s.GamesServiceServer
		var worldsService v1s.WorldsServiceServer
		var filestore v1s.FileStoreServiceServer

		var db *gorm.DB = nil
		ensureDB := func() *gorm.DB {
			if db == nil {
				db = gormbe.OpenWeewarDB(*db_endpoint, DEFAULT_DB_ENDPOINT)
			}
			return db
		}
		// v1s.RegisterWorldsServiceServer(server, fsbe.NewFSWorldsService(""))
		switch *worlds_service_be {
		case "pg":
			worldsService = gormbe.NewWorldsService(ensureDB(), clientMgr)
		case "local":
			worldsService = fsbe.NewFSWorldsService("", clientMgr)
		default:
			panic("Invalid world service be: " + *worlds_service_be)
		}

		switch *games_service_be {
		case "local":
			gamesService = fsbe.NewFSGamesService("", clientMgr)
		case "pg":
			gamesService = gormbe.NewGamesService(ensureDB(), clientMgr)
		default:
			panic("Invalid game service be: " + *games_service_be)
		}

		switch *filestore_be {
		case "local":
			filestore = fsbe.NewFileStoreService("", clientMgr)
		case "r2":
			r2Client, err := r2.NewR2Client(r2.R2Config{
				AccountID:       os.Getenv("R2_ACCOUNT_ID"),
				AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
				Bucket:          "weewar-assets",
				PublicURL:       "", // Leave empty for private bucket
			})
			if err != nil {
				panic(fmt.Sprintf("Could not instantiate r2 client: %v", err))
			}
			filestore = r2.NewR2FileStoreService(r2Client)
		default:
			panic("Invalid filestore be: " + *filestore_be)
		}

		// Create sync service for multiplayer real-time updates
		syncService := services.NewGameSyncService()

		v1s.RegisterWorldsServiceServer(server, worldsService)
		v1s.RegisterGamesServiceServer(server, gamesService)
		v1s.RegisterFileStoreServiceServer(server, filestore)
		v1s.RegisterGameSyncServiceServer(server, syncService)

		// TODO - use diferent kinds of db based on setup
		// v1s.RegisterIndexerServiceServer(server, gormbe.NewIndexerService(ensureDB()))
		return nil
	}
	app.AddServer(grpcServer)

	isDevMode := os.Getenv("WEEWAR_ENV") == "dev"
	app.AddServer(&web.WebAppServer{
		WebAppServer: goal.WebAppServer{
			GrpcAddress:   b.GrpcAddress,
			Address:       b.GatewayAddress,
			AllowLocalDev: isDevMode,
		},
	})
	b.App = app
	return app
}
