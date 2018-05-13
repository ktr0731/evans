package entity

import (
	"context"

	"google.golang.org/grpc"
)

type GRPCClient interface {
	Invoke(ctx context.Context, fqrn string, req, res interface{}) error
	NewClientStream(ctx context.Context, rpc RPC) grpc.ClientStream
}

type clientStream struct {
	grpc.ClientStream
}

func newClientStream(cs grpc.ClientStream) *clientStream {
	return &clientStream{cs}
}
