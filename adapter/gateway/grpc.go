package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"github.com/ktr0731/evans/config"
)

type GRPCClient struct {
	config *config.Config
	conn   *grpc.ClientConn
}

func NewGRPCClient(config *config.Config) (*GRPCClient, error) {
	// TODO: secure option
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port), grpc.WithInsecure())
	if err != nil {
		return nil, err
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

// fqrnToEndpoint converts FullQualifiedRPCName to endpoint
//
// e.g.
//	pkg_name.svc_name.rpc_name -> /pkg_name.svc_name/rpc_name
func (c *GRPCClient) fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	if len(sp) != 3 {
		return "", errors.New("invalid FQRN format")
	}
	return fmt.Sprintf("/%s.%s/%s", sp[0], sp[1], sp[2]), nil
}
