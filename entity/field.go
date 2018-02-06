package entity

const (
	NON_FIELD = int32(0)
)

// fieldable types:
//	enum, oneof, message, primitive
type field struct {
	Name, Type string
}

///////////////

type FieldType int

const (
	FieldTypePrimitive FieldType = iota
	FieldTypeEnum
	FieldTypeOneOf
	FieldTypeMessage
)

type Field interface {
	Name() string
	FQRN() string
	Type() FieldType
	IsRepeated() bool
}

type PrimitiveField interface{}

// EnumField appears as a field, but it has some EnumValueField.
// actual field is one of the its values.
type EnumField interface {
	Field
	Values() []EnumValueField
}

type EnumValueField interface {
	Field
	Number() int32
}

type OneOfField interface{}

type MessageField interface {
	Message
}

// TODO: map field
