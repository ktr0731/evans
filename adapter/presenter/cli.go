package presenter

import (
	"bytes"
	"io"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/usecase/port"
)

type CLIPresenter struct{}

func (p *CLIPresenter) Package() (*port.PackageResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Service() (*port.ServiceResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Describe() (*port.DescribeResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Show() (*port.ShowResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Header() (*port.HeaderResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Call(res *dynamic.Message) (io.Reader, error) {
	b, err := res.MarshalJSONIndent()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func NewCLIPresenter() *CLIPresenter {
	return &CLIPresenter{}
}
