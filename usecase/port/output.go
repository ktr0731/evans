package port

import (
	"io"

	"github.com/golang/protobuf/proto"
)

type OutputPort interface {
	Package() (*PackageResponse, error)
	Service() (*ServiceResponse, error)
	Describe() (*DescribeResponse, error)
	Show() (*ShowResponse, error)
	Header() (*HeaderResponse, error)
	Call(res proto.Message) (io.Reader, error)
}
