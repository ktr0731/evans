package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type primitiveField struct {
	d *desc.FieldDescriptor
}

func newPrimitiveField(f *desc.FieldDescriptor) entity.PrimitiveField {
	return &primitiveField{
		d: f,
	}
}

func (f *primitiveField) FieldName() string {
	return f.d.GetName()
}

func (f *primitiveField) FQRN() string {
	return f.d.GetFullyQualifiedName()
}

func (f *primitiveField) Type() string {
	return f.d.GetType().String()
}

func (f *primitiveField) Number() int32 {
	return f.d.GetNumber()
}

func (f *primitiveField) IsRepeated() bool {
	return f.d.IsRepeated()
}
