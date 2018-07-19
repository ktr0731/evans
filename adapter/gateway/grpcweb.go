package gateway

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/pkg/errors"
)

type GRPCWebClient struct {
	config *config.Config
	conn   *grpcweb.Client

	entity.GRPCClient
}

func NewGRPCWebClient(config *config.Config) *GRPCWebClient {
	conn := grpcweb.NewClient(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	return &GRPCWebClient{
		config: config,
		conn:   conn,
	}
}

func (c *GRPCWebClient) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	request, err := grpcweb.NewRequest(fqrn, req.(proto.Message), res.(proto.Message))
	if err != nil {
		return errors.Wrap(err, "failed to make new gRPC Web request")
	}
	return c.conn.Unary(ctx, request)
}
