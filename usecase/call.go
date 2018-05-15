package usecase

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

// EOS is return when inputting for streaming request is finished.
var EOS = errors.New("end of stream")

func Call(
	params *port.CallParams,
	outputPort port.OutputPort,
	inputter port.Inputter,
	grpcClient entity.GRPCClient,
	builder port.DynamicBuilder,
	env entity.Environment,
) (io.Reader, error) {
	rpc, err := env.RPC(params.RPCName)
	if err != nil {
		return nil, err
	}

	data := map[string]string{}
	for _, pair := range env.Headers() {
		if pair.Key != "user-agent" {
			data[pair.Key] = pair.Val
		}
	}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(data))

	var res proto.Message
	switch {
	case rpc.IsClientStreaming() && rpc.IsServerStreaming():
		panic("not implemented yet: bidirection")
	case rpc.IsClientStreaming():
		res, err = callClientStreaming(ctx, inputter, grpcClient, builder, rpc)
	case rpc.IsServerStreaming():
		res, err = callServerStreaming(ctx, inputter, grpcClient, builder, rpc)
	default:
		res, err = callUnary(ctx, inputter, grpcClient, builder, rpc)
	}
	if err := errors.Cause(err); err == context.Canceled {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return outputPort.Call(res)
}

func callUnary(
	ctx context.Context,
	inputter port.Inputter,
	grpcClient entity.GRPCClient,
	builder port.DynamicBuilder,
	rpc entity.RPC,
) (proto.Message, error) {
	req, err := inputter.Input(rpc.RequestMessage())
	if err != nil {
		return nil, err
	}

	res := builder.NewMessage(rpc.ResponseMessage())
	if err := grpcClient.Invoke(ctx, rpc.FQRN(), req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func callClientStreaming(
	ctx context.Context,
	inputter port.Inputter,
	grpcClient entity.GRPCClient,
	builder port.DynamicBuilder,
	rpc entity.RPC,
) (proto.Message, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	st, err := grpcClient.NewClientStream(ctx, rpc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client stream")
	}
	for {
		req, err := inputter.Input(rpc.RequestMessage())
		if err := errors.Cause(err); err == EOS {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed to input request message")
		}

		if err := st.Send(req); err != nil {
			return nil, errors.Wrap(err, "failed to send message")
		}
	}

	res := builder.NewMessage(rpc.ResponseMessage())
	if err := st.CloseAndReceive(res); err != nil {
		return nil, errors.Wrap(err, "stream closed with abnormal status")
	}
	return res, nil
}

type serverStreamingResult struct {
	res []proto.Message
}

func (r *serverStreamingResult) Append(m proto.Message) {
	r.res = append(r.res, m)
}

func (r *serverStreamingResult) Reset() {
	for i := range r.res {
		r.res[i].Reset()
	}
}

func (r *serverStreamingResult) String() string {
	var b *strings.Builder
	for i := range r.res {
		io.WriteString(b, r.res[i].String())
		b.WriteRune('\n')
	}
	return b.String()
}

func (r *serverStreamingResult) ProtoMessage() {
	for i := range r.res {
		r.res[i].ProtoMessage()
	}
}

func callServerStreaming(
	ctx context.Context,
	inputter port.Inputter,
	grpcClient entity.GRPCClient,
	builder port.DynamicBuilder,
	rpc entity.RPC,
) (proto.Message, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	st, err := grpcClient.NewServerStream(ctx, rpc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client stream")
	}
	req, err := inputter.Input(rpc.RequestMessage())
	if err != nil {
		return nil, errors.Wrap(err, "failed to input request message")
	}

	if err := st.Send(req); err != nil {
		return nil, errors.Wrap(err, "failed to send server streaming request")
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	resCh := make(chan proto.Message)
	go func() {
		defer cancel()
		for {
			res := builder.NewMessage(rpc.ResponseMessage())
			if err := st.Receive(res); err != nil {
				return
			}
			resCh <- res
		}
	}()

	// TODO: 動的に出力する
	res := &serverStreamingResult{
		res: []proto.Message{},
	}
	for {
		select {
		case <-ctx.Done():
			return res, nil
		case r := <-resCh:
			res.Append(r)
		default:
		}
	}
}
