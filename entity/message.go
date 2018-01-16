package entity

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

type Message struct {
	Name   string
	Fields []field
	OneOfs []*OneOf
	Nested Messages

	desc *desc.MessageDescriptor
}

func NewMessage(m *desc.MessageDescriptor) *Message {
	msg := Message{
		Name: m.GetName(),
		desc: m,
	}

	nested := m.GetNestedMessageTypes()
	nestedMsgs := make(Messages, len(nested))
	for i, d := range nested {
		nestedMsgs[i] = NewMessage(d)
	}
	msg.Nested = nestedMsgs

	// TODO: label, map, options
	fields := make([]field, len(m.GetFields()))
	for i, f := range m.GetFields() {
		fields[i] = newField(f)
	}
	msg.Fields = fields

	oneOfs := make([]*OneOf, len(m.GetOneOfs()))
	for i, o := range m.GetOneOfs() {
		choices := make([]field, len(o.GetChoices()))
		for j, c := range o.GetChoices() {
			choices[j] = newField(c)
		}
		oneOfs[i] = newOneOf(o.GetName(), choices, o)
	}
	msg.OneOfs = oneOfs

	return &msg
}

func (m *Message) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"field", "type"})
	rows := [][]string{}
	for _, field := range m.Fields {
		fType := field.Type.String()
		row := []string{field.Name, fType}
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
