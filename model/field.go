package model

import (
	"fmt"
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
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

func NewFields(msg *descriptor.DescriptorProto) []*Field {
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
			fmt.Println("msg!")

			f.IsMessage = true

			var desc descriptor.DescriptorProto
			data, _ := field.Descriptor()
			if err := proto.UnmarshalMerge(data, &desc); err != nil {
				log.Fatal(err)
				return nil
			}
			f.Fields = NewFields(&desc)
		}

		fields = append(fields, f)
	}
	return fields
}
