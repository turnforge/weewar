//go:build !wasm
// +build !wasm

package fsbe

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/panyam/goutils/storage"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var GAMES_STORAGE_DIR = ""

// FSGamesService implements the GamesService gRPC interface
type FSGamesService struct {
	services.BackendGamesService
	storage *storage.FileStorage // Storage area for all files
}

// NewGamesService creates a new GamesService implementation for server mode
func NewFSGamesService(storageDir string, clientMgr *services.ClientMgr) *FSGamesService {
	if storageDir == "" {
		if GAMES_STORAGE_DIR == "" {
			GAMES_STORAGE_DIR = DevDataPath("storage/games")
		}
		storageDir = GAMES_STORAGE_DIR
	}
	service := &FSGamesService{
		storage: storage.NewFileStorage(storageDir),
	}
	service.ClientMgr = clientMgr
	service.Self = service
	service.StorageProvider = service // FSGamesService implements GameStorageProvider
	service.GameStateUpdater = service
	service.InitializeCache() // Initialize cache at BackendGamesService level
	service.InitializeScreenshotIndexer()
	service.InitializeSyncBroadcast()

	return service
}

// LoadGame implements GameStorageProvider - loads game directly from file storage
func (s *FSGamesService) LoadGame(ctx context.Context, id string) (*v1.Game, error) {
	game, err := storage.LoadFSArtifact[*v1.Game](s.storage, id, "metadata")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "game %s not found", id)
		}
		return nil, fmt.Errorf("failed to load game metadata: %w", err)
	}
	// Populate screenshot URL if not set
	if len(game.PreviewUrls) == 0 {
		game.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", game.Id)}
	}
	return game, nil
}

// LoadGameState implements GameStorageProvider - loads game state directly from file storage
func (s *FSGamesService) LoadGameState(ctx context.Context, id string) (*v1.GameState, error) {
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, id, "state")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "game state for %s not found", id)
		}
		return nil, fmt.Errorf("failed to load game state: %w", err)
	}
	return gameState, nil
}

// LoadGameHistory implements GameStorageProvider - loads game history directly from file storage
func (s *FSGamesService) LoadGameHistory(ctx context.Context, id string) (*v1.GameMoveHistory, error) {
	gameHistory, err := storage.LoadFSArtifact[*v1.GameMoveHistory](s.storage, id, "history")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "game history for %s not found", id)
		}
		return nil, fmt.Errorf("failed to load game history: %w", err)
	}
	return gameHistory, nil
}

// SaveGame implements GameStorageProvider - saves game metadata to file storage
func (s *FSGamesService) SaveGame(ctx context.Context, id string, game *v1.Game) error {
	return s.storage.SaveArtifact(id, "metadata", game)
}

// SaveGameState implements GameStorageProvider - saves game state to file storage
func (s *FSGamesService) SaveGameState(ctx context.Context, id string, state *v1.GameState) error {
	return s.storage.SaveArtifact(id, "state", state)
}

// SaveGameHistory implements GameStorageProvider - saves game history to file storage
func (s *FSGamesService) SaveGameHistory(ctx context.Context, id string, history *v1.GameMoveHistory) error {
	return s.storage.SaveArtifact(id, "history", history)
}

// DeleteFromStorage implements GameStorageProvider - deletes game from file storage
func (s *FSGamesService) DeleteFromStorage(ctx context.Context, id string) error {
	return s.storage.DeleteEntity(id)
}

// SaveMoves implements GameStorageProvider - appends moves to history file
func (s *FSGamesService) SaveMoves(ctx context.Context, gameId string, group *v1.GameMoveGroup, currentGroupNumber int64) error {
	// Load current history (or create empty)
	history, _ := storage.LoadFSArtifact[*v1.GameMoveHistory](s.storage, gameId, "history")
	if history == nil {
		history = &v1.GameMoveHistory{GameId: gameId}
	}

	// Append group to history
	history.Groups = append(history.Groups, group)

	// Save history
	return s.storage.SaveArtifact(gameId, "history", history)
}

// GetGameStateVersion implements GameStateUpdater interface
func (s *FSGamesService) GetGameStateVersion(ctx context.Context, id string) (int64, error) {
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, id, "state")
	if err != nil {
		return 0, err
	}
	return gameState.Version, nil
}

