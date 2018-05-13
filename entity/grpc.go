package entity

import (
	"context"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type GRPCClient interface {
	Invoke(ctx context.Context, fqrn string, req, res interface{}) error
	NewClientStream(ctx context.Context, rpc RPC) ClientStream
}

type ClientStream interface {
	Send(proto.Message) error
	CloseAndReceive() (proto.Message, error)
}

type clientStream struct {
	cs grpc.ClientStream
}

func newClientStream(cs grpc.ClientStream) *clientStream {
	return &clientStream{cs}
}

func (s *clientStream) Send(m proto.Message) error {
	return s.cs.SendMsg(m)
}

func (s *clientStream) CloseAndReceive(res proto.Message) error {
	err := s.cs.RecvMsg(res)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to close and receive response")
	}
	return nil
}
