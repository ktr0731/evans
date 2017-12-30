package presenter

import "github.com/ktr0731/evans/usecase/port"

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

func (p *CLIPresenter) Call() (*port.CallResponse, error) {
	return nil, nil
}

func NewCLIPresenter() *CLIPresenter {
	return &CLIPresenter{}
}
