package usecase

import (
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Call(
	params *port.CallParams,
	outputPort port.OutputPort,
	env entity.Environment,
	inputter port.Inputter,
) (*port.CallResponse, error) {
	rpc, err := env.RPC(params.RPCName)
	if err != nil {
		return nil, err
	}
	req, err := inputter.Input(rpc.RequestType)
	if err != nil {
		return nil, err
	}

	// 知るべきではない？
	res := dynamic.NewMessage(rpc.ResponseType)

	return nil, nil
}
