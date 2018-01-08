package usecase

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/adapter/presenter"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/stretchr/testify/require"
)

type callEnv struct {
	entity.Environment

	rpc *entity.RPC
}

func (e *callEnv) RPC(rpcName string) (*entity.RPC, error) {
	return e.rpc, nil
}

type callInputter struct{}

func (i *callInputter) Input(reqType *desc.MessageDescriptor) (proto.Message, error) {
	return nil, nil
}

type callGRPCClient struct{}

func (c *callGRPCClient) Invoke(ctx context.Context, fqrn string, req, res interface{}) error {
	resText := "this is a response"
	res = &resText
	return nil
}

func dummyRPC(t *testing.T) *entity.RPC {
	set := helper.ReadProto(t, []string{filepath.Join("helloworld/helloworld.proto")})
	env := helper.NewEnv(t, set, helper.TestConfig().Env)
	rpc, err := env.RPC("SayHello")
	require.NoError(t, err)
	return rpc
}

func TestCall(t *testing.T) {
	params := &port.CallParams{"SayHello"}
	presenter := presenter.NewJSONCLIPresenter()

	env := &callEnv{rpc: dummyRPC(t)}
	inputter := &callInputter{}
	grpcClient := &callGRPCClient{}

	Call(params, presenter, inputter, grpcClient, env)
}
