package entity

import (
	"github.com/jhump/protoreflect/desc"
)

const (
	NON_FIELD = int32(0)
)

// fieldable types:
//	enum, oneof, message
type field interface {
	isField()

	Name() string
	Number() int32
	Type() string
}

func newField(desc *desc.FieldDescriptor) field {
	var f field
	switch {
	case IsMessageType(desc.AsFieldDescriptorProto().GetType()):
		f = newMessageAsField(desc)
	case IsEnumType(desc):
		f = newEnumAsField(desc)
	default: // primitive field
		f = newPrimitiveField(desc)
	}
	return f
}
