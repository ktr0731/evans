package protobuf

import (
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
)

// NewDynamicMessage is used from DynamicBuilder
// for extract *desc.MessageDescriptor from msg
func NewDynamicMessage(msg entity.Message) proto.Message {
	m := msg.(*message)
	return dynamic.NewMessage(m.d)
}
