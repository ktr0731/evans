package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type messageBuilder struct {
	m *message
	d *desc.MessageDescriptor

	// used to detect cycle fields
	usedMessage map[string]entity.Message
}

func (b *messageBuilder) buildMessageField(f *desc.FieldDescriptor) entity.Field {
	field := &messageField{
		d: f,
	}

	// self-referenced
	if f.GetMessageType().GetName() == b.m.Name() {
		b.m.isCycled = true
		field.Message = b.m
		return field
	}

	if m, ok := b.usedMessage[f.GetMessageType().GetName()]; ok {
		b.m.isCycled = true
		field.Message = m
		return field
	}

	b2 := &messageBuilder{
		m: &message{
			d: f.GetMessageType(),
		},
		d:           f.GetMessageType(),
		usedMessage: b.usedMessage,
	}

	field.Message = b2.build()

	b.usedMessage[f.GetMessageType().GetName()] = field.Message
	return field
}

func (b *messageBuilder) processMessageField(f *desc.FieldDescriptor) {
	b.add(b.buildMessageField(f))
}

func (b *messageBuilder) add(f entity.Field) {
	b.m.fields = append(b.m.fields, f)
}

func (b *messageBuilder) processOneOfField(d *desc.OneOfDescriptor) {
	choices := make([]entity.Field, 0, len(d.GetChoices()))
	for _, c := range d.GetChoices() {
		choices = append(choices, b.newField(c))
	}
	b.add(&oneOfField{
		choices: choices,
		d:       d,
	})
}

// TODO: naming
// root of all field type
func (b *messageBuilder) newField(d *desc.FieldDescriptor) entity.Field {
	var f entity.Field
	switch {
	case isMessageType(d.AsFieldDescriptorProto().GetType()):
		f = b.buildMessageField(d)
	case isEnumType(d):
		f = newEnumField(d)
	default: // primitive field
		f = newPrimitiveField(d)
	}
	return f
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
		b.processOneOfField(o)
	}

	// TODO: label, options
	for _, f := range b.d.GetFields() {
		// skip fields of oneofs
		if encounteredOneOfFields[f.GetFullyQualifiedName()] {
			continue
		}

		// self

		if isMessageType(f.GetType()) {
			b.processMessageField(f)
		} else {
			b.add(newField(f))
		}
	}

	return b.m
}
