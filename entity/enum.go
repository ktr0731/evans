package entity

import "github.com/jhump/protoreflect/desc"

type Enum struct {
	Name   string
	Values []*Enum

	desc      *desc.EnumDescriptor
	fieldDesc *desc.FieldDescriptor
}

func newEnum(d *desc.EnumDescriptor) *Enum {
	// TODO: 間違ってそう
	values := make([]*Enum, len(d.GetValues()))
	for i, v := range d.GetValues() {
		values[i] = newEnum(v.GetEnum())
	}
	return &Enum{
		Name:   d.GetName(),
		Values: values,
		desc:   d,
	}
}

func newEnumAsField(f *desc.FieldDescriptor) *Enum {
	enum := newEnum(f.GetEnumType())
	enum.fieldDesc = f
	return enum
}

func (e *Enum) isField() {}

func (e *Enum) name() string {
	return e.fieldDesc.GetType().String()
}

func (e *Enum) typ() string {
	if e.fieldDesc == nil {
		return ""
	}
	return e.fieldDesc.GetType().String()
}
