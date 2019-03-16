package entity

import (
	"context"
	"errors"

	"github.com/golang/protobuf/proto"
)

var ErrMutualAuthParamsAreNotEnough = errors.New("cert and certkey are required to authenticate mutually")

type GRPCClient interface {
	Invoke(ctx context.Context, fqrn string, req, res interface{}) error
	NewClientStream(ctx context.Context, rpc RPC) (ClientStream, error)
	NewServerStream(ctx context.Context, rpc RPC) (ServerStream, error)
	NewBidiStream(ctx context.Context, rpc RPC) (BidiStream, error)
	Close(ctx context.Context) error

	GRPCReflectionClient
}

type ClientStream interface {
	Send(req proto.Message) error
	CloseAndReceive(res *proto.Message) error
}

type ServerStream interface {
	Send(req proto.Message) error
	Receive(res *proto.Message) error
}

type BidiStream interface {
	Send(req proto.Message) error
	Receive(res *proto.Message) error
	CloseSend() error
}

type GRPCReflectionClient interface {
	ReflectionEnabled() bool
	ListPackages() ([]*Package, error)
}
