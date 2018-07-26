package protobuf

import (
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
)

// NewDynamicMessage is used from DynamicBuilder
// for extract *desc.MessageDescriptor from m
func NewDynamicMessage(m entity.Message) proto.Message {
	msg := m.(*message)
	return dynamic.NewMessage(msg.d)
}
