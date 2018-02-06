package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type Service struct {
	rpcs []entity.RPC

	d *desc.ServiceDescriptor
}

func newService(d *desc.ServiceDescriptor) *Service {
	return &Service{
		rpcs: newRPCs(d),
		d:    d,
	}
}

func (s *Service) Name() string {
	return s.d.GetName()
}

func (s *Service) FQRN() string {
	return s.d.GetFullyQualifiedName()
}

func (s *Service) RPCs() []entity.RPC {
	return s.rpcs
}
