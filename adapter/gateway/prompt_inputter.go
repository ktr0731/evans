package gateway

import (
	"github.com/gogo/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type PromptInputter struct {
	env entity.Environment
}

func NewPromptInputter(env entity.Environment) *PromptInputter {
	return &PromptInputter{env}
}

func (i *PromptInputter) Input(reqType *desc.MessageDescriptor) (proto.Message, error) {
	// fields := reqType.GetFields()
	// req := dynamic.NewMessage(reqType)
	return nil, nil
}

func (i *PromptInputter) inputFields() error {
	return nil
}
