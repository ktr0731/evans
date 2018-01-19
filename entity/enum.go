package entity

import "github.com/jhump/protoreflect/desc"

type Enum struct {
	Values []*EnumValue

	desc      *desc.EnumDescriptor
	fieldDesc *desc.FieldDescriptor
}

type EnumValue struct {
	Name   string
	Number int32

	desc *desc.EnumValueDescriptor
}

func newEnum(d *desc.EnumDescriptor) *Enum {
	values := make([]*EnumValue, len(d.GetValues()))
	for i, v := range d.GetValues() {
		values[i] = newEnumValue(v)
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

func (e *Enum) Name() string {
	return e.desc.GetName()
}

func (e *Enum) Type() string {
	if e.fieldDesc == nil {
		return ""
	}
	return e.fieldDesc.GetType().String()
}

func (e *Enum) Number() int32 {
	if e.fieldDesc == nil {
		return NON_FIELD
	}
	return e.fieldDesc.GetNumber()
}

func newEnumValue(v *desc.EnumValueDescriptor) *EnumValue {
	return &EnumValue{
		Name:   v.GetName(),
		Number: v.GetNumber(),
		desc:   v,
	}
}
