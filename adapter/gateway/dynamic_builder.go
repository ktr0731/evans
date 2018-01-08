package gateway

import (
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type DynamicBuilder struct{}

func (b *DynamicBuilder) NewMessage(md *desc.MessageDescriptor) proto.Message {
	return dynamic.NewMessage(md)
}
