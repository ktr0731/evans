package entity

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
)

// fieldable types:
//	enum, oneof, message
type field interface {
	isField()

	name() string
	typ() string
}

func newField(desc *desc.FieldDescriptor) field {
	var f field
	switch {
	case IsMessageType(desc.AsFieldDescriptorProto().GetType()):
		f = newMessageAsField(desc)
	case IsEnumType(desc):
		f = newEnumAsField(desc)
	case IsOneOf(desc):
		f = newOneOfAsField(desc)
	default:
		panic(fmt.Sprintf("unsupported type: %s", desc.GetFullyQualifiedJSONName()))
	}
	return f
}
