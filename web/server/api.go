package server

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gfn "github.com/panyam/goutils/fn"
	oa "github.com/panyam/oneauth"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1/v1connect"
	"github.com/panyam/turnengine/games/weewar/services"
	svc "github.com/panyam/turnengine/games/weewar/services"
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
	ClientMgr      *svc.ClientMgr

	// Here we can have to ways of accessing the services - either via clients or by actual service instead if you are not
	// running the services on a dedicated port
}

func (n *ApiHandler) Handler() http.Handler {
	return n.mux
}

func NewApiHandler(middleware *oa.Middleware, clients *svc.ClientMgr) *ApiHandler {
	out := ApiHandler{
		mux:            http.NewServeMux(),
		AuthMiddleware: middleware,
		ClientMgr:      clients,
	}
	gwmux, err := out.createSvcMux(clients.Address())
	if err != nil {
		log.Println("error creating grpc mux: ", err)
		panic(err)
	}
	out.mux.Handle("/v1/", gwmux)
	log.Println("Registered gRPC-gateway at /v1/")

	// Add Connect handlers
	/*
		if canvasService != nil {
			log.Println("Adding Connect handler...")
			adapter := NewConnectCanvasServiceAdapter(canvasService)
			connectPath, connectHandler := v1connect.NewCanvasServiceHandler(adapter)
			out.mux.Handle(connectPath, connectHandler)
			log.Printf("Registered Connect handler at: %s", connectPath)
		} else {
			log.Println("No CanvasService provided, skipping Connect handler")
		}
	*/

	// Add AppItems Connect handler
	// We will do this for each service we have registered
	log.Println("Adding Games Connect handler...")
	gamesAdapter := NewConnectGamesServiceAdapter(services.NewGamesService())
	gamesConnectPath, gamesConnectHandler := v1connect.NewGamesServiceHandler(gamesAdapter)
	out.mux.Handle(gamesConnectPath, gamesConnectHandler)
	log.Printf("Registered Games Connect handler at: %s", gamesConnectPath)

	worldsAdapter := NewConnectWorldsServiceAdapter(services.NewWorldsService())
	worldsConnectPath, worldsConnectHandler := v1connect.NewWorldsServiceHandler(worldsAdapter)
	out.mux.Handle(worldsConnectPath, worldsConnectHandler)
	log.Printf("Registered Worlds Connect handler at: %s", worldsConnectPath)

	return &out
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
				Error   string        `json:"error"`
				Message string        `json:"message"`
				Code    int           `json:"code"` // gRPC code
				Details []interface{} `json:"details,omitempty"`
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

	// Register existing services
	err := v1.RegisterGamesServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register appitems service: ", err)
		return nil, err
	}
	err = v1.RegisterWorldsServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register appitems service: ", err)
		return nil, err
	}
	return svcMux, nil // Return nil error on success
}
