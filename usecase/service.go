package usecase

import (
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Service(params *port.ServiceParams, outputPort port.OutputPort, env entity.Environment) (*port.ServiceResponse, error) {
	if err := env.UseService(params.SvcName); err != nil {
		return nil, err
	}
	return outputPort.Service()
}
