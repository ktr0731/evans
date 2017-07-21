package model

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
)

// TODO: モデルの設計が冗長
type Field struct {
	Name    string
	Type    descriptor.FieldDescriptorProto_Type
	Default string
	Desc    *desc.FieldDescriptor

	IsMessage bool
	Fields    []*Field
}

func NewFields(getMessage func(msgName string) *desc.MessageDescriptor, msg *desc.MessageDescriptor) []*Field {
	var fields []*Field
	for _, field := range msg.GetFields() {
		fmt.Println(field.GetName())
		f := &Field{
			Name: field.GetName(),
			Type: field.GetType(),
			Desc: field,
		}

		if field.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			f.IsMessage = true

			msg := getMessage(field.GetName())
			if msg == nil {
				return nil
			}

			f.Fields = NewFields(getMessage, msg)
		}

		fields = append(fields, f)
	}
	return fields
}
