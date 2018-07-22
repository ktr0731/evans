package usecase

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/testentity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type callEnv struct {
	entity.Environment

	rpc     entity.RPC
	headers []*entity.Header
}

func (e *callEnv) RPC(rpcName string) (entity.RPC, error) {
	return e.rpc, nil
}

func (e *callEnv) Headers() []*entity.Header {
	return e.headers
}

type callInputter struct {
	mu  sync.Mutex
	err error
}

func (i *callInputter) Input(_ entity.Message) (proto.Message, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return nil, i.err
}

type callGRPCClient struct {
	actualCtx context.Context
}

func (c *callGRPCClient) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	resText := "this is a response"
	res = &resText
	c.actualCtx = ctx
	return nil
}

type callDynamicBuilder struct{}

func (b *callDynamicBuilder) NewMessage(_ entity.Message) proto.Message {
	return nil
}

func TestCall(t *testing.T) {
	params := &port.CallParams{RPCName: "SayHello"}
	presenter := &presenter.StubPresenter{}

	env := &callEnv{rpc: testentity.NewRPC()}
	inputter := &callInputter{}
	grpcClient := &callGRPCClient{}
	builder := &callDynamicBuilder{}

	t.Run("normal", func(t *testing.T) {
		res, err := Call(params, presenter, inputter, grpcClient, builder, env)
		assert.NoError(t, err)

		assert.Equal(t, nil, res)
	})

	t.Run("with headers", func(t *testing.T) {
		env.headers = []*entity.Header{
			{"foo", "bar", false},
			{"hoge", "fuga", false},
			{"user-agent", "evans", false},
		}
		res, err := Call(params, presenter, inputter, grpcClient, builder, env)
		assert.NoError(t, err)

		assert.Equal(t, nil, res)

		md, ok := metadata.FromOutgoingContext(grpcClient.actualCtx)
		assert.True(t, ok)
		assert.Len(t, md, 2)

		// user cannot set "user-agent" header.
		// ref. #47
		_, ok = md["user-agent"]
		assert.False(t, ok)
	})
}

type callClientStream struct{}

func (s *callClientStream) Send(req proto.Message) error { return nil }

func (s *callClientStream) CloseAndReceive(res proto.Message) error { return nil }

func (c *callGRPCClient) NewClientStream(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
	return &callClientStream{}, nil
}

func TestCall_ClientStream(t *testing.T) {
	params := &port.CallParams{RPCName: "SayHello"}
	presenter := &presenter.StubPresenter{}

	rpc := testentity.NewRPC()
	rpc.FIsClientStreaming = true
	env := &callEnv{rpc: rpc}
	inputter := &callInputter{err: io.EOF}
	grpcClient := &callGRPCClient{}
	builder := &callDynamicBuilder{}

	t.Run("normal", func(t *testing.T) {
		res, err := Call(params, presenter, inputter, grpcClient, builder, env)
		assert.NoError(t, err)
		assert.Equal(t, nil, res)
	})
}

type callServerStream struct{}

func (s *callServerStream) Send(_ proto.Message) error { return nil }

func (s *callServerStream) Receive(req proto.Message) error { return nil }

func (c *callGRPCClient) NewServerStream(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
	return &callServerStream{}, nil
}

func TestCall_ServerStream(t *testing.T) {
	presenter := &presenter.StubPresenter{}
	rpc := testentity.NewRPC()
	rpc.FIsServerStreaming = true
	builder := &callDynamicBuilder{}
	grpcClient := &callGRPCClient{}

	t.Run("normal", func(t *testing.T) {
		inputter := &callInputter{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		_, err := callServerStreaming(ctx, presenter, inputter, grpcClient, builder, rpc)
		assert.NoError(t, err)
	})

	t.Run("inputting canceled", func(t *testing.T) {
		inputter := &callInputter{err: io.EOF}
		_, err := callServerStreaming(context.Background(), presenter, inputter, grpcClient, builder, rpc)
		assert.Equal(t, io.EOF, errors.Cause(err))
	})
}

type callBidiStream struct{}

func (s *callBidiStream) Send(req proto.Message) error    { return nil }
func (s *callBidiStream) Receive(res proto.Message) error { return nil }
func (s *callBidiStream) Close() error                    { return nil }

func (c *callGRPCClient) NewBidiStream(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
	return &callBidiStream{}, nil
}

func TestCall_BidiStream(t *testing.T) {
	presenter := &presenter.StubPresenter{}
	rpc := testentity.NewRPC()
	rpc.FIsServerStreaming = true
	rpc.FIsClientStreaming = true
	builder := &callDynamicBuilder{}
	grpcClient := &callGRPCClient{}

	t.Run("client end", func(t *testing.T) {
		inputter := &callInputter{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		_, err := callBidiStreaming(ctx, presenter, inputter, grpcClient, builder, rpc)
		assert.NoError(t, err)
	})
}
