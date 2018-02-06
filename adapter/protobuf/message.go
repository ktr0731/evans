package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type message struct {
	d *desc.MessageDescriptor

	fields []entity.Field
	oneOfs []entity.OneOfField

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

	// TODO: label, map, options
	fields := make([]entity.Field, 0, len(m.GetFields()))
	for _, f := range m.GetFields() {
		// self-referenced field
		if IsMessageType(f.GetType()) && f.GetMessageType().GetName() == m.GetName() {
			fields = append(fields, &msg)
		} else {
			fields = append(fields, newField(f))
		}
	}
	msg.fields = fields

	oneOfs := make([]entity.OneOfField, 0, len(m.GetOneOfs()))
	for _, o := range m.GetOneOfs() {
		oneOfs = append(oneOfs, newOneOf(o))
	}
	msg.oneOfs = oneOfs

	return &msg
}

func (m *message) Name() string {
	return m.d.GetName()
}

func (m *message) Fields() []entity.Field {
	return m.fields
}
