package port

import "github.com/gogo/protobuf/proto"

type Inputter interface {
	Input(req proto.Message) error
}
