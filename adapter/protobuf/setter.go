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
	var dm *dynamic.Message
	switch d := m.(type) {
	case *message:
		dm = dynamic.NewMessage(d.d)
	case *messageField: // nested message
		dm = dynamic.NewMessage(d.d.GetMessageType())
	default:
		panic("unknown type")
	}
	return &MessageSetter{
		m: dm,
	}
}

func (s *MessageSetter) SetField(field entity.Field, v interface{}) error {
	switch f := field.(type) {
	case *enumField:
		return s.m.TrySetField(f.d, v)
	case *messageField:
		if f.IsRepeated() {
			return s.m.TryAddRepeatedField(f.d, v)
		}
		return s.m.TrySetField(f.d, v)
	case *primitiveField:
		if f.IsRepeated() {
			return s.m.TryAddRepeatedField(f.d, v)
		}
		return s.m.TrySetField(f.d, v)
	default:
		return errors.New("unknown type: " + f.PBType())
	}
}

func (s *MessageSetter) Done() proto.Message {
	m := s.m
	s.m = nil
	return m
}
