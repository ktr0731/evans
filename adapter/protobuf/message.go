package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type message struct {
	d *desc.MessageDescriptor

	nestedMessages []entity.Message
	nestedEnums    []entity.Enum
}

func newMessage(m *desc.MessageDescriptor) entity.Message {
	msg := message{
		d: m,
	}

	msgs := make([]entity.Message, 0, len(m.GetNestedMessageTypes()))
	for _, d := range m.GetNestedMessageTypes() {
		msgs = append(msgs, newMessage(d))
	}
	msg.nestedMessages = msgs

	enums := make([]entity.Enum, 0, len(m.GetNestedEnumTypes()))
	for _, d := range m.GetNestedEnumTypes() {
		enums = append(enums, newEnum(d))
	}
	msg.nestedEnums = enums

	return &msg
}

func (m *message) Name() string {
	return m.d.GetName()
}

func (m *message) Fields() []entity.Field {
	return m.fields
}
