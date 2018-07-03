package protobuf

import (
	"fmt"

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

type messageBuilder struct {
	m *message
	d *desc.MessageDescriptor

	dependOn map[string]bool

	// used to detect cycle fields
	usedMessage map[string]entity.Message
}

func (b *messageBuilder) processMessageField(f *desc.FieldDescriptor) {
	field := &messageField{
		d: f,
	}
	// self-referenced
	if f.GetMessageType().GetName() == b.m.Name() {
		b.m.isCycled = true
		field.Message = b.m
		b.add(field)
		return
	} else if m, ok := b.usedMessage[f.GetMessageType().GetName()]; ok {
		b.m.isCycled = true
		field.Message = m
		fmt.Printf("%#v %v\n", field, m)
		b.add(field)
		return
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

		// self

		if isMessageType(f.GetType()) {
			b.processMessageField(f)
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

func (m *message) IsCycled() bool {
	return m.isCycled
}
