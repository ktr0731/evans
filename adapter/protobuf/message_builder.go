package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

// messageBuilder builds entity.Message from *desc.MessageDescriptor.
// process* methods are used to convert each field descriptors to entity.Field.
type messageBuilder struct {
	m *message
	d *desc.MessageDescriptor

	// used to detect cycle fields.
	// if top-level message has message fields, messageBuilder calls newMessageBuilder recursively.
	// at that time, new messageBuilder takes over the value of usedMessage of top-level messageBuilder.
	usedMessage map[string]entity.Message
}

func newMessageBuilder(d *desc.MessageDescriptor) *messageBuilder {
	msg := &message{
		d:      d,
		fields: make([]entity.Field, 0, len(d.GetFields())),
	}
	usedMessage := make(map[string]entity.Message)
	usedMessage[msg.Name()] = msg
	return &messageBuilder{
		m:           msg,
		d:           d,
		usedMessage: usedMessage,
	}
}

func (b *messageBuilder) add(f entity.Field) {
	b.m.fields = append(b.m.fields, f)
}

// maps are also interpret as a repeated message type.
// ref. https://developers.google.com/protocol-buffers/docs/proto#backwards-compatibility
func (b *messageBuilder) processMessageField(f *desc.FieldDescriptor) entity.Field {
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

	b2 := newMessageBuilder(f.GetMessageType())
	b2.usedMessage = b.usedMessage
	b2.usedMessage[b2.m.Name()] = b2.m

	field.Message = b2.build()

	b.usedMessage[f.GetMessageType().GetName()] = field.Message
	return field
}

func (b *messageBuilder) addOneOfField(d *desc.OneOfDescriptor) {
	choices := make([]entity.Field, 0, len(d.GetChoices()))
	for _, c := range d.GetChoices() {
		choices = append(choices, b.processField(c))
	}
	b.add(&oneOfField{
		choices: choices,
		d:       d,
	})
}

// processField converts passed field which is other than oneOfFields to entity.Field.
func (b *messageBuilder) processField(d *desc.FieldDescriptor) entity.Field {
	var f entity.Field
	switch {
	case isMessageType(d.AsFieldDescriptorProto().GetType()):
		f = b.processMessageField(d)
	case isEnumType(d):
		f = newEnumField(d)
	default: // primitive field
		f = newPrimitiveField(d)
	}
	return f
}

func (b *messageBuilder) addField(d *desc.FieldDescriptor) {
	b.add(b.processField(d))
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
	// because GetFields contains fields of oneofs.
	encounteredOneOfFields := map[string]bool{}
	for _, o := range b.d.GetOneOfs() {
		for _, c := range o.GetChoices() {
			encounteredOneOfFields[c.GetFullyQualifiedName()] = true
		}
		b.addOneOfField(o)
	}

	// TODO: label, options
	for _, f := range b.d.GetFields() {
		// skip fields of oneofs
		if encounteredOneOfFields[f.GetFullyQualifiedName()] {
			continue
		}

		b.addField(f)
	}

	return b.m
}
