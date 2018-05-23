package gateway

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
)

type GRPCClient struct {
	config *config.Config
	conn   *grpc.ClientConn
}

func NewGRPCClient(config *config.Config) (*GRPCClient, error) {
	// TODO: secure option
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial to gRPC server")
	}
	switch s := conn.GetState(); s {
	case connectivity.TransientFailure:
		return nil, errors.Errorf("connection transient failure, is the gRPC server running?: %s", s)
	case connectivity.Shutdown:
		return nil, errors.Errorf("the gRPC server was closed: %s", s)
	}
	return &GRPCClient{
		config: config,
		conn:   conn,
	}, nil
}

func (c *GRPCClient) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	endpoint, err := c.fqrnToEndpoint(fqrn)
	if err != nil {
		return err
	}
	return grpc.Invoke(ctx, endpoint, req, res, c.conn)
}

type clientStream struct {
	cs grpc.ClientStream
}

func (s *clientStream) Send(m proto.Message) error {
	return s.cs.SendMsg(m)
}

func (s *clientStream) CloseAndReceive(res proto.Message) error {
	if err := s.cs.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close client stream")
	}

	err := s.cs.RecvMsg(res)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to close and receive response")
	}
	return nil
}

func (c *GRPCClient) NewClientStream(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
	endpoint, err := c.fqrnToEndpoint(rpc.FQRN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert fqrn to endpoint")
	}
	cs, err := c.conn.NewStream(ctx, rpc.StreamDesc(), endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to instantiate gRPC stream")
	}
	return &clientStream{cs}, nil
}

type serverStream struct {
	*clientStream
}

func (s *serverStream) Receive(res proto.Message) error {
	return s.cs.RecvMsg(res)
}

func (c *GRPCClient) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
	s, err := c.NewClientStream(ctx, rpc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server stream")
	}
	return &serverStream{s.(*clientStream)}, nil
}

type bidiStream struct {
	s *serverStream
}

func (s *bidiStream) Send(res proto.Message) error {
	return s.s.cs.SendMsg(res)
}

func (s *bidiStream) Receive(res proto.Message) error {
	return s.s.cs.RecvMsg(res)
}

func (s *bidiStream) Close() error {
	return s.s.cs.CloseSend()
}

func (c *GRPCClient) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
	s, err := c.NewServerStream(ctx, rpc)
	if err != nil {
		return nil, err
	}
	return &bidiStream{s.(*serverStream)}, nil
}

// fqrnToEndpoint converts FullQualifiedRPCName to endpoint
//
// e.g.
//	pkg_name.svc_name.rpc_name -> /pkg_name.svc_name/rpc_name
func (c *GRPCClient) fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	if len(sp) < 3 {
		return "", errors.New("invalid FQRN format")
	}

	return fmt.Sprintf("/%s/%s", strings.Join(sp[:len(sp)-1], "."), sp[len(sp)-1]), nil
}
