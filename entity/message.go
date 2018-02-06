package entity

type IMessage interface {
	Name() string
	Fields() []Field
}
