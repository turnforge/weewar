package connectclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/gen/go/lilbattle/v1/services/lilbattlev1connect"
)

// authTransport wraps an http.RoundTripper to add Authorization headers
type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.base.RoundTrip(req)
}

// ConnectWorldsClient wraps a Connect client for the WorldsService
type ConnectWorldsClient struct {
	client lilbattlev1connect.WorldsServiceClient
}

// NewConnectWorldsClient creates a new Connect client for the WorldsService
func NewConnectWorldsClient(serverURL string) *ConnectWorldsClient {
	return NewConnectWorldsClientWithAuth(serverURL, "")
}

// NewConnectWorldsClientWithAuth creates a new Connect client with authentication
func NewConnectWorldsClientWithAuth(serverURL, token string) *ConnectWorldsClient {
	httpClient := http.DefaultClient
	if token != "" {
		httpClient = &http.Client{
			Transport: &authTransport{
				base:  http.DefaultTransport,
				token: token,
			},
		}
	}
	client := lilbattlev1connect.NewWorldsServiceClient(
		httpClient,
		serverURL,
	)
	return &ConnectWorldsClient{client: client}
}

// GetWorld returns a specific world with metadata via Connect
func (c *ConnectWorldsClient) GetWorld(ctx context.Context, req *v1.GetWorldRequest) (*v1.GetWorldResponse, error) {
	resp, err := c.client.GetWorld(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// ListWorlds returns all available worlds via Connect
func (c *ConnectWorldsClient) ListWorlds(ctx context.Context, req *v1.ListWorldsRequest) (*v1.ListWorldsResponse, error) {
	resp, err := c.client.ListWorlds(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// CreateWorld creates a new world via Connect
func (c *ConnectWorldsClient) CreateWorld(ctx context.Context, req *v1.CreateWorldRequest) (*v1.CreateWorldResponse, error) {
	resp, err := c.client.CreateWorld(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// UpdateWorld updates a world via Connect
func (c *ConnectWorldsClient) UpdateWorld(ctx context.Context, req *v1.UpdateWorldRequest) (*v1.UpdateWorldResponse, error) {
	resp, err := c.client.UpdateWorld(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// DeleteWorld deletes a world via Connect
func (c *ConnectWorldsClient) DeleteWorld(ctx context.Context, req *v1.DeleteWorldRequest) (*v1.DeleteWorldResponse, error) {
	resp, err := c.client.DeleteWorld(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// GetWorlds batch gets multiple worlds by ID via Connect
func (c *ConnectWorldsClient) GetWorlds(ctx context.Context, req *v1.GetWorldsRequest) (*v1.GetWorldsResponse, error) {
	resp, err := c.client.GetWorlds(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}
