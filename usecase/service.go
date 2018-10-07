package usecase

import (
	"io"

	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
)

func Service(params *port.ServiceParams, outputPort port.OutputPort, env env.Environment) (io.Reader, error) {
	if err := env.UseService(params.SvcName); err != nil {
		return nil, err
	}
	return outputPort.Service()
}
