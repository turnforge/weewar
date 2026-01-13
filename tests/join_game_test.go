//go:build !wasm
// +build !wasm

package tests

import (
	"context"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/services"
)

// MockStorageProvider implements GameStorageProvider for testing
type MockStorageProvider struct {
	Games     map[string]*v1.Game
	States    map[string]*v1.GameState
	Histories map[string]*v1.GameMoveHistory
}

func NewMockStorageProvider() *MockStorageProvider {
	return &MockStorageProvider{
		Games:     make(map[string]*v1.Game),
		States:    make(map[string]*v1.GameState),
		Histories: make(map[string]*v1.GameMoveHistory),
	}
}

func (m *MockStorageProvider) LoadGame(ctx context.Context, id string) (*v1.Game, error) {
	if game, ok := m.Games[id]; ok {
		return game, nil
	}
	return nil, nil
}

func (m *MockStorageProvider) LoadGameState(ctx context.Context, id string) (*v1.GameState, error) {
	if state, ok := m.States[id]; ok {
		return state, nil
	}
	return nil, nil
}

func (m *MockStorageProvider) LoadGameHistory(ctx context.Context, id string) (*v1.GameMoveHistory, error) {
	if history, ok := m.Histories[id]; ok {
		return history, nil
	}
	return nil, nil
}

func (m *MockStorageProvider) SaveGame(ctx context.Context, id string, game *v1.Game) error {
	m.Games[id] = game
	return nil
}

func (m *MockStorageProvider) SaveGameState(ctx context.Context, id string, state *v1.GameState) error {
	m.States[id] = state
	return nil
}

func (m *MockStorageProvider) SaveGameHistory(ctx context.Context, id string, history *v1.GameMoveHistory) error {
	m.Histories[id] = history
	return nil
}

func (m *MockStorageProvider) SaveMoves(ctx context.Context, gameId string, group *v1.GameMoveGroup, currentGroupNumber int64) error {
	return nil
}

func (m *MockStorageProvider) DeleteFromStorage(ctx context.Context, id string) error {
	delete(m.Games, id)
	delete(m.States, id)
	delete(m.Histories, id)
	return nil
}

// createTestGame creates a game with the given player configuration
func createTestGame(gameId string, players []*v1.GamePlayer) *v1.Game {
	return &v1.Game{
		Id:   gameId,
		Name: "Test Game",
		Config: &v1.GameConfiguration{
			Players: players,
		},
	}
}

// createTestGameState creates a minimal game state
func createTestGameState() *v1.GameState {
	return &v1.GameState{
		CurrentPlayer: 1,
		TurnCounter:   1,
		WorldData: &v1.WorldData{
			TilesMap: map[string]*v1.Tile{
				"0,0": {Q: 0, R: 0, TileType: 1, Player: 1},
				"1,0": {Q: 1, R: 0, TileType: 1, Player: 2},
			},
			UnitsMap: map[string]*v1.Unit{
				"0,0": {Q: 0, R: 0, Player: 1, UnitType: 1},
				"1,0": {Q: 1, R: 0, Player: 2, UnitType: 1},
			},
		},
		PlayerStates: map[int32]*v1.PlayerState{
			1: {Coins: 100, IsActive: true},
			2: {Coins: 100, IsActive: true},
		},
	}
}

