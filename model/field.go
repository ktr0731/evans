package model

import (
	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/k0kubun/pp"
)

type Field struct {
	Name     string
	JSONName string
	Type     descriptor.FieldDescriptorProto_Type
	TypeName string
	Default  string

	IsMessage bool
	Fields    []*Field
}

func NewFields(getMessage func(typeName string) *descriptor.DescriptorProto, msg *descriptor.DescriptorProto) []*Field {
	var fields []*Field
	for _, field := range msg.GetField() {
		f := &Field{
			Name:     field.GetName(),
			JSONName: field.GetJsonName(),
			Type:     field.GetType(),
			TypeName: field.GetTypeName(),
			Default:  field.GetDefaultValue(),
		}

		if field.Type.String() == "TYPE_MESSAGE" {
			f.IsMessage = true

			fmt.Println("さいき")
			msg := getMessage(field.GetTypeName())
			pp.Println(msg)
			f.Fields = NewFields(getMessage, msg)
		}

		fields = append(fields, f)
	}
	return fields
}
