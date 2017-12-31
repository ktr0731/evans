package usecase

import (
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Call(
	params *port.CallParams,
	outputPort port.OutputPort,
	env entity.Environment,
	inputter port.Inputter,
) (*port.CallResponse, error) {
	return nil, nil
}
