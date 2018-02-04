package entity

import (
	"bytes"

	"github.com/jhump/protoreflect/desc"
	"github.com/olekukonko/tablewriter"
)

// TODO: MessageImpl と Message インターフェースを定義して、余計なメソッドを公開しないようにする
type Message struct {
	desc      *desc.MessageDescriptor
	fieldDesc *desc.FieldDescriptor // fieldDesc is nil if this message is not used as a field
}

func newMessage(m *desc.MessageDescriptor) *Message {
	return &Message{
		desc: m,
	}
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

func (m *Message) String() string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"field", "type"})
	rows := [][]string{}
	for _, f := range m.Fields {
		row := []string{f.Name(), f.Type()}
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
		row := []string{message.Name()}
		rows = append(rows, row)
	}
	table.AppendBulk(rows)
	table.Render()

	return buf.String()
}
