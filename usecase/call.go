package usecase

import (
	"context"
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"google.golang.org/grpc/metadata"
)

func Call(
	params *port.CallParams,
	outputPort port.OutputPort,
	inputter port.Inputter,
	grpcPort port.GRPCPort,
	builder port.DynamicBuilder,
	env entity.Environment,
) (io.Reader, error) {
	rpc, err := env.RPC(params.RPCName)
	if err != nil {
		return nil, err
	}
	req, err := inputter.Input(rpc.RequestType)
	if err != nil {
		return nil, err
	}

	res := builder.NewMessage(rpc.ResponseType)

	data := map[string]string{}
	for _, pair := range env.Headers() {
		data[pair.Key] = pair.Val
	}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(data))

	if err := grpcPort.Invoke(ctx, rpc.FQRN, req, res); err != nil {
		return nil, err
	}

	return outputPort.Call(res)
}
