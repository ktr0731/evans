package grpc

import (
	"context"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/ktr0731/grpc-web-go-client/grpcweb/grpcweb_reflection_v1alpha"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const reflectionServiceName = "grpc.reflection.v1alpha.ServerReflection"

type reflectionClient struct {
	client *grpcreflect.Client
}

func newReflectionClient(conn *grpc.ClientConn) *reflectionClient {
	return &reflectionClient{
		client: grpcreflect.NewClient(context.Background(), grpc_reflection_v1alpha.NewServerReflectionClient(conn)),
	}
}

func newWebReflectionClient(conn *grpcweb.Client) *reflectionClient {
	return &reflectionClient{
		client: grpcreflect.NewClient(context.Background(), grpcweb_reflection_v1alpha.NewServerReflectionClient(conn)),
	}
}

func (c *reflectionClient) ReflectionEnabled() bool {
	return c != nil
}

func (c *reflectionClient) ListPackages() ([]*entity.Package, error) {
	ssvcs, err := c.client.ListServices()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list services from reflecton enabled gRPC server")
	}

	var fds []*desc.FileDescriptor
	for _, s := range ssvcs {
		if s == reflectionServiceName {
			continue
		}
		svc, err := c.client.ResolveService(s)
		if err != nil {
			return nil, err
		}
		fds = append(fds, svc.GetFile())
	}

	return protobuf.ToEntitiesFrom(fds)
}

func (c *reflectionClient) Close() {
	if c == nil {
		return
	}
	c.client.Reset()
}
