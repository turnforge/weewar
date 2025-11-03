//go:build !wasm
// +build !wasm

package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"
)

// Bridge implementation that converts Connect stream to gRPC stream interface
type ConnectStreamBridge[T any] struct {
	connectStream *connect.ServerStream[T]
	ctx           context.Context
}

// Implement the gRPC stream interface
func (b *ConnectStreamBridge[T]) Send(msg *T) error {
	return b.connectStream.Send(msg)
}

func (b *ConnectStreamBridge[T]) Context() context.Context {
	return b.ctx
}

// Implement other required gRPC stream methods
func (b *ConnectStreamBridge[T]) SendMsg(m interface{}) error {
	if msg, ok := m.(*T); ok {
		return b.connectStream.Send(msg)
	}
	return fmt.Errorf("invalid message type")
}

func (b *ConnectStreamBridge[T]) RecvMsg(m interface{}) error {
	// Not used for server streaming
	return fmt.Errorf("RecvMsg not supported for server streaming")
}

func (b *ConnectStreamBridge[T]) SetHeader(metadata.MD) error {
	// Handle metadata if needed
	return nil
}

func (b *ConnectStreamBridge[T]) SendHeader(metadata.MD) error {
	// Handle metadata if needed
	return nil
}

func (b *ConnectStreamBridge[T]) SetTrailer(metadata.MD) {
	// Handle metadata if needed
}
