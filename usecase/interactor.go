package usecase

import (
	"context"
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

type Interactor struct {
	env *entity.Env

	outputPort     port.OutputPort
	inputterPort   port.Inputter
	grpcPort       entity.GRPCClient
	dynamicBuilder port.DynamicBuilder
}

type InteractorParams struct {
	Env *entity.Env

	OutputPort     port.OutputPort
	InputterPort   port.Inputter
	DynamicBuilder port.DynamicBuilder
	GRPCClient     entity.GRPCClient
}

func NewInteractor(params *InteractorParams) *Interactor {
	return &Interactor{
		env:            params.Env,
		outputPort:     params.OutputPort,
		inputterPort:   params.InputterPort,
		grpcPort:       params.GRPCClient,
		dynamicBuilder: params.DynamicBuilder,
	}
}

func (i *Interactor) Package(params *port.PackageParams) (io.Reader, error) {
	return Package(params, i.outputPort, i.env)
}

func (i *Interactor) Service(params *port.ServiceParams) (io.Reader, error) {
	return Service(params, i.outputPort, i.env)
}

func (i *Interactor) Describe(params *port.DescribeParams) (io.Reader, error) {
	return Describe(params, i.outputPort, i.env)
}

func (i *Interactor) Show(params *port.ShowParams) (io.Reader, error) {
	return Show(params, i.outputPort, i.env)
}

func (i *Interactor) Header(params *port.HeaderParams) (io.Reader, error) {
	return Header(params, i.outputPort, i.env)
}

func (i *Interactor) Call(params *port.CallParams) (io.Reader, error) {
	return Call(params, i.outputPort, i.inputterPort, i.grpcPort, i.dynamicBuilder, i.env)
}

// Close closes all dependencies by each Close method.
func (i *Interactor) Close(ctx context.Context) error {
	return i.grpcPort.Close(ctx)
}
