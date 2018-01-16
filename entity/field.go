package entity

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
	// key is FQN of message, so it can extract by field.GetMessageType().GetFullyQualifiedName()
	localMessageCache := map[string]*Message{}
	resolveLocalMessage(localMessageCache, msg.desc)

	for _, field := range msg.desc.GetFields() {
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

			msg, ok = localMessageCache[field.GetMessageType().GetFullyQualifiedName()]
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

func resolveLocalMessage(cache map[string]*Message, msg *desc.MessageDescriptor) {
	nested := msg.GetNestedMessageTypes()
	if len(nested) == 0 {
		cache[msg.GetFullyQualifiedName()] = &Message{
			Name: msg.GetName(),
			desc: msg,
		}
		return
	}

	// 効率悪そう
	for _, d := range nested {
		resolveLocalMessage(cache, d)
	}
}