// TestJoinGame_Success tests successful joining of an open player slot
func TestJoinGame_Success(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	// Create a game with player 1 as human and player 2 as open
	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "human", UserId: "existing-user", Name: "Player 1"},
		{PlayerId: 2, PlayerType: "open", UserId: "", Name: "Player 2"},
	})
	mockStorage.Games["test-game"] = game
	mockStorage.States["test-game"] = createTestGameState()

	// Create the service
	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	// Use authenticated context with TestUserID
	ctx := ContextWithUserID(TestUserID)

	// Join as player 2
	resp, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 2,
	})

	if err != nil {
		t.Fatalf("JoinGame failed: %v", err)
	}

	if resp == nil {
		t.Fatal("JoinGame returned nil response")
	}

	if resp.PlayerId != 2 {
		t.Errorf("Expected PlayerId 2, got %d", resp.PlayerId)
	}

	// Verify the player was updated
	updatedGame := mockStorage.Games["test-game"]
	player2 := updatedGame.Config.Players[1]

	if player2.PlayerType != "human" {
		t.Errorf("Expected PlayerType 'human', got '%s'", player2.PlayerType)
	}

	if player2.UserId != TestUserID {
		t.Errorf("Expected UserId '%s', got '%s'", TestUserID, player2.UserId)
	}

	t.Logf("Player 2 successfully joined with UserId: %s", player2.UserId)
}

// TestJoinGame_SlotNotOpen tests that joining a non-open slot fails
func TestJoinGame_SlotNotOpen(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	// Create a game with both players as human (not open)
	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "human", UserId: "user1", Name: "Player 1"},
		{PlayerId: 2, PlayerType: "human", UserId: "user2", Name: "Player 2"},
	})
	mockStorage.Games["test-game"] = game
	mockStorage.States["test-game"] = createTestGameState()

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID("new-user")

	// Try to join as player 2 (already taken by human)
	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 2,
	})

	if err == nil {
		t.Fatal("Expected error when joining non-open slot, got nil")
	}

	expectedMsg := "not open for joining"
	if !containsSubstring(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}

	t.Logf("Correctly rejected: %v", err)
}

// TestJoinGame_AISlotNotJoinable tests that AI slots cannot be joined
func TestJoinGame_AISlotNotJoinable(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	// Create a game with player 2 as AI
	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "human", UserId: "user1", Name: "Player 1"},
		{PlayerId: 2, PlayerType: "ai", UserId: "", Name: "AI Player"},
	})
	mockStorage.Games["test-game"] = game
	mockStorage.States["test-game"] = createTestGameState()

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID("new-user")

	// Try to join as player 2 (AI slot)
	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 2,
	})

	if err == nil {
		t.Fatal("Expected error when joining AI slot, got nil")
	}

	expectedMsg := "not open for joining"
	if !containsSubstring(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}

	t.Logf("Correctly rejected AI slot join: %v", err)
}

// TestJoinGame_AlreadyPlayer tests that a user cannot join if already a player
func TestJoinGame_AlreadyPlayer(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	// Create a game where TestUserID is already player 1
	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "human", UserId: TestUserID, Name: "Player 1"},
		{PlayerId: 2, PlayerType: "open", UserId: "", Name: "Player 2"},
	})
	mockStorage.Games["test-game"] = game
	mockStorage.States["test-game"] = createTestGameState()

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	// Use same user ID that's already player 1
	ctx := ContextWithUserID(TestUserID)

	// Try to join as player 2 (user is already player 1)
	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 2,
	})

	if err == nil {
		t.Fatal("Expected error when user already a player, got nil")
	}

	expectedMsg := "you are already a player"
	if !containsSubstring(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}

	t.Logf("Correctly rejected duplicate join: %v", err)
}

// TestJoinGame_PlayerSlotNotFound tests joining a non-existent player slot
func TestJoinGame_PlayerSlotNotFound(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	// Create a game with only 2 players
	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "human", UserId: "user1", Name: "Player 1"},
		{PlayerId: 2, PlayerType: "open", UserId: "", Name: "Player 2"},
	})
	mockStorage.Games["test-game"] = game
	mockStorage.States["test-game"] = createTestGameState()

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID("new-user")

	// Try to join as player 3 (doesn't exist)
	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 3,
	})

	if err == nil {
		t.Fatal("Expected error when player slot doesn't exist, got nil")
	}

	expectedMsg := "not found"
	if !containsSubstring(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}

	t.Logf("Correctly rejected non-existent slot: %v", err)
}

