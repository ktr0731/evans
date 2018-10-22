package grpc

import "github.com/ktr0731/evans/entity"

var _ entity.GRPCReflectionClient = (*reflectionClient)(nil)
