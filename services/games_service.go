package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var GAMES_STORAGE_DIR = weewar.DevDataPath("storage/games")

// GameMetadata represents the metadata stored in metadata.json
type GameMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatorID   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GameData represents the complete game data stored in data.json
type GameData struct {
	// World ID this game was created from
	WorldID string `json:"world_id"`
	// Starting world state (snapshot of the world when game was created)
	StartingWorld *v1.World `json:"starting_world"`
	// All moves made in this game
	Moves []*GameMoveEntry `json:"moves"`
}

// GameMoveEntry represents a single move entry in the game log
type GameMoveEntry struct {
	PlayerID  int32             `json:"player_id"`
	Timestamp time.Time         `json:"timestamp"`
	Move      *v1.GameMove      `json:"move"`
	Changes   []*v1.WorldChange `json:"changes"`
}

// GamesServiceImpl implements the GamesService gRPC interface
type GamesServiceImpl struct {
	v1.UnimplementedGamesServiceServer
	storageDir    string
	worldsService *WorldsServiceImpl
}

// NewGamesService creates a new GamesService implementation
func NewGamesService() *GamesServiceImpl {
	service := &GamesServiceImpl{
		storageDir:    GAMES_STORAGE_DIR,
		worldsService: NewWorldsService(),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(service.storageDir, 0755); err != nil {
		log.Printf("Failed to create games storage directory: %v", err)
	}

	return service
}

// ListGames returns all available games (metadata only for performance)
func (s *GamesServiceImpl) ListGames(ctx context.Context, req *v1.ListGamesRequest) (resp *v1.ListGamesResponse, err error) {
	resp = &v1.ListGamesResponse{
		Items: []*v1.Game{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}

	// Read all game directories
	entries, err := os.ReadDir(s.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Storage directory doesn't exist yet, return empty list
			return resp, nil
		}
		return nil, fmt.Errorf("failed to read games storage directory: %w", err)
	}

	var games []*v1.Game
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		gameID := entry.Name()
		metadata, err := s.loadGameMetadata(gameID)
		if err != nil {
			log.Printf("Failed to load metadata for game %s: %v", gameID, err)
			continue
		}

		// Filter by owner if specified
		if req.OwnerId != "" && metadata.CreatorID != req.OwnerId {
			continue
		}

		// Only return metadata for listing (not full game data)
		games = append(games, s.convertMetadataToProto(metadata))
	}

	resp.Items = games
	resp.Pagination.TotalResults = int32(len(games))

	return resp, nil
}

// GetGame returns a specific game with complete data including moves
func (s *GamesServiceImpl) GetGame(ctx context.Context, req *v1.GetGameRequest) (resp *v1.GetGameResponse, err error) {
	if req.Id == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	metadata, err := s.loadGameMetadata(req.Id)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	gameData, err := s.loadGameData(req.Id)
	if err != nil {
		// If data.json doesn't exist, create empty game data
		log.Printf("Game data not found for %s, creating empty data: %v", req.Id, err)
		gameData = &GameData{
			Moves: []*GameMoveEntry{},
		}
	}

	resp = &v1.GetGameResponse{
		Game: s.convertToFullProto(metadata, gameData),
	}

	return resp, nil
}

// CreateGame creates a new game from a world
func (s *GamesServiceImpl) CreateGame(ctx context.Context, req *v1.CreateGameRequest) (resp *v1.CreateGameResponse, err error) {
	if req.Game == nil {
		return nil, fmt.Errorf("game data is required")
	}

	// Get world_id from the proper field
	worldID := req.Game.WorldId
	if worldID == "" {
		return nil, fmt.Errorf("world_id is required to create a game")
	}

	var gameID string

	// Determine which game ID to use
	if req.Game.Id != "" {
		// Game ID provided - check if it's available
		exists, err := s.checkGameId(req.Game.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to check game ID: %w", err)
		}

		if exists {
			// Game ID is taken, generate a new one
			gameID, err = s.newGameId()
			if err != nil {
				return nil, fmt.Errorf("failed to generate new game ID: %w", err)
			}
		} else {
			// Game ID is available, use it
			gameID = req.Game.Id
		}
	} else {
		// No game ID provided, generate a new one
		gameID, err = s.newGameId()
		if err != nil {
			return nil, fmt.Errorf("failed to generate game ID: %w", err)
		}
	}

	// Load the world data first - fail if world doesn't exist
	worldResp, err := s.worldsService.GetWorld(ctx, &v1.GetWorldRequest{Id: worldID})
	if err != nil {
		return nil, fmt.Errorf("failed to load world %s: %w", worldID, err)
	}

	now := time.Now()
	metadata := &GameMetadata{
		ID:          gameID,
		Name:        req.Game.Name,
		Description: req.Game.Description,
		Tags:        req.Game.Tags,
		CreatorID:   req.Game.CreatorId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.saveGameMetadata(gameID, metadata); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Create game data with snapshot of the world
	gameData := &GameData{
		WorldID:       worldID,
		StartingWorld: worldResp.World,
		Moves:         []*GameMoveEntry{},
	}

	if err := s.saveGameData(gameID, gameData); err != nil {
		return nil, fmt.Errorf("failed to create game data: %w", err)
	}

	resp = &v1.CreateGameResponse{
		Game: s.convertToFullProto(metadata, gameData),
	}

	return resp, nil
}

// ProcessMoves processes moves for an existing game
func (s *GamesServiceImpl) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (resp *v1.ProcessMovesResponse, err error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}
	if len(req.Moves) == 0 {
		return nil, fmt.Errorf("at least one move is required")
	}

	// Load existing game data
	gameData, err := s.loadGameData(req.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to load game %s: %w", req.GameId, err)
	}

	// Process each move and collect results
	var moveResults []*v1.GameMoveResult

	// Add each move to the game log
	for _, move := range req.Moves {
		moveEntry := &GameMoveEntry{
			PlayerID:  move.Player,
			Timestamp: time.Now(),
			Move:      move,
			Changes:   []*v1.WorldChange{}, // TODO: Calculate actual changes using game engine
		}
		gameData.Moves = append(gameData.Moves, moveEntry)

		// TODO: Execute move and calculate result
		// For now, create a placeholder result
		moveResult := &v1.GameMoveResult{
			IsPermanent: false, // Most moves are non-permanent (can be undone)
			Changes:     moveEntry.Changes,
		}
		moveResults = append(moveResults, moveResult)
	}

	// Save updated game data
	if err := s.saveGameData(req.GameId, gameData); err != nil {
		return nil, fmt.Errorf("failed to save game data: %w", err)
	}

	resp = &v1.ProcessMovesResponse{
		MoveResults: moveResults,
	}

	return resp, nil
}

