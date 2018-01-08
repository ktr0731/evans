package port

import (
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
)

type OutputPort interface {
	Package() (io.Reader, error)
	Service() (io.Reader, error)
	Describe(msg *entity.Message) (io.Reader, error)
	Show() (io.Reader, error)
	Header() (io.Reader, error)
	Call(res proto.Message) (io.Reader, error)
}
