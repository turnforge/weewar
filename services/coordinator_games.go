package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/panyam/turnengine/engine/coordination"
	"github.com/panyam/turnengine/engine/storage"
	turnengine "github.com/panyam/turnengine/games/weewar/gen/go/turnengine/v1"
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	weewar "github.com/panyam/turnengine/games/weewar/lib"
	"google.golang.org/protobuf/proto"
)

// CoordinatorGamesService wraps FSGamesService with coordination support
type CoordinatorGamesService struct {
	FSGamesServiceImpl
	coordinator *coordination.Service
}

// NewCoordinatorGamesService creates a new games service with coordination
func NewCoordinatorGamesService() (*CoordinatorGamesService, error) {
	// Create base file storage for games
	if GAMES_STORAGE_DIR == "" {
		GAMES_STORAGE_DIR = weewar.DevDataPath("storage/games")
	}

	// Create coordination storage (could be same dir or separate)
	coordStorage, err := coordination.NewFileCoordinationStorage(GAMES_STORAGE_DIR)
	if err != nil {
		return nil, err
	}

	// Create the service
	service := &CoordinatorGamesService{
		FSGamesServiceImpl: *NewFSGamesService(),
	}

	// Create coordinator with callbacks
	config := coordination.Config{
		RequiredValidators: 1, // Start with 1 for testing
		ValidationTimeout:  5 * time.Minute,
	}

	service.coordinator = coordination.NewService(coordStorage, config, service)
	service.Self = service // Update Self pointer

	return service, nil
}

// Implement coordination.Callbacks interface

// OnProposalStarted is called when a proposal is accepted for validation
func (s *CoordinatorGamesService) OnProposalStarted(gameID string, proposal *turnengine.ProposalInfo) error {
	// Update the game state with proposal info
	game, err := storage.LoadFSArtifact[*v1.Game](s.storage, gameID, "metadata")
	if err != nil {
		return err
	}

	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return err
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

// OnProposalAccepted is called when consensus approves the proposal
func (s *CoordinatorGamesService) OnProposalAccepted(gameID string, proposal *turnengine.ProposalInfo) error {
	// The new state is in the proposal's NewStateBlob
	var newState v1.GameState
	if err := proto.Unmarshal(proposal.NewStateBlob, &newState); err != nil {
		return fmt.Errorf("failed to unmarshal new state: %w", err)
	}

	// Clear the proposal info since it's now committed
	newState.ProposalInfo = nil

	// Update the state hash
	newState.StateHash = proposal.ToStateHash

	// Save the new committed state
	if err := s.storage.AtomicSaveArtifact(gameID, "state", &newState); err != nil {
		return fmt.Errorf("failed to save new state: %w", err)
	}

	// Also update the game metadata (e.g., last updated time)
	game, _ := storage.LoadFSArtifact[*v1.Game](s.storage, gameID, "metadata")
	if game != nil {
		s.storage.SaveArtifact(gameID, "metadata", game)
	}

	return nil
}

// OnProposalFailed is called when proposal is rejected or times out
func (s *CoordinatorGamesService) OnProposalFailed(gameID string, proposal *turnengine.ProposalInfo, reason string) error {
	// Clear the proposal from game state
	gameState, err := storage.LoadFSArtifact[*v1.GameState](s.storage, gameID, "state")
	if err != nil {
		return err
	}

	// Clear proposal info
	gameState.ProposalInfo = nil

	// Save the state
	return s.storage.SaveArtifact(gameID, "state", gameState)
}

// Override ProcessMoves to use coordinator

// ProcessMoves now just forwards to coordinator - server doesn't run game logic
func (s *CoordinatorGamesService) ProcessMoves(ctx context.Context, req *v1.ProcessMovesRequest) (*v1.ProcessMovesResponse, error) {
	// This is called by the WASM after it has already validated moves locally
	// The request should contain the computed results

	// We expect the WASM to send us:
	// - The moves it wants to make
	// - The changes computed by ProcessMoves
	// - The new state after applying changes
	// - Hash of current and new state

	// For now, return an error indicating this needs to be implemented differently
	return nil, fmt.Errorf("ProcessMoves should be called with pre-computed results from WASM")
}

// SubmitProposal is the new endpoint for WASM to submit validated moves
func (s *CoordinatorGamesService) SubmitProposal(ctx context.Context, req *v1.SubmitProposalRequest) (*v1.SubmitProposalResponse, error) {
	// Convert to coordinator request
	coordReq := &turnengine.SubmitProposalRequest{
		SessionId:     req.GameId,
		ProposerId:    req.PlayerId,
		FromStateHash: req.FromStateHash,
		ToStateHash:   req.ToStateHash,
		MovesBlob:     req.MovesBlob,
		ChangesBlob:   req.ChangesBlob,
		NewStateBlob:  req.NewStateBlob,
		Nonce:         req.Nonce,
	}

	// Submit to coordinator
	coordResp, err := s.coordinator.SubmitProposal(coordReq)
	if err != nil {
		return nil, err
	}

	// Convert response
	return &v1.SubmitProposalResponse{
		Status:             coordResp.Status == turnengine.SubmitProposalResponse_STATUS_ACCEPTED,
		ProposalId:         coordResp.ProposalId,
		Reason:             coordResp.Reason,
		AssignedValidators: coordResp.AssignedValidators,
	}, nil
}

// ValidateProposal is called by validators to submit their validation
func (s *CoordinatorGamesService) ValidateProposal(ctx context.Context, req *v1.ValidateProposalRequest) (*v1.ValidateProposalResponse, error) {
	// Convert to coordinator request
	coordReq := &turnengine.SubmitValidationRequest{
		SessionId:    req.GameId,
		ProposalId:   req.ProposalId,
		ValidatorId:  req.ValidatorId,
		Approved:     req.Approved,
		ComputedHash: req.ComputedHash,
		ErrorReason:  req.ErrorReason,
		Signature:    req.Signature,
	}

	// Submit validation
	coordResp, err := s.coordinator.SubmitValidation(coordReq)
	if err != nil {
		return nil, err
	}

	// Convert response
	return &v1.ValidateProposalResponse{
		Recorded:          coordResp.Recorded,
		ConsensusReached:  coordResp.ConsensusReached,
		ConsensusApproved: coordResp.ConsensusApproved,
	}, nil
}

// GetPendingValidation checks if a game needs validation from a specific validator
func (s *CoordinatorGamesService) GetPendingValidation(ctx context.Context, req *v1.GetPendingValidationRequest) (*v1.GetPendingValidationResponse, error) {
	validation, err := s.coordinator.GetPendingValidationForGame(req.GameId, req.ValidatorId)
	if err != nil {
		return nil, err
	}

	if validation == nil {
		return &v1.GetPendingValidationResponse{}, nil
	}

	// Convert to WeeWar format
	return &v1.GetPendingValidationResponse{
		ProposalId:    validation.ProposalId,
		ProposerId:    validation.ProposerId,
		FromStateHash: validation.FromStateHash,
		MovesBlob:     validation.MovesBlob,
		ChangesBlob:   validation.ChangesBlob,
		Deadline:      validation.Deadline,
		Nonce:         validation.Nonce,
	}, nil
}

// Helper to compute state hash
func computeStateHash(state *v1.GameState) string {
	data, _ := proto.Marshal(state)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
