package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

func newMessageField(f *desc.FieldDescriptor) entity.MessageField {
	msg := newMessage(f.GetMessageType())
	msg.fieldDesc = f
	return msg
}

func (m *message) Name() string {
	if m.fieldDesc != nil {
		return m.fieldDesc.GetName()
	}
	return m.desc.GetName()
}
func (m *message) Number() int32 {
	if m.fieldDesc == nil {
		return NON_FIELD
	}
	return m.fieldDesc.GetNumber()
}

func (m *message) IsRepeated() bool {
	if m.fieldDesc == nil {
		return false
	}
	return m.fieldDesc.IsRepeated()
}
