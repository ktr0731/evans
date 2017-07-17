package model

import (
	"bytes"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/olekukonko/tablewriter"
)

type Message struct {
	Name string
}

func NewMessage(message *descriptor.DescriptorProto) *Message {
	// TODO: resolving nested message
	return &Message{
		Name: message.GetName(),
	}
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
