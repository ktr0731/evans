package port

import (
	"github.com/gogo/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
)

type Inputter interface {
	Input(reqType *desc.MessageDescriptor) (proto.Message, error)
}
