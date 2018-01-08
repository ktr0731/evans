package usecase

import (
	"errors"
	"io"

	"github.com/ktr0731/evans/usecase/port"
)

func Header(params *port.HeaderParams, outputPort port.OutputPort) (io.Reader, error) {
	return nil, errors.New("not implemented yet (https://github.com/ktr0731/evans/issues/7)")
}
