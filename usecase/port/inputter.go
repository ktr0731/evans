package port

import (
	"github.com/ktr0731/evans/entity"
)

type Inputter interface {
	Input(reqMsg entity.Message) (interface{}, error)
}
