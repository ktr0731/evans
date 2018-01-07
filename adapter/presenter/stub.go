package presenter

import "github.com/ktr0731/evans/usecase/port"

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

func (p *StubPresenter) Call() (*port.CallResponse, error) {
	return nil, nil
}

func NewStubPresenter() *StubPresenter {
	return &StubPresenter{}
}
