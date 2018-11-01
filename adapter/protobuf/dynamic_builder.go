package protobuf

import (
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
)

type DynamicBuilder struct{}

func NewDynamicBuilder() *DynamicBuilder {
	return &DynamicBuilder{}
}

func (b *DynamicBuilder) NewMessage(m entity.Message) proto.Message {
	msg := m.(*message)
	return dynamic.NewMessage(msg.d)
}
