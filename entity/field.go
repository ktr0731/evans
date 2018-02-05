package entity

const (
	NON_FIELD = int32(0)
)

// fieldable types:
//	enum, oneof, message
type field struct {
	Name, Type string
}
