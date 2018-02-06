package protobuf

import (
	"errors"

	"github.com/golang/protobuf/proto"
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

func (s *MessageSetter) SetEnumField(f entity.EnumField, v int32) error {
	e, ok := f.(*enumField)
	if !ok {
		return errors.New("type assertion failed")
	}
	return s.m.TrySetField(e.d, v)
}

func (s *MessageSetter) Done() proto.Message {
	m := s.m
	s.m = nil
	return m
}
