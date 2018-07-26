package protobuf

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
)

func isMessageType(typeName descriptor.FieldDescriptorProto_Type) bool {
	return typeName == descriptor.FieldDescriptorProto_TYPE_MESSAGE
}

func isEnumType(f *desc.FieldDescriptor) bool {
	return f.GetEnumType() != nil
}
