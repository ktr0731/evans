package entity

import "github.com/jhump/protoreflect/desc"

type OneOf struct {
	Name    string
	Choices []field

	desc      *desc.OneOfDescriptor
	fieldDesc *desc.FieldDescriptor
}

func (o *OneOf) isField() {}

func newOneOf(d *desc.OneOfDescriptor) *OneOf {
	choices := make([]field, len(d.GetChoices()))
	for i, c := range d.GetChoices() {
		choices[i] = newField(c)
	}
	return &OneOf{
		Name:    d.GetName(),
		Choices: choices,
		desc:    d,
	}
}

func newOneOfAsField(f *desc.FieldDescriptor) *OneOf {
	oneOf := newOneOf(f.GetOneOf())
	oneOf.fieldDesc = f
	return oneOf
}

func (o *OneOf) name() string {
	return o.Name
}

func (o *OneOf) typ() string {
	if o.fieldDesc == nil {
		return ""
	}
	return o.fieldDesc.GetType().String()
}
