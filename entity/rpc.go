package entity

import (
	"github.com/jhump/protoreflect/desc"
)

type RPC struct {
	Name         string
	FQRN         string
	RequestType  *desc.MessageDescriptor
	ResponseType *desc.MessageDescriptor
}

type RPCs []*RPC

// NewRPCs collects RPCs in ServiceDescriptor.
// Only NewRPCs receive a raw descriptor because it is called by NewService.
// So, NewRPCs needs to receive a raw descriptor instead of entity.Service.
func newRPCs(svc *desc.ServiceDescriptor) RPCs {
	rpcs := make(RPCs, len(svc.GetMethods()))
	for i, rpc := range svc.GetMethods() {
		r := &RPC{
			Name:         rpc.GetName(),
			FQRN:         rpc.GetFullyQualifiedName(),
			RequestType:  rpc.GetInputType(),
			ResponseType: rpc.GetOutputType(),
		}
		rpcs[i] = r
	}
	return rpcs
}
