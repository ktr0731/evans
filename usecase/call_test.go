package usecase

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/testentity"
	"github.com/ktr0731/evans/usecase/port"
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

type callInputter struct{}

func (i *callInputter) Input(_ entity.Message) (proto.Message, error) {
	return nil, nil
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
