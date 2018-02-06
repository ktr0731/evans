package entity

type FieldType int

const (
	FieldTypePrimitive FieldType = iota
	FieldTypeEnum
	FieldTypeOneOf
	FieldTypeMessage
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

// EnumField is a value of EnumDescriptor.GetValues()
type EnumField interface {
	Field
	Number() int32
}

type OneOfField interface {
	Field
}

type MessageField interface {
	Field
	Message
}

// TODO: map field
