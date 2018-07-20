package entity

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type GRPCClient interface {
	Invoke(ctx context.Context, fqrn string, req, res interface{}) error
	NewClientStream(ctx context.Context, rpc RPC) (ClientStream, error)
	NewServerStream(ctx context.Context, rpc RPC) (ServerStream, error)
	NewBidiStream(ctx context.Context, rpc RPC) (BidiStream, error)
}

type ClientStream interface {
	Send(req proto.Message) error
	CloseAndReceive(res proto.Message) error
}

type ServerStream interface {
	Send(req proto.Message) error
	Receive(res proto.Message) error
}

type BidiStream interface {
	Send(req proto.Message) error
	Receive(res proto.Message) error
	Close() error
}

type GRPCReflectionClient interface {
	Enabled() bool
	ListServices() []Service
}
