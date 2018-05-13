package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
	"google.golang.org/grpc"
)

type rpc struct {
	d                 *desc.MethodDescriptor
	sd                *grpc.StreamDesc
	requestMessage    entity.Message
	responseMessage   entity.Message
	isServerStreaming bool
	isClientStreaming bool
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

func (r *rpc) IsServerStreaming() bool {
	return r.isServerStreaming
}

func (r *rpc) IsClientStreaming() bool {
	return r.isClientStreaming
}

func (r *rpc) StreamDesc() *grpc.StreamDesc {
	if r.sd == nil {
		panic("receiver RPC is not streaming RPC")
	}
	return r.sd
}

func newRPCs(svc *desc.ServiceDescriptor) []entity.RPC {
	rpcs := make([]entity.RPC, 0, len(svc.GetMethods()))
	for _, m := range svc.GetMethods() {
		r := &rpc{
			d:                 m,
			requestMessage:    newMessage(m.GetInputType()),
			responseMessage:   newMessage(m.GetOutputType()),
			isServerStreaming: m.IsServerStreaming(),
			isClientStreaming: m.IsClientStreaming(),
		}
		if r.IsClientStreaming() || r.IsServerStreaming() {
			r.sd = &grpc.StreamDesc{
				StreamName:    m.GetName(),
				ServerStreams: m.IsServerStreaming(),
				ClientStreams: m.IsClientStreaming(),
			}
		}
		rpcs = append(rpcs, r)
	}
	return rpcs
}