// UpdateGameStateScreenshotIndexInfo implements GameStateUpdater interface
// Note: This does NOT increment version - IndexInfo is internal bookkeeping
// that shouldn't invalidate user's optimistic lock
func (s *FSGamesService) UpdateGameStateScreenshotIndexInfo(ctx context.Context, id string, oldVersion int64, lastIndexedAt time.Time, needsIndexing bool) error {
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, id, "state")
	if err != nil {
		return err
	}

	// Check version matches - if not, content was updated and we'll re-index later
	if gameState.Version != oldVersion {
		return fmt.Errorf("version mismatch - content was updated, will re-index later")
	}

	// Update only IndexInfo fields, don't touch version
	if gameState.WorldData.ScreenshotIndexInfo == nil {
		gameState.WorldData.ScreenshotIndexInfo = &v1.IndexInfo{}
	}
	gameState.WorldData.ScreenshotIndexInfo.LastIndexedAt = tspb.New(lastIndexedAt)
	gameState.WorldData.ScreenshotIndexInfo.NeedsIndexing = needsIndexing
	// Note: NOT incrementing version - this is internal bookkeeping

	// Save updated game state
	err = s.storage.SaveArtifact(id, "state", gameState)
	if err != nil {
		return fmt.Errorf("failed to save game state: %w", err)
	}
	return nil
}

// ListGames returns all available games (metadata only for performance)
func (s *FSGamesService) ListGames(ctx context.Context, req *v1.ListGamesRequest) (resp *v1.ListGamesResponse, err error) {
	resp = &v1.ListGamesResponse{
		Items: []*v1.Game{},
		Pagination: &v1.PaginationResponse{
			HasMore:      false,
			TotalResults: 0,
		},
	}
	resp.Items, err = storage.ListFSEntities[*v1.Game](s.storage, nil)
	resp.Pagination.TotalResults = int32(len(resp.Items))

	// Populate screenshot URLs for all games
	for _, game := range resp.Items {
		if len(game.PreviewUrls) == 0 {
			game.PreviewUrls = []string{fmt.Sprintf("/screenshots/games/%s/default.png", game.Id)}
		}
	}

	return resp, nil
}

