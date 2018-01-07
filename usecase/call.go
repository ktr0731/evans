package usecase

import (
	"context"
	"io"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"google.golang.org/grpc/metadata"
)

func Call(
	params *port.CallParams,
	outputPort port.OutputPort,
	inputter port.Inputter,
	grpcPort port.GRPCPort,
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

	res := dynamic.NewMessage(rpc.ResponseType)

	md := metadata.MD{}
	for _, h := range env.Headers() {
		md[h.Key] = []string{h.Val}
	}
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	if err := grpcPort.Invoke(ctx, rpc.FQRN, req, res); err != nil {
		return nil, err
	}

	return outputPort.Call(res)
}
