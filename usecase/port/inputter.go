package port

import (
	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
)

type Inputter interface {
	Input(reqMsg entity.Message) (proto.Message, error)
}
