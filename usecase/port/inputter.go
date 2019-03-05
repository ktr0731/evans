package port

import (
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
)

type Inputter interface {
	Input(req *desc.MessageDescriptor) (proto.Message, error)
}
