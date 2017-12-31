package usecase

import (
	"github.com/k0kubun/pp"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Show(params *port.ShowParams, outputPort port.OutputPort, env entity.Environment) (*port.ShowResponse, error) {
	var showable port.Showable
	switch params.Type {
	case port.ShowTypePackage:
		showable = port.Packages(env.Packages())
	case port.ShowTypeService:
		showable = port.Services(env.Services())
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
	}

	pp.Println(showable)

	return outputPort.Show()
}
