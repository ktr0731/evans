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

	// it need to resolve oneofs before resolve other fields.
	// GetFields contains fields of oneofs.
	encounteredOneOfFields := map[string]bool{}
	fields := make([]entity.Field, 0, len(m.GetFields()))
	for _, o := range m.GetOneOfs() {
		for _, c := range o.GetChoices() {
			encounteredOneOfFields[c.GetFullyQualifiedName()] = true
		}
		fields = append(fields, newOneOfField(o))
	}

	// TODO: label, map, options
	for _, f := range m.GetFields() {
		// skip fields of oneofs
		if encounteredOneOfFields[f.GetFullyQualifiedName()] {
			continue
		}

		// self-referenced field
		if isMessageType(f.GetType()) && f.GetMessageType().GetName() == m.GetName() {
			fields = append(fields, &messageField{d: f, Message: &msg})
		} else {
			fields = append(fields, newField(f))
		}
	}
	msg.fields = fields

	return &msg
}

func (m *message) Name() string {
	return m.d.GetName()
}

func (m *message) Fields() []entity.Field {
	return m.fields
}
