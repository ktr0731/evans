package port

import (
	"io"

	"github.com/golang/protobuf/proto"
)

type OutputPort interface {
	Package() (io.Reader, error)
	Service() (io.Reader, error)
	Describe(showable Showable) (io.Reader, error)
	Show(showable Showable) (io.Reader, error)
	Header() (io.Reader, error)
	Call(res proto.Message) (io.Reader, error)
}
