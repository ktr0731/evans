package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type messageField struct {
	d *desc.FieldDescriptor

	fields []entity.Field
	oneOfs []entity.OneOfField
}

func newMessageField(m *desc.FieldDescriptor) entity.MessageField {
	msg := &messageField{
		d: m,
	}

	mt := m.GetMessageType()

	// TODO: label, map, options
	fields := make([]entity.Field, 0, len(mt.GetFields()))
	for _, f := range mt.GetFields() {
		// self-referenced field
		if isMessageType(f.GetType()) && f.GetMessageType().GetName() == m.GetName() {
			fields = append(fields, msg)
		} else {
			fields = append(fields, newField(f))
		}
	}
	msg.fields = fields

	oneOfs := make([]entity.OneOfField, 0, len(mt.GetOneOfs()))
	for _, o := range mt.GetOneOfs() {
		oneOfs = append(oneOfs, newOneOf(o))
	}
	msg.oneOfs = oneOfs

	return msg
}

func (f *messageField) FieldName() string {
	return f.d.GetName()
}

func (f *messageField) FQRN() string {
	return f.d.GetFullyQualifiedName()
}

func (f *messageField) Number() int32 {
	return f.d.GetNumber()
}

func (f *messageField) IsRepeated() bool {
	return f.d.IsRepeated()
}

func (f *messageField) PBType() string {
	return f.d.GetType().String()
}
