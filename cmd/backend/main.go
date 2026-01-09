package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/joho/godotenv"
	goal "github.com/panyam/goapplib"
	v1s "github.com/turnforge/weewar/gen/go/weewar/v1/services"
	"github.com/turnforge/weewar/services"
	"github.com/turnforge/weewar/services/fsbe"
	"github.com/turnforge/weewar/services/gaebe"
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
	worlds_service_be = flag.String("worlds_service_be", "", "Storage for worlds service - 'local', 'pg', 'gae'. Env: WORLDS_SERVICE_BE. Default: pg")
	games_service_be  = flag.String("games_service_be", "", "Storage for games service - 'local', 'pg', 'gae'. Env: GAMES_SERVICE_BE. Default: pg")
	filestore_be      = flag.String("filestore_be", "", "Storage for filestore - 'local', 'r2', 'gae'. Env: FILESTORE_BE. Default: local")
	gae_project       = flag.String("gae_project", "", "Google Cloud project ID for GAE/Datastore. Env: GAE_PROJECT")
	gae_namespace     = flag.String("gae_namespace", "", "Datastore namespace (optional, for multi-tenancy). Env: GAE_NAMESPACE")
)

// getBackendConfig returns the backend configuration value with priority:
// command line flag -> environment variable -> default value
func getBackendConfig(flagValue *string, envVar string, defaultValue string) string {
	// If flag was explicitly set (non-empty), use it
	if flagValue != nil && *flagValue != "" {
		return *flagValue
	}
	// Check environment variable
	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}
	// Return default
	return defaultValue
}

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
	// Default to dev mode, use WEEWAR_ENV=production for production
	envfile := "configs/.env.dev"
	weewarEnv := os.Getenv("WEEWAR_ENV")
	log.Println("Environment: ", weewarEnv)
	if weewarEnv == "production" {
		envfile = "configs/.env"
	} else {
		// Dev mode - enable debug logging
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

		// Get backend configurations with priority: flag -> env var -> default
		worldsBE := getBackendConfig(worlds_service_be, "WORLDS_SERVICE_BE", "pg")
		gamesBE := getBackendConfig(games_service_be, "GAMES_SERVICE_BE", "pg")
		filestoreBE := getBackendConfig(filestore_be, "FILESTORE_BE", "local")

		log.Printf("Backend configuration: worlds=%s, games=%s, filestore=%s", worldsBE, gamesBE, filestoreBE)

		var db *gorm.DB = nil
		ensureDB := func() *gorm.DB {
			if db == nil {
				db = gormbe.OpenWeewarDB(*db_endpoint, DEFAULT_DB_ENDPOINT)
			}
			return db
		}

		var dsClient *datastore.Client = nil
		ensureDatastore := func() *datastore.Client {
			if dsClient == nil {
				// Check GAE_PROJECT first, then GOOGLE_CLOUD_PROJECT (set by App Engine)
				projectID := getBackendConfig(gae_project, "GAE_PROJECT", "")
				if projectID == "" {
					projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
				}
				if projectID == "" {
					panic("GAE_PROJECT or GOOGLE_CLOUD_PROJECT environment variable (or --gae_project flag) is required for GAE backend")
				}
				var err error
				dsClient, err = datastore.NewClient(context.Background(), projectID)
				if err != nil {
					panic(fmt.Sprintf("Failed to create Datastore client: %v", err))
				}
				log.Printf("Datastore client initialized for project: %s", projectID)
			}
			return dsClient
		}
		dsNamespace := getBackendConfig(gae_namespace, "GAE_NAMESPACE", "")

		// v1s.RegisterWorldsServiceServer(server, fsbe.NewFSWorldsService(""))
		switch worldsBE {
		case "pg":
			worldsService = gormbe.NewWorldsService(ensureDB(), clientMgr)
		case "local":
			worldsService = fsbe.NewFSWorldsService("", clientMgr)
		case "gae":
			worldsService = gaebe.NewWorldsService(ensureDatastore(), dsNamespace, clientMgr)
		default:
			panic("Invalid worlds_service_be: " + worldsBE + ". Valid options: local, pg, gae")
		}

		switch gamesBE {
		case "local":
			gamesService = fsbe.NewFSGamesService("", clientMgr)
		case "pg":
			gamesService = gormbe.NewGamesService(ensureDB(), clientMgr)
		case "gae":
			gamesService = gaebe.NewGamesService(ensureDatastore(), dsNamespace, clientMgr)
		default:
			panic("Invalid games_service_be: " + gamesBE + ". Valid options: local, pg, gae")
		}

		switch filestoreBE {
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
		case "gae":
			panic("GAE/Datastore backend not yet implemented for filestore")
		default:
			panic("Invalid filestore_be: " + filestoreBE + ". Valid options: local, r2, gae")
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
