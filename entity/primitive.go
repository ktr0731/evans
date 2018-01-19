package entity

import "github.com/jhump/protoreflect/desc"

type primitiveField struct {
	desc *desc.FieldDescriptor
}

func newPrimitiveField(f *desc.FieldDescriptor) *primitiveField {
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
