package port

import (
	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
)

type DynamicBuilder interface {
	NewMessage(m entity.Message) proto.Message
}
