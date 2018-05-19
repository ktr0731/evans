package usecase

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/k0kubun/pp"
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

type serverStreamingResultWriter struct {
	ctx    context.Context
	cancel func()
	s      entity.ServerStream

	output func(proto.Message) (io.Reader, error)

	newMessage func() proto.Message

	r *bufio.Reader
	w *bytes.Buffer

	isFinished bool
}

func newServerStramingResultWriter(
	ctx context.Context,
	s entity.ServerStream,
	outputPort func(proto.Message) (io.Reader, error),
	newMessage func() proto.Message,
) *serverStreamingResultWriter {
	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	w := &serverStreamingResultWriter{
		s:          s,
		output:     outputPort,
		newMessage: newMessage,
		r:          bufio.NewReader(buf),
		w:          buf,
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	go w.receive()
	return w
}

func (w *serverStreamingResultWriter) receive() {
	pp.Println("waiting...")
	resCh := make(chan proto.Message)
	go func() {
		defer w.cancel()
		for {
			res := w.newMessage()
			err := w.s.Receive(res)
			if err == io.EOF {
				w.isFinished = true
			}
			if err != nil {
				return
			}
			resCh <- res
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case r := <-resCh:
			pp.Println("received")
			res, err := w.output(r)
			if err != nil {
				// return 0, errors.Wrap(err, "failed to output server streaming response")
				panic(err)
			}
			_, err = io.Copy(w.w, res)
			if err != nil {
				// return 0, errors.Wrap(err, "failed to write server streaming response")
				panic(err)
			}
			_, err = io.WriteString(w.w, "\n")
			if err != nil {
				// return 0, err
				panic(err)
			}
			pp.Println("wrote")
		default:
		}
	}

}

func (w *serverStreamingResultWriter) Read(b []byte) (int, error) {
	// sigCh := make(chan os.Signal)
	// signal.Notify(sigCh, os.Interrupt)
	// go func() {
	// 	<-sigCh
	// 	w.cancel()
	// }()

	for w.w.Len() == 0 && !w.isFinished {
		time.Sleep(100 * time.Millisecond)
	}

	n, err := w.r.Read(b)
	if err == io.EOF {
		pp.Println("END")
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
