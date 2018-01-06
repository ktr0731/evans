package gateway

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
)

type GRPCClient struct {
	config *config.Config
	env    entity.Environment
	conn   *grpc.ClientConn
}

func NewGRPCClient(config *config.Config, env entity.Environment) (*GRPCClient, error) {
	// TODO: secure option
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		config: config,
		env:    env,
		conn:   conn,
	}, nil
}

func (c *GRPCClient) Invoke(ctx context.Context, req, res interface{}) error {
	return grpc.Invoke(ctx, c.endpoint(), req, res, c.conn)
}

func (c *GRPCClient) endpoint() string {
	return fmt.Sprintf("/%s.%s/%s", e.state.currentPackage, e.state.currentService, rpcName)
}
