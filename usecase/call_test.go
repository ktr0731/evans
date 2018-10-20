package usecase

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/testentity"
	mockentity "github.com/ktr0731/evans/tests/mock/entity"
	"github.com/ktr0731/evans/tests/mock/entity/mockenv"
	"github.com/ktr0731/evans/tests/mock/usecase/mockport"
	"github.com/ktr0731/evans/usecase/internal/usecasetest"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newGRPCClient(t *testing.T) *mockentity.GRPCClientMock {
	return &mockentity.GRPCClientMock{
		InvokeFunc: func(ctx context.Context, fqrn string, req, res interface{}) error {
			resText := "this is a response"
			res = &resText
			return nil
		},
		CloseFunc: func(context.Context) error { return nil },
		NewClientStreamFunc: func(ctx context.Context, rpc entity.RPC) (entity.ClientStream, error) {
			return &mockentity.ClientStreamMock{
				SendFunc:            func(req proto.Message) error { return nil },
				CloseAndReceiveFunc: func(res *proto.Message) error { return nil },
			}, nil
		},
		NewServerStreamFunc: func(ctx context.Context, rpc entity.RPC) (entity.ServerStream, error) {
			return &mockentity.ServerStreamMock{
				SendFunc:    func(req proto.Message) error { return nil },
				ReceiveFunc: func(res *proto.Message) error { return nil },
			}, nil
		},
		NewBidiStreamFunc: func(ctx context.Context, rpc entity.RPC) (entity.BidiStream, error) {
			return &mockentity.BidiStreamMock{
				SendFunc:    func(req proto.Message) error { return nil },
				ReceiveFunc: func(res *proto.Message) error { return nil },
				CloseFunc:   func() error { return nil },
			}, nil
		},
	}
}

func newDynamicBuilder(t *testing.T) *mockport.DynamicBuilderMock {
	return &mockport.DynamicBuilderMock{
		NewMessageFunc: func(entity.Message) proto.Message {
			return nil
		},
	}
}

func TestCall(t *testing.T) {
	params := &port.CallParams{RPCName: "SayHello"}
	presenter := usecasetest.NewPresenter()

	newEnv := func(t *testing.T) *mockenv.EnvironmentMock {
		return &mockenv.EnvironmentMock{
			RPCFunc:     func(name string) (entity.RPC, error) { return testentity.NewRPC(), nil },
			HeadersFunc: func() []*entity.Header { return []*entity.Header{} },
		}
	}
	inputter := &mockport.InputterMock{
		InputFunc: func(entity.Message) (proto.Message, error) { return nil, nil },
	}
	builder := newDynamicBuilder(t)

	t.Run("normal", func(t *testing.T) {
		grpcClient := newGRPCClient(t)
		env := newEnv(t)
		res, err := Call(params, presenter, inputter, grpcClient, builder, env)
		assert.NoError(t, err)

		assert.Equal(t, nil, res)
	})

	t.Run("with headers", func(t *testing.T) {
		grpcClient := newGRPCClient(t)
		env := newEnv(t)
		env.HeadersFunc = func() []*entity.Header {
			return []*entity.Header{
				{"foo", "bar", false},
				{"hoge", "fuga", false},
				{"user-agent", "evans", false},
			}
		}

		res, err := Call(params, presenter, inputter, grpcClient, builder, env)
		assert.NoError(t, err)

		assert.Equal(t, nil, res)

		require.Len(t, grpcClient.InvokeCalls(), 1)
		actualCtx := grpcClient.InvokeCalls()[0].Ctx

		md, ok := metadata.FromOutgoingContext(actualCtx)
		assert.True(t, ok)
		assert.Len(t, md, 2)

		// user cannot set "user-agent" header.
		// ref. #47
		_, ok = md["user-agent"]
		assert.False(t, ok)
	})
}

func TestCall_ClientStream(t *testing.T) {
	params := &port.CallParams{RPCName: "SayHello"}
	presenter := usecasetest.NewPresenter()

	rpc := testentity.NewRPC()
	rpc.FIsClientStreaming = true

	newEnv := func(t *testing.T) *mockenv.EnvironmentMock {
		return &mockenv.EnvironmentMock{
			RPCFunc:     func(name string) (entity.RPC, error) { return rpc, nil },
			HeadersFunc: func() []*entity.Header { return []*entity.Header{} },
		}
	}
	env := newEnv(t)

	inputter := &mockport.InputterMock{
		InputFunc: func(entity.Message) (proto.Message, error) { return nil, io.EOF },
	}
	grpcClient := newGRPCClient(t)
	builder := newDynamicBuilder(t)

	t.Run("normal", func(t *testing.T) {
		res, err := Call(params, presenter, inputter, grpcClient, builder, env)
		assert.NoError(t, err)
		assert.Equal(t, nil, res)
	})
}

func TestCall_ServerStream(t *testing.T) {
	presenter := usecasetest.NewPresenter()
	rpc := testentity.NewRPC()
	rpc.FIsServerStreaming = true
	builder := newDynamicBuilder(t)
	grpcClient := newGRPCClient(t)

	t.Run("normal", func(t *testing.T) {
		inputter := &mockport.InputterMock{
			InputFunc: func(entity.Message) (proto.Message, error) { return nil, nil },
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		_, err := callServerStreaming(ctx, presenter, inputter, grpcClient, builder, rpc)
		assert.NoError(t, err)
	})

	t.Run("inputting canceled", func(t *testing.T) {
		inputter := &mockport.InputterMock{
			InputFunc: func(entity.Message) (proto.Message, error) { return nil, io.EOF },
		}
		_, err := callServerStreaming(context.Background(), presenter, inputter, grpcClient, builder, rpc)
		assert.Equal(t, io.EOF, errors.Cause(err))
	})
}

func TestCall_BidiStream(t *testing.T) {
	presenter := usecasetest.NewPresenter()
	rpc := testentity.NewRPC()
	rpc.FIsServerStreaming = true
	rpc.FIsClientStreaming = true
	builder := newDynamicBuilder(t)
	grpcClient := newGRPCClient(t)

	t.Run("client end", func(t *testing.T) {
		inputter := &mockport.InputterMock{
			InputFunc: func(entity.Message) (proto.Message, error) { return nil, nil },
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		_, err := callBidiStreaming(ctx, presenter, inputter, grpcClient, builder, rpc)
		assert.NoError(t, err)
	})
}
