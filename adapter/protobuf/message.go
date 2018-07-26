package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type message struct {
	d *desc.MessageDescriptor

	fields []entity.Field

	nestedMessages []entity.Message
	nestedEnums    []entity.Enum

	isCycled bool
}

func newMessage(d *desc.MessageDescriptor) entity.Message {
	msg := &message{
		d:      d,
		fields: make([]entity.Field, 0, len(d.GetFields())),
	}
	usedMessage := make(map[string]entity.Message)
	usedMessage[msg.Name()] = msg
	b := &messageBuilder{
		m:           msg,
		d:           d,
		usedMessage: usedMessage,
	}
	return b.build()
}

func (m *message) Name() string {
	return m.d.GetName()
}

func (m *message) Fields() []entity.Field {
	return m.fields
}

func (m *message) IsCycled() bool {
	return m.isCycled
}
