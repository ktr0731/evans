package entity

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

type Message struct {
	Name           string
	Fields         []field
	OneOfs         []*OneOf
	NestedMessages Messages
	NestedEnums    []*Enum

	desc      *desc.MessageDescriptor
	fieldDesc *desc.FieldDescriptor // fieldDesc is nil if this message is not used as a field
}

func NewMessage(m *desc.MessageDescriptor) *Message {
	msg := Message{
		Name: m.GetName(),
		desc: m,
	}

	msgs := make(Messages, len(m.GetNestedMessageTypes()))
	for i, d := range m.GetNestedMessageTypes() {
		msgs[i] = NewMessage(d)
	}
	msg.NestedMessages = msgs

	enums := make([]*Enum, len(m.GetNestedEnumTypes()))
	for i, d := range m.GetNestedEnumTypes() {
		enums[i] = newEnum(d)
	}
	msg.NestedEnums = enums

	// TODO: label, map, options
	fields := make([]field, len(m.GetFields()))
	for i, f := range m.GetFields() {
		fields[i] = newField(f)
	}
	msg.Fields = fields

	oneOfs := make([]*OneOf, len(m.GetOneOfs()))
	for i, o := range m.GetOneOfs() {
		oneOfs[i] = newOneOf(o)
	}
	msg.OneOfs = oneOfs

	return &msg
}

func newMessageAsField(f *desc.FieldDescriptor) *Message {
	msg := NewMessage(f.GetMessageType())
	msg.fieldDesc = f
	return msg
}

func (m *Message) isField() {}

func (m *Message) name() string {
	return m.Name
}

func (m *Message) typ() string {
	if m.fieldDesc == nil {
		return ""
	}
	return m.fieldDesc.GetType().String()
}

func (m *Message) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"field", "type"})
	rows := [][]string{}
	for _, f := range m.Fields {
		row := []string{f.name(), f.typ()}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}

type Messages []*Message

func (m Messages) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"message"})
	rows := [][]string{}
	for _, message := range m {
		row := []string{message.Name}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
