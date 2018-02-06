package protobuf

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

func newField(desc *desc.FieldDescriptor) entity.Field {
	var f entity.Field
	switch {
	case isMessageType(desc.AsFieldDescriptorProto().GetType()):
		f = newMessageField(desc)
	case isEnumType(desc):
		f = newEnumField(desc)
	default: // primitive field
		f = newPrimitiveField(desc)
	}
	return f
}

func isMessageType(typeName descriptor.FieldDescriptorProto_Type) bool {
	return typeName == descriptor.FieldDescriptorProto_TYPE_MESSAGE
}

func isEnumType(f *desc.FieldDescriptor) bool {
	return f.GetEnumType() != nil
}
