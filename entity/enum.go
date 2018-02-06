package entity

type Enum interface {
	Name() string
	Values() []EnumValue
}

type EnumValue interface {
	Name() string
	Number() int32
}
