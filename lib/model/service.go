package model

import "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

type RPC struct {
	Name string
}

type Service struct {
	Name string
	RPCs []RPC
}

func NewService(service *descriptor.ServiceDescriptorProto) *Service {
	rpcs := make([]RPC, len(service.GetMethod()))
	for i, rpc := range service.GetMethod() {
		rpcs[i] = RPC{
			Name: rpc.GetName(),
		}
	}
	return &Service{
		Name: service.GetName(),
		RPCs: rpcs,
	}
}
