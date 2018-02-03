package usecase

import (
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Header(
	params *port.HeaderParams,
	outputPort port.OutputPort,
	env entity.Environment,
) (io.Reader, error) {
	if err := env.AddHeaders(params.Headers...); err != nil {
		return nil, err
	}
	return outputPort.Header()
}
