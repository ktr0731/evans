package entity

import (
	"github.com/jhump/protoreflect/desc"
)

type Message struct {
	fields []*field

	desc      *desc.MessageDescriptor
	fieldDesc *desc.FieldDescriptor // fieldDesc is nil if this message is not used as a field
}

func newMessage(m *desc.MessageDescriptor) *Message {
	msg := Message{
		desc: m,
	}

	// TODO: label, map, options
	fields := make([]*field, len(m.GetFields()))
	for i, f := range m.GetFields() {
		fields[i] = &field{
			Name: f.GetName(),
			Type: f.GetType().String(),
		}
	}
	msg.fields = fields

	return &msg
}

func newMessageAsField(f *desc.FieldDescriptor) *Message {
	msg := newMessage(f.GetMessageType())
	msg.fieldDesc = f
	return msg
}

func (m *Message) isField() {}

func (m *Message) Name() string {
	if m.fieldDesc != nil {
		return m.fieldDesc.GetName()
	}
	return m.desc.GetName()
}

func (m *Message) Type() string {
	if m.fieldDesc == nil {
		return ""
	}
	return m.fieldDesc.GetType().String()
}

func (m *Message) Number() int32 {
	if m.fieldDesc == nil {
		return NON_FIELD
	}
	return m.fieldDesc.GetNumber()
}

func (m *Message) IsRepeated() bool {
	if m.fieldDesc == nil {
		return false
	}
	return m.fieldDesc.IsRepeated()
}

func (m *Message) Fields() []*field {
	return m.fields
}

type Messages []*Message
