package usecase

import (
	"errors"
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Show(params *port.ShowParams, outputPort port.OutputPort, env entity.Environment) (io.Reader, error) {
	var showable port.Showable
	switch params.Type {
	case port.ShowTypePackage:
		showable = port.Packages(env.Packages())
	case port.ShowTypeService:
		svcs, err := env.Services()
		if err != nil {
			return nil, err
		}
		showable = port.Services(svcs)
	case port.ShowTypeMessage:
		msgs, err := env.Messages()
		if err != nil {
			return nil, err
		}
		showable = port.Messages(msgs)
	case port.ShowTypeRPC:
		rpcs, err := env.RPCs()
		if err != nil {
			return nil, err
		}
		showable = port.RPCs(rpcs)
	default:
		return nil, errors.New("unknown showable type")
	}

	return outputPort.Show(showable)
}
