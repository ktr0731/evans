package model

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

type Message struct {
	Name   string
	Fields []*Field
}

func NewMessage(message *desc.MessageDescriptor) *Message {
	var msg Message
	msg.Name = message.GetName()
	return &msg
}

func (m *Message) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"field", "type"})
	rows := [][]string{}
	for _, field := range m.Fields {
		fType := field.Type.String()
		if field.TypeName != "" {
			fType = field.TypeName
		}
		row := []string{field.JSONName, fType}
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
