package entity

import "github.com/jhump/protoreflect/desc"

type Message interface {
	Name() string
	Fields() []Field
	IsCycled() bool
	// NOTE: Desc is a temporary API for refactoring.
	Desc() *desc.MessageDescriptor
}
