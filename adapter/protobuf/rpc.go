package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type rpc struct {
	d               *desc.MethodDescriptor
	requestMessage  entity.Message
	responseMessage entity.Message
}

func (r *rpc) Name() string {
	return r.d.GetName()
}

func (r *rpc) FQRN() string {
	return r.d.GetFullyQualifiedName()
}

func (r *rpc) RequestMessage() entity.Message {
	return r.requestMessage
}

func (r *rpc) ResponseMessage() entity.Message {
	return r.responseMessage
}

func newRPCs(svc *desc.ServiceDescriptor) []entity.RPC {
	rpcs := make([]entity.RPC, 0, len(svc.GetMethods()))
	for _, r := range svc.GetMethods() {
		r := &rpc{
			d:               r,
			requestMessage:  newMessage(r.GetInputType()),
			responseMessage: newMessage(r.GetOutputType()),
		}
		rpcs = append(rpcs, r)
	}
	return rpcs
}
