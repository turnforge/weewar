package server

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gfn "github.com/panyam/goutils/fn"
	oa "github.com/panyam/oneauth"
	"github.com/panyam/servicekit/grpcws"
	gohttp "github.com/panyam/servicekit/http"
	models "github.com/turnforge/weewar/gen/go/weewar/v1/models"
	v1s "github.com/turnforge/weewar/gen/go/weewar/v1/services"
	v1connect "github.com/turnforge/weewar/gen/go/weewar/v1/services/weewarv1connect"
	"github.com/turnforge/weewar/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type ApiHandler struct {
	mux            *http.ServeMux
	AuthMiddleware *oa.Middleware
	ClientMgr      *services.ClientMgr

	// Here we can have to ways of accessing the services - either via clients or by actual service instead if you are not
	// running the services on a dedicated port
	DisableIndexer       bool
	DisableGamesService  bool
	DisableWorldsService bool
}

func (n *ApiHandler) Handler() http.Handler {
	return n.mux
}

func (a *ApiHandler) Init() error {
	a.mux = http.NewServeMux()
	gwmux, err := a.createSvcMux(a.ClientMgr.Address())
	if err != nil {
		log.Println("error creating grpc mux: ", err)
		panic(err)
	}
	a.mux.Handle("/v1/", gwmux)
	log.Println("Registered gRPC-gateway at /v1/")

	// WebSocket endpoint for GameSyncService Subscribe using servicekit grpcws
	wsHandler := grpcws.NewServerStreamHandler(
		func(ctx context.Context, req *models.SubscribeRequest) (grpc.ServerStreamingClient[models.GameUpdate], error) {
			return a.ClientMgr.GetGameSyncSvcClient().Subscribe(ctx, req)
		},
		func(r *http.Request) (*models.SubscribeRequest, error) {
			gameId := r.PathValue("game_id")
			playerId := r.URL.Query().Get("player_id")
			fromSeq := int64(0)
			if fs := r.URL.Query().Get("from_sequence"); fs != "" {
				fromSeq, _ = strconv.ParseInt(fs, 10, 64)
			}
			return &models.SubscribeRequest{
				GameId:       gameId,
				PlayerId:     playerId,
				FromSequence: fromSeq,
			}, nil
		},
	)
	a.mux.HandleFunc("/ws/v1/sync/games/{game_id}/subscribe", gohttp.WSServe(wsHandler, nil))
	log.Println("Registered GameSync WebSocket handler at /ws/v1/sync/games/{game_id}/subscribe")

	return a.setupConnectHandlers()
}

func (out *ApiHandler) setupConnectHandlers() error {
	// Add AppItems Connect handler
	// We will do this for each service we have registered
	if !out.DisableGamesService {
		log.Println("Adding Games Connect handler...")
		gamesSvcClient := out.ClientMgr.GetGamesSvcClient()
		gamesAdapter := NewConnectGamesServiceAdapter(gamesSvcClient)
		gamesConnectPath, gamesConnectHandler := v1connect.NewGamesServiceHandler(gamesAdapter)
		out.mux.Handle(gamesConnectPath, gamesConnectHandler)
		log.Printf("Registered Games Connect handler at: %s", gamesConnectPath)
	}

	if !out.DisableWorldsService {
		worldsSvcClient := out.ClientMgr.GetWorldsSvcClient()
		worldsAdapter := NewConnectWorldsServiceAdapter(worldsSvcClient)
		worldsConnectPath, worldsConnectHandler := v1connect.NewWorldsServiceHandler(worldsAdapter)
		out.mux.Handle(worldsConnectPath, worldsConnectHandler)
		log.Printf("Registered Worlds Connect handler at: %s", worldsConnectPath)
	}

	// if we are colocating indexer in our current bundle
	if !out.DisableIndexer {
		/* - TODO we are creating a new service via NewIndexService - instead this pattern should be using the client.
		* Needs to be investigated
		log.Println("Adding Indexer Connect handler...")
		indexerSvcClient := out.ClientMgr.GetIndexerSvcClient()
		indexerAdapter := NewConnectIndexerServiceAdapter(indexerSvcClient)
		indexerConnectPath, indexerConnectHandler := v1connect.NewIndexerServiceHandler(indexerAdapter)
		out.mux.Handle(indexerConnectPath, indexerConnectHandler)
		log.Printf("Registered Indexer Connect handler at: %s", indexerConnectPath)
		*/
	}

	// Register GameSyncService for multiplayer real-time updates
	gameSyncSvcClient := out.ClientMgr.GetGameSyncSvcClient()
	gameSyncAdapter := NewConnectGameSyncServiceAdapter(gameSyncSvcClient)
	gameSyncConnectPath, gameSyncConnectHandler := v1connect.NewGameSyncServiceHandler(gameSyncAdapter)
	out.mux.Handle(gameSyncConnectPath, gameSyncConnectHandler)
	log.Printf("Registered GameSync Connect handler at: %s", gameSyncConnectPath)

	return nil
}

