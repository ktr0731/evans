package entity

type FieldType int

const (
	FieldTypePrimitive FieldType = iota
	FieldTypeEnum
	FieldTypeOneOf
	FieldTypeMessage
	FieldTypeMap
)

// fieldable types:
//	enum, oneof, message, primitive
type Field interface {
	FieldName() string
	FQRN() string
	Type() FieldType
	IsRepeated() bool

	// *desc.FieldDescriptor.GetType().String()
	// used only from port.Showable
	PBType() string
}

type PrimitiveField interface {
	Field
}

// EnumField is set of values
type EnumField interface {
	Field
	Enum
}

type OneOfField interface {
	Field
	Choices() []Field
}

type MessageField interface {
	Field
	Message
}

type MapField interface {
	Field
	Key() Field
	Val() Field
}
