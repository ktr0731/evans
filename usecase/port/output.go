package port

import (
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
)

type OutputPort interface {
	Package() (*PackageResponse, error)
	Service() (*ServiceResponse, error)
	Describe(msg *entity.Message) (io.Reader, error)
	Show() (*ShowResponse, error)
	Header() (*HeaderResponse, error)
	Call(res proto.Message) (io.Reader, error)
}
