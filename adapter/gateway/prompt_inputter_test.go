package gateway

import (
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPrompt struct{}

func (p *mockPrompt) Input() string {
	return "foo"
}

func TestPromptInputter_Input(t *testing.T) {
	set := helper.ReadProto(t, []string{filepath.Join("helloworld/helloworld.proto")})
	env := helper.NewEnv(t, set, helper.TestConfig().Env)

	inputter := &PromptInputter{newPromptInputter(&mockPrompt{}, env)}

	err := env.UsePackage("helloworld")
	require.NoError(t, err)

	err = env.UseService("Greeter")
	require.NoError(t, err)

	rpc, err := env.RPC("SayHello")
	require.NoError(t, err)

	dmsg, err := inputter.Input(rpc.RequestType)
	require.NoError(t, err)

	msg, ok := dmsg.(*dynamic.Message)
	require.True(t, ok)

	assert.Equal(t, `name:"foo" message:"foo"`, msg.String())

}
