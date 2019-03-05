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
	return newMessageBuilder(d).build()
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

func (m *message) Desc() *desc.MessageDescriptor {
	return m.d
}
