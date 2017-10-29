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

func NewFields(pkg *Package, msg *Message) ([]*Field, error) {
	var fields []*Field

	// inner message definitions
	localMessageCache := map[string]*Message{}
	for _, d := range msg.Desc.GetNestedMessageTypes() {
		localMessageCache[d.GetName()] = &Message{
			Name: d.GetName(),
			Desc: d,
		}
	}

	for _, field := range msg.Desc.GetFields() {
		f := &Field{
			Name: field.GetName(),
			Type: field.GetType(),
			Desc: field,
		}

		if field.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			f.IsMessage = true

			var msg *Message
			var ok bool
			var err error
			msg, ok = localMessageCache[field.GetMessageType().GetName()]
			if !ok {
				// TODO: 別パッケージの msg が取得できない
				msg, err = pkg.GetMessage(field.GetMessageType().GetName())
				if err != nil {
					return nil, err
				}
			}

			f.Fields, err = NewFields(pkg, msg)
		}

		fields = append(fields, f)
	}
	return fields, nil
}
