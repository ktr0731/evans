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
	for _, h := range params.Headers {
		if h.NeedToRemove {
			env.RemoveHeader(h.Key)
		} else {
			if err := env.AddHeader(h); err != nil {
				return nil, err
			}
		}
	}
	return outputPort.Header()
}
