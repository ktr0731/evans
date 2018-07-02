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

type messageBuilder struct {
	m *message
	d *desc.MessageDescriptor

	// used to detect cycle fields
	usedMessage map[string]entity.Message
}

func (b *messageBuilder) nextMessageField(f *desc.FieldDescriptor) {
	field := &messageField{
		d: f,
	}
	if m, ok := b.usedMessage[f.GetMessageType().GetName()]; ok {
		field.Message = m
		b.add(field)
		return
	}
	msg := &message{
		d: f.GetMessageType(),
	}
	b2 := &messageBuilder{
		m:           msg,
		d:           f.GetMessageType(),
		usedMessage: b.usedMessage,
	}
	field.Message = b2.build()
	b.usedMessage[f.GetMessageType().GetName()] = field.Message
	b.add(field)
}

func (b *messageBuilder) add(f entity.Field) {
	b.m.fields = append(b.m.fields, f)
}

func (b *messageBuilder) build() entity.Message {
	// collect messages and enums which declared in the target message

	msgs := make([]entity.Message, 0, len(b.d.GetNestedMessageTypes()))
	for _, d := range b.d.GetNestedMessageTypes() {
		msgs = append(msgs, newMessage(d))
	}
	b.m.nestedMessages = msgs

	enums := make([]entity.Enum, 0, len(b.d.GetNestedEnumTypes()))
	for _, d := range b.d.GetNestedEnumTypes() {
		enums = append(enums, newEnum(d))
	}
	b.m.nestedEnums = enums

	// it need to resolve oneofs before resolve other fields.
	// GetFields contains fields of oneofs.
	encounteredOneOfFields := map[string]bool{}
	for _, o := range b.d.GetOneOfs() {
		for _, c := range o.GetChoices() {
			encounteredOneOfFields[c.GetFullyQualifiedName()] = true
		}
		b.add(newOneOfField(o))
	}

	// TODO: label, options
	for _, f := range b.d.GetFields() {
		// skip fields of oneofs
		if encounteredOneOfFields[f.GetFullyQualifiedName()] {
			continue
		}

		if isMessageType(f.GetType()) {
			// self-referenced field
			if f.GetMessageType().GetName() == b.d.GetName() {
				b.add(&messageField{d: f, Message: b.m})
			} else {
				b.nextMessageField(f)
			}
		} else {
			b.add(newField(f))
		}
	}

	return b.m
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
