package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type messageField struct {
	d *desc.FieldDescriptor
	entity.Message
}

func (f *messageField) FieldName() string {
	return f.d.GetName()
}

func (f *messageField) FQRN() string {
	return f.d.GetFullyQualifiedName()
}

func (f *messageField) Type() entity.FieldType {
	return entity.FieldTypeMessage
}

func (f *messageField) IsRepeated() bool {
	return f.d.IsRepeated()
}

func (f *messageField) PBType() string {
	return f.d.GetType().String()
}
