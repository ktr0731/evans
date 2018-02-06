package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type enum struct {
	values []entity.EnumValue

	d *desc.EnumDescriptor
}

func newEnum(d *desc.EnumDescriptor) entity.Enum {
	values := make([]entity.EnumValue, len(d.GetValues()))
	for _, v := range d.GetValues() {
		values = append(values, newEnumValue(v))
	}
	return &enum{
		values: values,
		d:      d,
	}
}

func (e *enum) Name() string {
	return e.d.GetName()
}

func (e *enum) Values() []entity.EnumValue {
	return e.values
}

// enum value is the field in enum declaration
type enumValue struct {
	d *desc.EnumValueDescriptor
}

func newEnumValue(d *desc.EnumValueDescriptor) entity.EnumValue {
	return &enumValue{
		d: d,
	}
}

func (e *enumValue) Name() string {
	return e.d.GetName()
}

func (e *enumValue) Number() int32 {
	return e.d.GetNumber()
}
