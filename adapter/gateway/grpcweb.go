package gateway

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/k0kubun/pp"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/pkg/errors"
)

type GRPCWebClient struct {
	config *config.Config
	conn   *grpcweb.Client

	*gRPCReflectoinClient
}

func NewGRPCWebClient(config *config.Config) *GRPCWebClient {
	pp.Println("WEB!!!")
	conn := grpcweb.NewClient(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	return &GRPCWebClient{
		config: config,
		conn:   conn,
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

func (c *GRPCWebClient) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
	panic("not implemented yet")
	return nil, nil
}

func (c *GRPCWebClient) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
	panic("not implemented yet")
	return nil, nil
}

func (c *GRPCWebClient) Close(ctx context.Context) error {
	// TODO
	return nil
}
