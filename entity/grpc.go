package entity

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type GRPCClient interface {
	Invoke(ctx context.Context, fqrn string, req, res interface{}) error
	NewClientStream(ctx context.Context, rpc RPC) (ClientStream, error)
}

type ClientStream interface {
	Send(req proto.Message) error
	CloseAndReceive(res proto.Message) error
}
