package grpc

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type client struct {
	conn *grpc.ClientConn

	*reflectionClient
}

// NewClient creates a new gRPC client.
// It dials to the server specified by addr.
// addr format is same as the first argument of grpc.Dial.
// If useReflection is true, the gRPC client enables gRPC reflection.
func NewClient(addr string, useReflection bool) (entity.GRPCClient, error) {
	// TODO: secure option
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial to gRPC server")
	}
	switch s := conn.GetState(); s {
	case connectivity.TransientFailure:
		return nil, errors.Errorf("connection transient failure, is the gRPC server running?: %s", s)
	case connectivity.Shutdown:
		return nil, errors.Errorf("the gRPC server was closed: %s", s)
	}

	client := &client{
		conn: conn,
	}

	if useReflection {
		client.reflectionClient = newReflectionClient(conn)
	}

	return client, nil
}

func (c *client) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return err
	}
	return grpc.Invoke(ctx, endpoint, req, res, c.conn)
}

func (c *client) Close(ctx context.Context) error {
	doneCh := make(chan error)
	go func() {
		var result error
		c.reflectionClient.Close()
		if err := c.conn.Close(); err != nil {
			result = multierror.Append(result, errors.Wrap(err, "failed to close gRPC client"))
		}
		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-doneCh:
		return err
	}
}

type clientStream struct {
	cs grpc.ClientStream
}

func (s *clientStream) Send(m proto.Message) error {
	return s.cs.SendMsg(m)
}

func (s *clientStream) CloseAndReceive(res *proto.Message) error {
	if err := s.cs.CloseSend(); err != nil {
		return errors.Wrap(err, "failed to close client stream")
	}

	err := s.cs.RecvMsg(*res)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to close and receive response")
	}
	return nil
}

func (c *client) NewClientStream(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
	endpoint, err := fqrnToEndpoint(rpc.FQRN())
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

func (s *serverStream) Receive(res *proto.Message) error {
	return s.cs.RecvMsg(*res)
}

func (c *client) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
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

func (s *bidiStream) Receive(res *proto.Message) error {
	return s.s.cs.RecvMsg(*res)
}

func (s *bidiStream) Close() error {
	return s.s.cs.CloseSend()
}

func (c *client) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
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
func fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	if len(sp) < 3 {
		return "", errors.New("invalid FQRN format")
	}

	return fmt.Sprintf("/%s/%s", strings.Join(sp[:len(sp)-1], "."), sp[len(sp)-1]), nil
}
