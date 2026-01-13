//go:build !wasm
// +build !wasm

package gaebe

import (
	"context"
	"errors"
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

const name = "github.com/turnforge/lilbattle/gaebe"

var (
	Tracer = otel.Tracer(name)
	Meter  = otel.Meter(name)
	Logger = otelslog.NewLogger(name)
)

var InvalidIDError = errors.New("ID is invalid or empty")
var EntityNotFoundError = errors.New("entity not found")
var VersionMismatchError = errors.New("version mismatch - content was updated")

// Config holds Datastore client configuration
type Config struct {
	ProjectID string
	Namespace string // Optional: for multi-tenant isolation
}

// NewClient creates a Datastore client from configuration
func NewClient(ctx context.Context, cfg Config) (*datastore.Client, error) {
	projectID := cfg.ProjectID
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		projectID = os.Getenv("DATASTORE_PROJECT_ID")
	}
	if projectID == "" {
		projectID = os.Getenv("GAE_PROJECT")
	}

	log.Printf("Connecting to Datastore: project=%s, namespace=%s", projectID, cfg.Namespace)

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Datastore client: %v", err)
		return nil, err
	}

	log.Printf("Successfully connected to Datastore: project=%s", projectID)
	return client, nil
}

// OpenDatastore creates a Datastore client with fallback to environment variables
func OpenDatastore(projectID string, defaultProject string) *datastore.Client {
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			projectID = os.Getenv("DATASTORE_PROJECT_ID")
		}
		if projectID == "" {
			projectID = os.Getenv("GAE_PROJECT")
		}
		if projectID == "" {
			projectID = defaultProject
		}
	}

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return client
}

// NamespacedKey creates a key with optional namespace
func NamespacedKey(kind, id string, namespace string) *datastore.Key {
	key := datastore.NameKey(kind, id, nil)
	if namespace != "" {
		key.Namespace = namespace
	}
	return key
}

// NamespacedQuery creates a query with optional namespace
func NamespacedQuery(kind string, namespace string) *datastore.Query {
	q := datastore.NewQuery(kind)
	if namespace != "" {
		q = q.Namespace(namespace)
	}
	return q
}
