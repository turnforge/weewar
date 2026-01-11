//go:build wasm
// +build wasm

// Package authz provides authorization utilities for WeeWar services.
// This is the WASM stub - authorization is not applicable in browser context.
package authz

import (
	"context"
	"fmt"

	v1 "github.com/turnforge/weewar/gen/go/weewar/v1/models"
)

// Common authorization errors
var (
	ErrUnauthenticated = fmt.Errorf("authentication required")
	ErrForbidden       = fmt.Errorf("access denied")
	ErrNotOwner        = fmt.Errorf("you are not the owner of this resource")
	ErrNotPlayer       = fmt.Errorf("you are not a player in this game")
	ErrNotYourTurn     = fmt.Errorf("it is not your turn")
)

// GetUserIDFromContext returns empty string in WASM context.
func GetUserIDFromContext(ctx context.Context) string {
	return ""
}

// RequireAuthenticated always succeeds in WASM context.
func RequireAuthenticated(ctx context.Context) (string, error) {
	return "wasm-user", nil
}

// RequireOwnership always succeeds in WASM context.
func RequireOwnership(ctx context.Context, creatorID string) error {
	return nil
}

// CanModifyGame always succeeds in WASM context.
func CanModifyGame(ctx context.Context, game *v1.Game) error {
	return nil
}

// CanModifyWorld always succeeds in WASM context.
func CanModifyWorld(ctx context.Context, world *v1.World) error {
	return nil
}

// RequireGamePlayer returns player 1 in WASM context.
func RequireGamePlayer(ctx context.Context, game *v1.Game) (int32, error) {
	return 1, nil
}

// RequireCurrentPlayer returns player 1 in WASM context.
func RequireCurrentPlayer(ctx context.Context, game *v1.Game, currentPlayer int32) (int32, error) {
	return currentPlayer, nil
}

// CanSubmitMoves always succeeds in WASM context.
func CanSubmitMoves(ctx context.Context, game *v1.Game, currentPlayer int32) error {
	return nil
}
