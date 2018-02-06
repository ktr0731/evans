package entity

import (
	"github.com/jhump/protoreflect/desc"
)

type IService interface {
	Name() string
	FQRN() string
	RPCs() []IRPC
}

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

type Services []*Service
