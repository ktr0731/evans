package port

import (
	"io"

	"github.com/jhump/protoreflect/dynamic"
)

type OutputPort interface {
	Package() (*PackageResponse, error)
	Service() (*ServiceResponse, error)
	Describe() (*DescribeResponse, error)
	Show() (*ShowResponse, error)
	Header() (*HeaderResponse, error)
	Call(res *dynamic.Message) (io.Reader, error)
}
