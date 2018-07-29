package gateway

import (
	"context"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const reflectionServiceName = "grpc.reflection.v1alpha.ServerReflection"

type gRPCReflectoinClient struct {
	client *grpcreflect.Client
}

func newGRPCReflectionClient(conn *grpc.ClientConn) *gRPCReflectoinClient {
	return &gRPCReflectoinClient{
		client: grpcreflect.NewClient(context.Background(), grpc_reflection_v1alpha.NewServerReflectionClient(conn)),
	}
}

func (c *gRPCReflectoinClient) ReflectionEnabled() bool {
	return c != nil
}

func (c *gRPCReflectoinClient) ListServices() ([]entity.Service, error) {
	ssvcs, err := c.client.ListServices()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list services from reflecton enabled gRPC server")
	}

	svcs := make([]*desc.ServiceDescriptor, 0, len(ssvcs)-1)
	for _, s := range ssvcs {
		if s == reflectionServiceName {
			continue
		}
		svc, err := c.client.ResolveService(s)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, svc)
	}

	return protobuf.ToEntitiesFromServiceDescriptors(svcs), nil
}

func (c *gRPCReflectoinClient) Close() {
	if c == nil {
		return
	}
	c.client.Reset()
}
