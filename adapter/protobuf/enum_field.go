package protobuf

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type enumField struct {
	d      *desc.FieldDescriptor
	values []entity.EnumValue
}

func newEnumField(d *desc.FieldDescriptor) entity.EnumField {
	values := make([]entity.EnumValue, 0, len(d.GetEnumType().GetValues()))
	for _, v := range d.GetEnumType().GetValues() {
		values = append(values, newEnumValue(v))
	}
	return &enumField{
		d:      d,
		values: values,
	}
}

func (e *enumField) Name() string {
	m := e.d.GetParent().(*desc.MessageDescriptor)
	return m.GetName()
}

func (e *enumField) FieldName() string {
	return e.d.GetName()
}

func (e *enumField) FQRN() string {
	return e.d.GetFullyQualifiedName()
}

func (e *enumField) Type() entity.FieldType {
	return entity.FieldTypeEnum
}

func (e *enumField) IsRepeated() bool {
	return e.d.IsRepeated()
}
func (e *enumField) IsCycled() bool {
	return false
}

func (e *enumField) PBType() string {
	return descriptor.FieldDescriptorProto_TYPE_ENUM.String()
}

func (e *enumField) Values() []entity.EnumValue {
	return e.values
}
