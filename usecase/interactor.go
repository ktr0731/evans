package usecase

import (
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

type Interactor struct {
	env *entity.Env

	outputPort     port.OutputPort
	logger         port.Logger
	inputterPort   port.Inputter
	grpcPort       port.GRPCPort
	dynamicBuilder port.DynamicBuilder
}

type InteractorParams struct {
	OutputPort   port.OutputPort
	InputterPort port.Inputter
	Logger       port.Logger
}

func NewInteractor(params *InteractorParams) *Interactor {
	return &Interactor{
		outputPort:   params.OutputPort,
		inputterPort: params.InputterPort,
	}
}

func (i *Interactor) Package(params *port.PackageParams) (*port.PackageResponse, error) {
	return Package(params, i.outputPort, i.logger, i.env)
}

func (i *Interactor) Service(params *port.ServiceParams) (*port.ServiceResponse, error) {
	return Service(params, i.outputPort, i.env)
}

func (i *Interactor) Describe(params *port.DescribeParams) (*port.DescribeResponse, error) {
	return Describe(params, i.outputPort, i.env)
}

func (i *Interactor) Show(params *port.ShowParams) (*port.ShowResponse, error) {
	return Show(params, i.outputPort, i.env)
}

func (i *Interactor) Header(params *port.HeaderParams) (*port.HeaderResponse, error) {
	return i.outputPort.Header()
}

func (i *Interactor) Call(params *port.CallParams) (io.Reader, error) {
	return Call(params, i.outputPort, i.inputterPort, i.grpcPort, i.dynamicBuilder, i.env)
}
