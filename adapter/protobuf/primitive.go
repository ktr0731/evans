package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type primitiveField struct {
	desc *desc.FieldDescriptor
}

func newPrimitiveField(f *desc.FieldDescriptor) entity.PrimitiveField {
	return &primitiveField{
		desc: f,
	}
}

func (f *primitiveField) isField() {}

func (f *primitiveField) Name() string {
	return f.desc.GetName()
}

func (f *primitiveField) Type() string {
	return f.desc.GetType().String()
}

func (f *primitiveField) Number() int32 {
	return f.desc.GetNumber()
}

func (f *primitiveField) IsRepeated() bool {
	return f.desc.IsRepeated()
}