// checkGameId checks if a world ID already exists in storage
func (s *GamesServiceImpl) checkGameId(gameID string) (bool, error) {
	metadataPath := s.getMetadataPath(gameID)
	_, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Game doesn't exist - ID is available
			return false, nil
		}
		// Other file system error
		return false, fmt.Errorf("failed to check game ID %s: %w", gameID, err)
	}
	// Game exists - ID is taken
	return true, nil
}

// newGameId generates a new unique game ID of specified length (default 8 chars)
func (s *GamesServiceImpl) newGameId(numChars ...int) (string, error) {
	const maxRetries = 10

	// Default to 8 characters if not specified
	length := 8
	if len(numChars) > 0 && numChars[0] > 0 {
		length = numChars[0]
	}

	// Calculate number of bytes needed (2 hex chars per byte)
	numBytes := (length + 1) / 2

	for attempt := range maxRetries {
		// Generate random bytes
		bytes := make([]byte, numBytes)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}

		// Convert to hex string and truncate to exact length
		gameID := hex.EncodeToString(bytes)[:length]

		// Check if this ID is already taken
		exists, err := s.checkGameId(gameID)
		if err != nil {
			return "", fmt.Errorf("failed to check game ID uniqueness: %w", err)
		}

		if !exists {
			// Found a unique ID
			return gameID, nil
		}

		// ID collision, try again
		log.Printf("Game ID collision detected (attempt %d/%d): %s", attempt+1, maxRetries, gameID)
	}

	return "", fmt.Errorf("failed to generate unique game ID after %d attempts", maxRetries)
}

// getMetadataPath returns the metadata.json file path for a game
func (s *GamesServiceImpl) getMetadataPath(gameID string) string {
	return filepath.Join(s.getGamePath(gameID), "metadata.json")
}

// getDataPath returns the data.json file path for a game
func (s *GamesServiceImpl) getDataPath(gameID string) string {
	return filepath.Join(s.getGamePath(gameID), "data.json")
}

// loadGameMetadata loads metadata from metadata.json
func (s *GamesServiceImpl) loadGameMetadata(gameID string) (*GameMetadata, error) {
	metadataPath := s.getMetadataPath(gameID)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata for game %s: %w", gameID, err)
	}

	var metadata GameMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata for game %s: %w", gameID, err)
	}

	return &metadata, nil
}

// loadGameData loads complete game data from data.json
func (s *GamesServiceImpl) loadGameData(gameID string) (*GameData, error) {
	dataPath := s.getDataPath(gameID)
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read game data for game %s: %w", gameID, err)
	}

	var gameData GameData
	if err := json.Unmarshal(data, &gameData); err != nil {
		return nil, fmt.Errorf("failed to parse game data for game %s: %w", gameID, err)
	}

	return &gameData, nil
}

// saveGameMetadata saves metadata to metadata.json
func (s *GamesServiceImpl) saveGameMetadata(gameID string, metadata *GameMetadata) error {
	gameDir := s.getGamePath(gameID)
	if err := os.MkdirAll(gameDir, 0755); err != nil {
		return fmt.Errorf("failed to create game directory %s: %w", gameDir, err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for game %s: %w", gameID, err)
	}

	metadataPath := s.getMetadataPath(gameID)
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata for game %s: %w", gameID, err)
	}

	return nil
}

// saveGameData saves complete game data to data.json
func (s *GamesServiceImpl) saveGameData(gameID string, gameData *GameData) error {
	data, err := json.MarshalIndent(gameData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal game data for game %s: %w", gameID, err)
	}

	dataPath := s.getDataPath(gameID)
	if err := os.WriteFile(dataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write game data for game %s: %w", gameID, err)
	}

	return nil
}

// getGamePath returns the directory path for a game
func (s *GamesServiceImpl) getGamePath(gameID string) string {
	return filepath.Join(s.storageDir, gameID)
}

// convertMetadataToProto converts GameMetadata to protobuf Game (metadata only)
func (s *GamesServiceImpl) convertMetadataToProto(metadata *GameMetadata) *v1.Game {
	return &v1.Game{
		Id:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Tags:        metadata.Tags,
		CreatorId:   metadata.CreatorID,
		CreatedAt:   timestamppb.New(metadata.CreatedAt),
		UpdatedAt:   timestamppb.New(metadata.UpdatedAt),
	}
}

// convertToFullProto converts metadata + data to complete protobuf Game
func (s *GamesServiceImpl) convertToFullProto(metadata *GameMetadata, gameData *GameData) *v1.Game {
	protoGame := s.convertMetadataToProto(metadata)
	protoGame.WorldId = gameData.WorldID
	return protoGame
}
