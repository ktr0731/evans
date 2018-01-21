package port

import (
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
)

type DynamicBuilder interface {
	NewMessage(md *desc.MessageDescriptor) proto.Message
}
