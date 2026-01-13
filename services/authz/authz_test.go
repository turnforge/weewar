//go:build !wasm
// +build !wasm

package authz

import (
	"context"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"google.golang.org/grpc/metadata"
)

// contextWithUserID creates a context with the user ID set in gRPC metadata.
// This simulates what the auth interceptor does in production.
// Uses "x-user-id" which is oagrpc.DefaultMetadataKeyUserID from oneauth.
func contextWithUserID(userID string) context.Context {
	md := metadata.Pairs("x-user-id", userID)
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestRequireAuthenticated_NoUser(t *testing.T) {
	ctx := context.Background()

	_, err := RequireAuthenticated(ctx)
	if err != ErrUnauthenticated {
		t.Errorf("Expected ErrUnauthenticated, got %v", err)
	}
}

func TestRequireAuthenticated_WithUser(t *testing.T) {
	ctx := contextWithUserID("user123")

	userID, err := RequireAuthenticated(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if userID != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", userID)
	}
}

func TestRequireOwnership_NotOwner(t *testing.T) {
	ctx := contextWithUserID("user123")

	err := RequireOwnership(ctx, "user456")
	if err != ErrNotOwner {
		t.Errorf("Expected ErrNotOwner, got %v", err)
	}
}

func TestRequireOwnership_IsOwner(t *testing.T) {
	ctx := contextWithUserID("user123")

	err := RequireOwnership(ctx, "user123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRequireGamePlayer_NoConfig(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id:     "game1",
		Config: nil,
	}

	_, err := RequireGamePlayer(ctx, game)
	if err != ErrNotPlayer {
		t.Errorf("Expected ErrNotPlayer, got %v", err)
	}
}

func TestRequireGamePlayer_NotAPlayer(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user789"},
			},
		},
	}

	_, err := RequireGamePlayer(ctx, game)
	if err != ErrNotPlayer {
		t.Errorf("Expected ErrNotPlayer, got %v", err)
	}
}

func TestRequireGamePlayer_IsPlayer(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user123"},
			},
		},
	}

	playerID, err := RequireGamePlayer(ctx, game)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if playerID != 2 {
		t.Errorf("Expected playerID 2, got %d", playerID)
	}
}

func TestRequireCurrentPlayer_NotYourTurn(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user123"},
			},
		},
	}

	// User123 is player 2, but current player is 1
	_, err := RequireCurrentPlayer(ctx, game, 1)
	if err != ErrNotYourTurn {
		t.Errorf("Expected ErrNotYourTurn, got %v", err)
	}
}

func TestRequireCurrentPlayer_IsYourTurn(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user123"},
			},
		},
	}

	// User123 is player 2, and current player is 2
	playerID, err := RequireCurrentPlayer(ctx, game, 2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if playerID != 2 {
		t.Errorf("Expected playerID 2, got %d", playerID)
	}
}

func TestCanSubmitMoves_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
			},
		},
	}

	err := CanSubmitMoves(ctx, game, 1)
	if err != ErrUnauthenticated {
		t.Errorf("Expected ErrUnauthenticated, got %v", err)
	}
}

func TestCanSubmitMoves_NotAPlayer(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user789"},
			},
		},
	}

	err := CanSubmitMoves(ctx, game, 1)
	if err != ErrNotPlayer {
		t.Errorf("Expected ErrNotPlayer, got %v", err)
	}
}

func TestCanSubmitMoves_NotYourTurn(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user123"},
			},
		},
	}

	// User123 is player 2, but current player is 1
	err := CanSubmitMoves(ctx, game, 1)
	if err != ErrNotYourTurn {
		t.Errorf("Expected ErrNotYourTurn, got %v", err)
	}
}

func TestCanSubmitMoves_Success(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id: "game1",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, UserId: "user456"},
				{PlayerId: 2, UserId: "user123"},
			},
		},
	}

	// User123 is player 2, and current player is 2
	err := CanSubmitMoves(ctx, game, 2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCanModifyGame_NotOwner(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id:        "game1",
		CreatorId: "user456",
	}

	err := CanModifyGame(ctx, game)
	if err != ErrNotOwner {
		t.Errorf("Expected ErrNotOwner, got %v", err)
	}
}

func TestCanModifyGame_IsOwner(t *testing.T) {
	ctx := contextWithUserID("user123")
	game := &v1.Game{
		Id:        "game1",
		CreatorId: "user123",
	}

	err := CanModifyGame(ctx, game)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCanModifyWorld_NotOwner(t *testing.T) {
	ctx := contextWithUserID("user123")
	world := &v1.World{
		Id:        "world1",
		CreatorId: "user456",
	}

	err := CanModifyWorld(ctx, world)
	if err != ErrNotOwner {
		t.Errorf("Expected ErrNotOwner, got %v", err)
	}
}

func TestCanModifyWorld_IsOwner(t *testing.T) {
	ctx := contextWithUserID("user123")
	world := &v1.World{
		Id:        "world1",
		CreatorId: "user123",
	}

	err := CanModifyWorld(ctx, world)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
