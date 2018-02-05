package entity

import "github.com/jhump/protoreflect/desc"

type OneOf struct {
	Choices []field

	desc *desc.OneOfDescriptor
}

// TODO: Show で oneof かどうかの表示をしたい

// func newOneOf(d *desc.OneOfDescriptor) *OneOf {
// 	choices := make([]field, len(d.GetChoices()))
// 	for i, c := range d.GetChoices() {
// 		choices[i] = newField(c)
// 	}
// 	return &OneOf{
// 		Choices: choices,
// 		desc:    d,
// 	}
// }
//
// func (o *OneOf) Name() string {
// 	return o.desc.GetName()
// }
