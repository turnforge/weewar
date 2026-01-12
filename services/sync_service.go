package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/panyam/gocurrent"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	v1s "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/services"
	"google.golang.org/grpc"
)

// GameSyncService handles real-time synchronization of game state across
// multiple connected clients for multiplayer gameplay.
//
// This is a pure pub/sub service - it does NOT handle:
// - Move validation (that's GamesService)
// - RNG/seed management (that's lib/)
//
// Architecture:
// - Uses gocurrent.FanOut for efficient per-game message broadcasting
// - GamesService calls Broadcast RPC after ProcessMoves succeeds
// - Subscribers receive GameUpdates via streaming RPC
type GameSyncService struct {
	v1s.UnimplementedGameSyncServiceServer

	// Per-game FanOut instances for broadcasting updates
	// gameId -> FanOut
	fanOuts map[string]*gocurrent.FanOut[*v1.GameUpdate]

	// Per-game sequence numbers for ordering
	sequences map[string]int64

	mu sync.RWMutex
}

// NewGameSyncService creates a new sync service
func NewGameSyncService() *GameSyncService {
	return &GameSyncService{
		fanOuts:   make(map[string]*gocurrent.FanOut[*v1.GameUpdate]),
		sequences: make(map[string]int64),
	}
}

// getFanOut returns (or creates) the FanOut for a game
func (s *GameSyncService) getFanOut(gameId string) *gocurrent.FanOut[*v1.GameUpdate] {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fo, exists := s.fanOuts[gameId]; exists {
		return fo
	}

	// Create new FanOut with buffered input to prevent blocking
	fo := gocurrent.NewFanOut[*v1.GameUpdate](
		gocurrent.WithFanOutInputBuffer[*v1.GameUpdate](100),
	)
	s.fanOuts[gameId] = fo
	return fo
}

// Subscribe streams game updates to a client.
// Supports reconnection via from_sequence.
func (s *GameSyncService) Subscribe(req *v1.SubscribeRequest, stream grpc.ServerStreamingServer[v1.GameUpdate]) error {
	gameId := req.GameId
	playerId := req.PlayerId

	// Get current sequence
	s.mu.RLock()
	currentSeq := s.sequences[gameId]
	s.mu.RUnlock()

	// Send initial state (game state should be loaded separately by client via GetGame)
	initialState := &v1.SubscribeResponse{
		CurrentSequence: currentSeq,
	}

	err := stream.Send(&v1.GameUpdate{
		Sequence: currentSeq,
		UpdateType: &v1.GameUpdate_InitialState{
			InitialState: initialState,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send initial state: %w", err)
	}

	// TODO: If req.FromSequence > 0, send missed updates from history
	// For now, we only support fresh subscriptions

	// Get or create FanOut for this game
	fanOut := s.getFanOut(gameId)

	// Create output channel for this subscriber
	outputChan := fanOut.New(nil)
	defer func() {
		<-fanOut.Remove(outputChan, true)
	}()

	// Broadcast player joined
	s.broadcastInternal(gameId, &v1.GameUpdate{
		Sequence: s.nextSequence(gameId),
		UpdateType: &v1.GameUpdate_PlayerJoined{
			PlayerJoined: &v1.PlayerJoined{
				PlayerId: playerId,
			},
		},
	})

	// Stream updates to client until disconnect
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			// Client disconnected - broadcast player left
			s.broadcastInternal(gameId, &v1.GameUpdate{
				Sequence: s.nextSequence(gameId),
				UpdateType: &v1.GameUpdate_PlayerLeft{
					PlayerLeft: &v1.PlayerLeft{
						PlayerId: playerId,
					},
				},
			})
			return nil

		case update, ok := <-outputChan:
			if !ok {
				// Channel closed (FanOut stopped)
				return nil
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// Broadcast sends a GameUpdate to all subscribers of a game.
// Called by GamesService (via gRPC client) after ProcessMoves succeeds.
func (s *GameSyncService) Broadcast(ctx context.Context, req *v1.BroadcastRequest) (*v1.BroadcastResponse, error) {
	gameId := req.GameId
	update := req.Update

	// Ensure sequence is set
	if update.Sequence == 0 {
		update.Sequence = s.nextSequence(gameId)
	}

	count := s.broadcastInternal(gameId, update)

	return &v1.BroadcastResponse{
		SubscriberCount: int32(count),
		Sequence:        update.Sequence,
	}, nil
}

// broadcastInternal sends a GameUpdate to all subscribers (internal use)
func (s *GameSyncService) broadcastInternal(gameId string, update *v1.GameUpdate) int {
	s.mu.RLock()
	fo, exists := s.fanOuts[gameId]
	s.mu.RUnlock()

	if !exists || fo == nil {
		return 0
	}

	count := fo.Count()
	log.Println("Broadcasting to: ", count, update)
	if count > 0 {
		fo.Send(update)
	}
	return count
}

// nextSequence atomically increments and returns the next sequence number for a game
func (s *GameSyncService) nextSequence(gameId string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sequences[gameId]++
	return s.sequences[gameId]
}

// SubscriberCount returns the number of subscribers for a game
func (s *GameSyncService) SubscriberCount(gameId string) int {
	s.mu.RLock()
	fo, exists := s.fanOuts[gameId]
	s.mu.RUnlock()

	if !exists || fo == nil {
		return 0
	}
	return fo.Count()
}
