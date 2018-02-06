package protobuf

import "github.com/jhump/protoreflect/desc"

type Service struct {
	Name string
	RPCs RPCs

	desc *desc.ServiceDescriptor
}

func newService(d *desc.ServiceDescriptor) *Service {
	return &Service{
		Name: d.GetName(),
		RPCs: newRPCs(d),
		desc: d,
	}
}
