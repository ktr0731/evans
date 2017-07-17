package model

import (
	"bytes"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/k0kubun/pp"
	"github.com/olekukonko/tablewriter"
)

type Field struct {
	Name     string
	JSONName string
	Type     descriptor.FieldDescriptorProto_Type
	TypeName string
	Default  string
}

type Message struct {
	Name   string
	Fields []*Field
}

func NewMessage(message *descriptor.DescriptorProto) *Message {
	var fields []*Field
	for _, field := range message.GetField() {
		pp.Println(field)
		fields = append(fields, &Field{
			Name:     field.GetName(),
			JSONName: field.GetJsonName(),
			Type:     field.GetType(),
			TypeName: field.GetTypeName(),
			Default:  field.GetDefaultValue(),
		})
	}
	return &Message{
		Name:   message.GetName(),
		Fields: fields,
	}
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
