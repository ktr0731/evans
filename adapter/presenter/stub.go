package presenter

import (
	"bytes"
	"io"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/usecase/port"
)

type StubPresenter struct{}

func (p *StubPresenter) Package() (*port.PackageResponse, error) {
	return nil, nil
}

func (p *StubPresenter) Service() (*port.ServiceResponse, error) {
	return nil, nil
}

func (p *StubPresenter) Describe() (*port.DescribeResponse, error) {
	return nil, nil
}

func (p *StubPresenter) Show() (*port.ShowResponse, error) {
	return nil, nil
}

func (p *StubPresenter) Header() (*port.HeaderResponse, error) {
	return nil, nil
}

func (p *StubPresenter) Call(res *dynamic.Message) (io.Reader, error) {
	b, err := res.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func NewStubPresenter() *StubPresenter {
	return &StubPresenter{}
}
