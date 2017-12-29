package usecase

import "github.com/ktr0731/evans/usecase/port"

type Interactor struct {
	// presenter port.OutputPort
}

func NewInteractor() *Interactor {
	return &Interactor{}
}

func (i *Interactor) Package(params *port.PackageParams) (*port.PackageResponse, error) {
	// return i.presenter.Package(foo), nil
	return nil, nil
}

func (i *Interactor) Service(params *port.ServiceParams) (*port.ServiceResponse, error) {
	return nil, nil
}

func (i *Interactor) Describe(params *port.DescribeParams) (*port.DescribeResponse, error) {
	return nil, nil
}

func (i *Interactor) Show(params *port.ShowParams) (*port.ShowResponse, error) {
	return nil, nil
}

func (i *Interactor) Header(params *port.HeaderParams) (*port.HeaderResponse, error) {
	return nil, nil
}

func (i *Interactor) Call(params *port.CallParams) (*port.CallResponse, error) {
	return nil, nil
}
