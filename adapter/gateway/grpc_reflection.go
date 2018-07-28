package gateway

import (
	"context"

	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/ktr0731/evans/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

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

func (c *gRPCReflectoinClient) ListServices() []entity.Service {
	return nil
}
