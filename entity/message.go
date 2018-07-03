package entity

type Message interface {
	Name() string
	Fields() []Field
	IsCycled() bool
}
