package usecase

import (
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Show(params *port.ShowParams, outputPort port.OutputPort, env entity.Environment) (*port.ShowResponse, error) {
	return outputPort.Show()
}
