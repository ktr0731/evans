package model

import (
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

func NewFields(getMessage func(msgName string) (*Message, error), msg *Message) ([]*Field, error) {
	var fields []*Field
	for _, field := range msg.Desc.GetFields() {
		f := &Field{
			Name: field.GetName(),
			Type: field.GetType(),
			Desc: field,
		}

		if field.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			f.IsMessage = true

			// TODO: 別パッケージの msg が取得できない
			msg, err := getMessage(field.GetMessageType().GetName())
			if err != nil {
				return nil, err
			}

			f.Fields, err = NewFields(getMessage, msg)
		}

		fields = append(fields, f)
	}
	return fields, nil
}
