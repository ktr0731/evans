package protobuf

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

type oneOfField struct {
	d       *desc.OneOfDescriptor
	choices []entity.Field
}

func newOneOfField(d *desc.OneOfDescriptor) entity.OneOfField {
	choices := make([]entity.Field, 0, len(d.GetChoices()))
	for _, c := range d.GetChoices() {
		choices = append(choices, newField(c))
	}
	return &oneOfField{
		choices: choices,
		d:       d,
	}
}

func (o *oneOfField) FieldName() string {
	return o.d.GetName()
}

func (o *oneOfField) FQRN() string {
	return o.d.GetFullyQualifiedName()
}

func (o *oneOfField) Type() entity.FieldType {
	return entity.FieldTypeOneOf
}

func (o *oneOfField) IsRepeated() bool {
	return false
}

// oneof hasn't any type
func (o *oneOfField) PBType() string {
	return "oneof"
}

func (o *oneOfField) Choices() []entity.Field {
	return o.choices
}
