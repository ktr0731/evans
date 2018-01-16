package entity

import "github.com/jhump/protoreflect/desc"

type OneOf struct {
	Name    string
	Choices []field

	desc *desc.OneOfDescriptor
}

func newOneOf(name string, choices []field, desc *desc.OneOfDescriptor) *OneOf {
	return &OneOf{
		Name:    name,
		Choices: choices,
		desc:    desc,
	}
}
