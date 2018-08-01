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
	return &GRPCWebClient{
		config:  config,
		conn:    conn,
		builder: builder,
	}
}

func (c *GRPCWebClient) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	endpoint, err := fqrnToEndpoint(fqrn)
	if err != nil {
		return errors.Wrap(err, "failed to convert FQRN to endpoint")
	}
	request, err := grpcweb.NewRequest(endpoint, req.(proto.Message), res.(proto.Message))
	if err != nil {
		return errors.Wrap(err, "failed to make new gRPC Web request")
	}
	return c.conn.Unary(ctx, request)
}

func (c *GRPCWebClient) NewClientStream(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
	panic("not implemented yet")
	return nil, nil
}

type webServerStream struct {
	req  *grpcweb.Request
	conn *grpcweb.ServerStreamClient
}

func (s *webServerStream) Send(_ proto.Message) error {
	// do nothing.
	// gRPC Web client sends a request at ServerStream method
	return nil
}

func (s *webServerStream) Receive(res *proto.Message) error {
	resp, err := s.conn.Recv()
	if err != nil {
		return err
	}
	*res = resp
	return nil
}

func (c *GRPCWebClient) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
	endpoint, err := fqrnToEndpoint(rpc.FQRN())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert FQRN to endpoint")
	}
	request, err := grpcweb.NewRequest(
		endpoint,
		c.builder.NewMessage(rpc.RequestMessage()),
		c.builder.NewMessage(rpc.ResponseMessage()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make new server streaming gRPC Web request")
	}
	sc, _ := c.conn.ServerStreaming(ctx, request)
	return &webServerStream{conn: sc}, nil
}

func (c *GRPCWebClient) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
	panic("not implemented yet")
	return nil, nil
}

func (c *GRPCWebClient) Close(ctx context.Context) error {
	// TODO
	return nil
}