// CreateGame creates a new game
func (s *FSGamesService) CreateGame(ctx context.Context, req *v1.CreateGameRequest) (resp *v1.CreateGameResponse, err error) {
	// Load world data first so we can validate players have units/tiles
	worldsSvcClient := s.ClientMgr.GetWorldsSvcClient()
	world, err := worldsSvcClient.GetWorld(ctx, &v1.GetWorldRequest{Id: req.Game.WorldId})
	if err != nil {
		return nil, fmt.Errorf("Error loading world: %w", err)
	}

	// Validate the request (duplicate players, players with units/tiles, etc.)
	if err := s.ValidateCreateGameRequest(req.Game, world.WorldData); err != nil {
		return nil, err
	}

	// Create game entity directory
	customId := req.Game.Id
	req.Game.Id, err = s.storage.CreateEntity(req.Game.Id)
	if err != nil {
		// Check if this is an ID conflict (custom ID already exists)
		if customId != "" {
			// Generate a suggested ID by adding a random suffix
			suggestedId := customId + "-" + shortRandSuffix()
			resp = &v1.CreateGameResponse{
				FieldErrors: map[string]string{
					"id": suggestedId,
				},
			}
			return resp, nil
		}
		return nil, err
	}

	now := time.Now()
	req.Game.CreatedAt = tspb.New(now)
	req.Game.UpdatedAt = tspb.New(now)

	// Save a new empty game state and a new move list
	gs := &v1.GameState{
		GameId:        req.Game.Id,
		CurrentPlayer: 1, // Game starts with player 1
		TurnCounter:   1, // First turn starts at 1 for lazy top-up pattern
		WorldData:     world.WorldData,
	}

	// Auto-migrate WorldData from old list-based format to new map-based format
	lib.MigrateWorldData(gs.WorldData)

	// Generate shortcuts for tiles and units
	lib.EnsureShortcuts(gs.WorldData)

	// Initialize player runtime state with starting coins + base income
	s.InitializePlayerStates(gs, req.Game.Config)

	// Save game metadata (after adding base income to player coins)
	if err := s.storage.SaveArtifact(req.Game.Id, "metadata", req.Game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Units start with default zero values (current_turn=0, distance_left=0, available_health=0)
	// They will be lazily topped-up when accessed if unit.current_turn < game.turn_counter
	// This eliminates the need to initialize all units at game creation
	if err := s.storage.SaveArtifact(req.Game.Id, "state", gs); err != nil {
		log.Printf("Failed to create state for game %s: %v", req.Game.Id, err)
	}

	// Save a new empty game history and a new move list
	if err := s.storage.SaveArtifact(req.Game.Id, "history", &v1.GameMoveHistory{GameId: req.Game.Id}); err != nil {
		log.Printf("Failed to create state for game %s: %v", req.Game.Id, err)
	}

	resp = &v1.CreateGameResponse{
		Game:      req.Game,
		GameState: gs,
	}

	return resp, nil
}

// Helper functions for serialization

// serialize converts a protobuf message to bytes
func serialize(msg proto.Message) []byte {
	if msg == nil {
		return nil
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Failed to serialize: %v", err)
		return nil
	}
	return data
}

// computeHash generates a SHA256 hash of any protobuf message
func computeHash(msg proto.Message) string {
	if msg == nil {
		return ""
	}
	data := serialize(msg)
	if data == nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *FSGamesService) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (*v1.ProcessMovesResponse, error) {
	// If client didn't provide expected results, run ProcessMoves locally
	if req.ExpectedResponse == nil {
		return s.BaseGamesService.ProcessMoves(ctx, req)
	}

	// Client provided expected results - validate through coordinator

	// Get current game state
	gameresp, err := s.Self.GetGame(ctx, &v1.GetGameRequest{Id: req.GameId})
	if err != nil || gameresp.Game == nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	if gameresp.State == nil {
		return nil, fmt.Errorf("game state cannot be nil")
	}

	// Apply expected changes to current state to get new state
	// This is what validators will do independently
	/*
		newState := proto.Clone(gameresp.State).(*v1.GameState)
		// TODO: Apply req.ExpectedChanges to newState

		// Compute hashes (validators will compute same hashes)
			fromStateHash := computeHash(gameresp.State)
			toStateHash := computeHash(newState)

			// Serialize for coordinator (game-agnostic blobs)
			movesBlob := serialize(&v1.ProcessMovesRequest{
				GameId: req.GameId,
				Moves:  req.Moves,
			})
			changesBlob := serialize(req.ExpectedResponse)
			newStateBlob := serialize(newState)

			// Create proposal for coordinator
				proposal := &turnengine.SubmitProposalRequest{
					SessionId:     req.GameId,
					ProposerId:    "player1", // TODO: Get from context/session
					FromStateHash: fromStateHash,
					ToStateHash:   toStateHash,
					MovesBlob:     movesBlob,
					ChangesBlob:   changesBlob,
					NewStateBlob:  newStateBlob,
				}

				// TODO: Submit to coordinator when wired up
				_ = proposal
	*/

	// For now, just apply the changes directly (simulating accepted proposal)
	return req.ExpectedResponse, nil
}

// ement coordination.Callbacks interface

// OnProposalStarted is called when a proposal is accepted for validation
/*
func (s *FSGamesService) OnProposalStarted(gameID string, proposal *turnengine.ProposalInfo) error {
	// Load the game state
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	// Set the proposal tracking info
		gameState.ProposalInfo = &turnengine.ProposalTrackingInfo{
			ProposalId:     proposal.ProposalId,
			ProposerId:     proposal.ProposerId,
			Phase:          turnengine.ProposalPhase_PROPOSAL_PHASE_COLLECTING,
			CreatedAt:      proposal.CreatedAt,
			ValidatorCount: int32(len(proposal.AssignedValidators)),
			VotesReceived:  0,
		}

	// Save the updated state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}
*/

// OnProposalAccepted is called when consensus approves the proposal
/*
func (s *FSGamesService) OnProposalAccepted(gameID string, proposal *turnengine.ProposalInfo) error {
	// The new state is in the proposal's new_state_blob
	// We need to save it as the new game state

	// Note: In a real implementation, we'd unmarshal proposal.NewStateBlob
	// For now, just clear the proposal info

	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	// Clear proposal info and update state hash
	// gameState.ProposalInfo = nil
	gameState.StateHash = proposal.ToStateHash

	// Save the state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}
*/

// OnProposalFailed is called when proposal is rejected or times out
/*
func (s *FSGamesService) OnProposalFailed(gameID string, proposal *turnengine.ProposalInfo, reason string) error {
	// Clear the proposal info from game state
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	// Clear proposal info
	// gameState.ProposalInfo = nil

	// Save the state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}
*/

// ListMoves returns moves from game history, optionally filtered by group range
func (s *FSGamesService) ListMoves(ctx context.Context, req *v1.ListMovesRequest) (*v1.ListMovesResponse, error) {
	if req.GameId == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	// Load history
	history, err := storage.LoadFSArtifact[*v1.GameMoveHistory](s.storage, req.GameId, "history")
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	var groups []*v1.GameMoveGroup
	for _, group := range history.Groups {
		// Filter by group range
		if req.FromGroup > 0 && group.GroupNumber < req.FromGroup {
			continue
		}
		if req.ToGroup > 0 && group.GroupNumber > req.ToGroup {
			break
		}
		groups = append(groups, group)
	}

	return &v1.ListMovesResponse{
		MoveGroups: groups,
		HasMore:    req.FromGroup > 0 && len(history.Groups) > 0 && history.Groups[0].GroupNumber < req.FromGroup,
	}, nil
}
