package usecase

import (
	"context"
	"io"
	"os"
	"os/signal"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

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
		return callBidiStreaming(ctx, outputPort, inputter, grpcClient, builder, rpc)
	case rpc.IsClientStreaming():
		res, err = callClientStreaming(ctx, inputter, grpcClient, builder, rpc)
	case rpc.IsServerStreaming():
		return callServerStreaming(ctx, outputPort, inputter, grpcClient, builder, rpc)
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
		if err := errors.Cause(err); err == io.EOF {
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

type serverStreamingResultWriter struct {
	ctx    context.Context
	cancel func()
	s      entity.ServerStream

	output func(proto.Message) (io.Reader, error)

	newMessage func() proto.Message

	r *io.PipeReader
	w *io.PipeWriter

	closeCh chan struct{}
}

func newServerStramingResultWriter(
	ctx context.Context,
	s entity.ServerStream,
	outputPort func(proto.Message) (io.Reader, error),
	newMessage func() proto.Message,
) *serverStreamingResultWriter {
	r, w := io.Pipe()
	writer := &serverStreamingResultWriter{
		s:          s,
		output:     outputPort,
		newMessage: newMessage,
		r:          r,
		w:          w,
		closeCh:    make(chan struct{}),
	}
	writer.ctx, writer.cancel = context.WithCancel(ctx)
	go writer.receiveResponse()
	return writer
}

func (w *serverStreamingResultWriter) receiveResponse() {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	defer close(sigCh)

	go func() {
		defer w.cancel()
		for {
			select {
			case <-sigCh:
				w.w.CloseWithError(io.EOF)
			case <-w.closeCh:
			}
			return
		}
	}()

	resCh := make(chan proto.Message)
	go func() {
		defer w.cancel()
		for {
			select {
			case <-sigCh:
				return
			case <-w.closeCh:
				return
			default:
				res := w.newMessage()
				err := w.s.Receive(res)
				if err != nil {
					w.w.CloseWithError(err)
					close(resCh)
					return
				}
				resCh <- res
			}
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.closeCh:
			return
		case r, ok := <-resCh:
			if !ok {
				w.w.CloseWithError(io.EOF)
				return
			}

			res, err := w.output(r)
			if err != nil {
				w.w.CloseWithError(errors.Wrap(err, "failed to output server streaming response"))
				return
			}
			_, err = io.Copy(w.w, res)
			if err != nil {
				w.w.CloseWithError(errors.Wrap(err, "failed to write server streaming response"))
				return
			}
			_, err = io.WriteString(w.w, "\n")
			if err != nil {
				w.w.CloseWithError(err)
				return
			}
		default:
		}
	}

}

func (w *serverStreamingResultWriter) Read(b []byte) (int, error) {
	n, err := w.r.Read(b)
	if err != nil {
		w.cancel()
	}
	return n, err
}

func callServerStreaming(
	ctx context.Context,
	outputPort port.OutputPort,
	inputter port.Inputter,
	grpcClient entity.GRPCClient,
	builder port.DynamicBuilder,
	rpc entity.RPC,
) (io.Reader, error) {
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

	return newServerStramingResultWriter(
		ctx,
		st,
		outputPort.Call,
		func() proto.Message {
			return builder.NewMessage(rpc.ResponseMessage())
		}), nil
}

func callBidiStreaming(
	ctx context.Context,
	outputPort port.OutputPort,
	inputter port.Inputter,
	grpcClient entity.GRPCClient,
	builder port.DynamicBuilder,
	rpc entity.RPC,
) (io.Reader, error) {
	st, err := grpcClient.NewBidiStream(ctx, rpc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client stream")
	}

	w := newServerStramingResultWriter(
		ctx,
		st,
		outputPort.Call,
		func() proto.Message {
			return builder.NewMessage(rpc.ResponseMessage())
		})

	// TODO: resolve goroutine leak
	go func() {
		for {
			req, err := inputter.Input(rpc.RequestMessage())
			if err == io.EOF {
				st.Close()
				return
			}
			if err != nil {
				w.w.CloseWithError(errors.Wrap(err, "failed to input request message"))
				return
			}

			if err := st.Send(req); err != nil {
				w.w.CloseWithError(errors.Wrap(err, "failed to send server streaming request"))
				return
			}
		}
	}()

	return w, nil
}
