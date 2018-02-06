package protobuf

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type enumValueField struct {
	d *desc.FieldDescriptor
}

func newEnumField(d *desc.FieldDescriptor) entity.EnumField {
	return &enumValueField{
		d: d,
	}
}

func (e *enumValueField) FieldName() string {
	return e.d.GetName()
}

func (e *enumValueField) FQRN() string {
	return e.d.GetFullyQualifiedName()
}

func (e *enumValueField) Type() entity.FieldType {
	return entity.FieldTypeEnum
}

func (e *enumValueField) IsRepeated() bool {
	return false
}

func (e *enumValueField) PBType() string {
	return descriptor.FieldDescriptorProto_TYPE_ENUM
}

func (e *enumValueField) Number() int32 {
	return e.d.GetNumber()
}
