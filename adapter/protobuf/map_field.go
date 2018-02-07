package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type mapField struct {
	d        *desc.FieldDescriptor
	key, val entity.Field
}

func newMapField(d *desc.FieldDescriptor) entity.MapField {
	return &mapField{
		d: d,
	}
}

func (f *mapField) FieldName() string {
	return f.d.GetName()
}

func (f *mapField) FQRN() string {
	return f.d.GetFullyQualifiedName()
}

func (f *mapField) Type() entity.FieldType {
	return entity.FieldTypeMap
}

// we parse mapField as a message field in fieldInputter.inputField
// https://developers.google.com/protocol-buffers/docs/proto3#backwards-compatibility
func (f *mapField) IsRepeated() bool {
	return true
}

func (f *mapField) PBType() string {
	return f.d.GetType().String()
}

func (f *mapField) Key() entity.Field {
	return f.key
}

func (f *mapField) Val() entity.Field {
	return f.val
}
