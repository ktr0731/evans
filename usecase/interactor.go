package usecase

import (
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

type Interactor struct {
	env *entity.Env

	outputPort   port.OutputPort
	inputterPort port.Inputter
}

type InteractorParams struct {
	OutputPort   port.OutputPort
	InputterPort port.Inputter
}

func NewInteractor(params *InteractorParams) *Interactor {
	return &Interactor{
		outputPort:   params.OutputPort,
		inputterPort: params.InputterPort,
	}
}

func (i *Interactor) Package(params *port.PackageParams) (*port.PackageResponse, error) {
	return i.outputPort.Package()
}

func (i *Interactor) Service(params *port.ServiceParams) (*port.ServiceResponse, error) {
	return i.outputPort.Service()
}

func (i *Interactor) Describe(params *port.DescribeParams) (*port.DescribeResponse, error) {
	return i.outputPort.Describe()
}

func (i *Interactor) Show(params *port.ShowParams) (*port.ShowResponse, error) {
	return i.outputPort.Show()
}

func (i *Interactor) Header(params *port.HeaderParams) (*port.HeaderResponse, error) {
	return i.outputPort.Header()
}

func (i *Interactor) Call(params *port.CallParams) (*port.CallResponse, error) {
	return i.outputPort.Call()
}
