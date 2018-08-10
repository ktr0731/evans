package gateway

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/pkg/errors"
)

type GRPCWebClient struct {
	config *config.Config
	conn   *grpcweb.Client

	builder port.DynamicBuilder

	*gRPCReflectoinClient
}

func NewGRPCWebClient(config *config.Config, builder port.DynamicBuilder) *GRPCWebClient {
	conn := grpcweb.NewClient(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	client := &GRPCWebClient{
		config:  config,
		conn:    conn,
		builder: builder,
	}

	if config.Server.Reflection {
		// TODO: gRPC Web + gRPC reflection
	}

	return client
}

func (c *GRPCWebClient) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return errors.Wrap(err, "failed to convert FQRN to endpoint")
	}
	request := grpcweb.NewRequest(endpoint, req.(proto.Message), res.(proto.Message))
	res, err = c.conn.Unary(ctx, request)
	return err
}

type webClientStream struct {
	conn grpcweb.ClientStreamClient

	newRequest func(req proto.Message) *grpcweb.Request
}

func (s *webClientStream) Send(req proto.Message) error {
	request := s.newRequest(req)
	return s.conn.Send(request)
}

func (s *webClientStream) CloseAndReceive(res *proto.Message) error {
	response, err := s.conn.CloseAndReceive()
	if err != nil {
		return err
	}
	*res = response.Content.(proto.Message)
	return nil
}

func (c *GRPCWebClient) NewClientStream(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
	endpoint, err := fqrnToEndpoint(rpc.FQRN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}

	if err != nil {
		return nil, err
	}
	cc, err := c.conn.ClientStreaming(ctx)
	if err != nil {
		return nil, err
	}
	return &webClientStream{
		conn: cc,
		newRequest: func(req proto.Message) *grpcweb.Request {
			return grpcweb.NewRequest(
				endpoint,
				req,
				c.builder.NewMessage(rpc.ResponseMessage()),
			)
		},
	}, nil
}

type webServerStream struct {
	newClient func(req proto.Message) (grpcweb.ServerStreamClient, error)
	conn      grpcweb.ServerStreamClient
}

func (s *webServerStream) Send(req proto.Message) (err error) {
	s.conn, err = s.newClient(req)
	return
}

func (s *webServerStream) Receive(res *proto.Message) error {
	if s.conn == nil {
		return errors.New("Receive must be call after Send method")
	}
	resp, err := s.conn.Receive()
	if err != nil {
		return err
	}
	*res = resp.Content.(proto.Message)
	return nil
}

func (c *GRPCWebClient) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
	endpoint, err := fqrnToEndpoint(rpc.FQRN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}

	newClient := func(req proto.Message) (grpcweb.ServerStreamClient, error) {
		request := grpcweb.NewRequest(
			endpoint,
			req,
			c.builder.NewMessage(rpc.ResponseMessage()),
		)

		sc, err := c.conn.ServerStreaming(ctx, request)
		if err != nil {
			return nil, err
		}

		return sc, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to make new server streaming gRPC Web request")
	}
	return &webServerStream{newClient: newClient}, nil
}

type webBidiStream struct {
	conn grpcweb.BidiStreamClient

	newRequest func(req proto.Message) *grpcweb.Request
}

func (s *webBidiStream) Send(req proto.Message) error {
	request := s.newRequest(req)
	return s.conn.Send(request)
}

func (s *webBidiStream) Receive(res *proto.Message) error {
	response, err := s.conn.Receive()
	if err != nil {
		return err
	}
	*res = response.Content.(proto.Message)
	return nil
}

func (s *webBidiStream) Close() error {
	return s.conn.Close()
}

func (c *GRPCWebClient) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
	endpoint, err := fqrnToEndpoint(rpc.FQRN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}

	newRequest := func(req proto.Message) *grpcweb.Request {
		return grpcweb.NewRequest(
			endpoint,
			req,
			c.builder.NewMessage(rpc.ResponseMessage()),
		)
	}

	req := newRequest(c.builder.NewMessage(rpc.RequestMessage()))
	sc := c.conn.BidiStreaming(ctx, req)

	return &webBidiStream{
		conn:       sc,
		newRequest: newRequest,
	}, nil
}

func (c *GRPCWebClient) Close(ctx context.Context) error {
	c.gRPCReflectoinClient.Close()
	return nil
}