func (web *ApiHandler) createSvcMux(grpc_addr string) (*runtime.ServeMux, error) {
	svcMux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			// metadata.AppendToOutgoingContext(ctx)
			md := metadata.Pairs()
			return md
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			// Custom Error Handling: Convert gRPC status to HTTP status
			s := status.Convert(err)
			httpStatus := runtime.HTTPStatusFromCode(s.Code())

			// Log the error with details
			log.Printf("gRPC Gateway Error: code=%s, http_status=%d, msg=%s, details=%v\n", s.Code(), httpStatus, s.Message(), s.Details())

			// Prepare response body
			body := struct {
				Error   string `json:"error"`
				Message string `json:"message"`
				Code    int    `json:"code"` // gRPC code
				Details []any  `json:"details,omitempty"`
			}{
				Error:   s.Code().String(),
				Message: s.Message(),
				Code:    int(s.Code()),
				Details: gfn.Map(s.Proto().Details, func(detail *anypb.Any) any {
					var msg proto.Message
					msg, err = anypb.UnmarshalNew(detail, proto.UnmarshalOptions{})
					if err != nil {
						// Attempt to convert the known proto message to a world
						// This might need a custom function depending on the marshaler
						// For standard JSON, structpb.NewStruct might work if it was a struct
						// For simplicity, let's just use the detail itself for now.
						log.Printf("Failed to unmarshal error detail: %v", err)
					}
					return msg
				}),
			}

			// Set headers and write response
			writer.Header().Del("Trailer") // Important: Remove Trailer header
			writer.Header().Set("Content-Type", marshaler.ContentType(body))
			writer.WriteHeader(httpStatus)
			if err := marshaler.NewEncoder(writer).Encode(body); err != nil {
				log.Printf("Failed to marshal error response: %v", err)
				// Fallback to DefaultErrorHandler if marshaling fails
				runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, err)
			}
		}),
	)

	// TODO - Secure credentials for etc
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	ctx := context.Background()
	var err error

	// Register existing services
	if !web.DisableGamesService {
		err := v1s.RegisterGamesServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
		if err != nil {
			log.Fatal("Unable to register games service: ", err)
			return nil, err
		}
	}
	if !web.DisableWorldsService {
		err = v1s.RegisterWorldsServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
		if err != nil {
			log.Fatal("Unable to register worlds service: ", err)
			return nil, err
		}
	}

	if web.DisableIndexer {
		err = v1s.RegisterIndexerServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
		if err != nil {
			log.Fatal("Unable to register indexer service: ", err)
			return nil, err
		}
	}

	// Register GameSyncService via grpc-gateway for Broadcast endpoint
	// (Subscribe streaming is also available via SSE at /v1/..., but prefer WebSocket at /ws/...)
	err = v1s.RegisterGameSyncServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register game sync service: ", err)
		return nil, err
	}

	return svcMux, nil // Return nil error on success
}