// TestJoinGame_GameNotFound tests joining a non-existent game
func TestJoinGame_GameNotFound(t *testing.T) {
	mockStorage := NewMockStorageProvider()
	// Don't add any game to storage

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID(TestUserID)

	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "nonexistent-game",
		PlayerId: 1,
	})

	if err == nil {
		t.Fatal("Expected error when game doesn't exist, got nil")
	}

	t.Logf("Correctly rejected non-existent game: %v", err)
}

// TestJoinGame_InvalidPlayerId tests joining with invalid player ID
func TestJoinGame_InvalidPlayerId(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "open", UserId: "", Name: "Player 1"},
	})
	mockStorage.Games["test-game"] = game

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID(TestUserID)

	// Test with player ID 0
	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 0,
	})

	if err == nil {
		t.Fatal("Expected error for player ID 0, got nil")
	}

	// Test with negative player ID
	_, err = svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: -1,
	})

	if err == nil {
		t.Fatal("Expected error for negative player ID, got nil")
	}

	t.Log("Correctly rejected invalid player IDs")
}

// TestJoinGame_EmptyGameId tests joining with empty game ID
func TestJoinGame_EmptyGameId(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID(TestUserID)

	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "",
		PlayerId: 1,
	})

	if err == nil {
		t.Fatal("Expected error for empty game ID, got nil")
	}

	expectedMsg := "game ID is required"
	if !containsSubstring(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}

	t.Logf("Correctly rejected empty game ID: %v", err)
}

// TestJoinGame_MultipleOpenSlots tests joining when multiple slots are open
func TestJoinGame_MultipleOpenSlots(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	// Create a 4-player game with 3 open slots
	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "human", UserId: "creator", Name: "Player 1"},
		{PlayerId: 2, PlayerType: "open", UserId: "", Name: "Player 2"},
		{PlayerId: 3, PlayerType: "open", UserId: "", Name: "Player 3"},
		{PlayerId: 4, PlayerType: "open", UserId: "", Name: "Player 4"},
	})
	mockStorage.Games["test-game"] = game
	mockStorage.States["test-game"] = createTestGameState()

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	ctx := ContextWithUserID(TestUserID)

	// Join as player 3 (middle open slot)
	resp, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 3,
	})

	if err != nil {
		t.Fatalf("JoinGame failed: %v", err)
	}

	if resp.PlayerId != 3 {
		t.Errorf("Expected PlayerId 3, got %d", resp.PlayerId)
	}

	// Verify only player 3 was updated
	updatedGame := mockStorage.Games["test-game"]

	if updatedGame.Config.Players[1].PlayerType != "open" {
		t.Error("Player 2 should still be open")
	}

	if updatedGame.Config.Players[2].PlayerType != "human" {
		t.Error("Player 3 should now be human")
	}

	if updatedGame.Config.Players[3].PlayerType != "open" {
		t.Error("Player 4 should still be open")
	}

	t.Log("Successfully joined specific open slot among multiple")
}

// TestJoinGame_Unauthenticated tests that unauthenticated users cannot join
func TestJoinGame_Unauthenticated(t *testing.T) {
	mockStorage := NewMockStorageProvider()

	game := createTestGame("test-game", []*v1.GamePlayer{
		{PlayerId: 1, PlayerType: "open", UserId: "", Name: "Player 1"},
	})
	mockStorage.Games["test-game"] = game

	svc := &services.BackendGamesService{
		StorageProvider: mockStorage,
	}

	// Use context without user ID (unauthenticated)
	ctx := context.Background()

	_, err := svc.JoinGame(ctx, &v1.JoinGameRequest{
		GameId:   "test-game",
		PlayerId: 1,
	})

	if err == nil {
		t.Fatal("Expected error for unauthenticated user, got nil")
	}

	t.Logf("Correctly rejected unauthenticated join: %v", err)
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
