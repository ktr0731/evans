package protobuf

import (
	"errors"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
)

type MessageSetter struct {
	m *dynamic.Message
}

func NewMessageSetter(m entity.Message) *MessageSetter {
	return &MessageSetter{
		m: dynamic.NewMessage(m.(*message).d),
	}
}

func (s *MessageSetter) SetField(field entity.Field, v interface{}) error {
	switch f := field.(type) {
	case *enumField:
		return s.m.TrySetField(f.d, v)
	default:
		return errors.New("type assertion failed")
	}
}
