package gateway

import (
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPrompt struct{}

func (p *mockPrompt) Input() string {
	return "foo"
}

func TestPromptInputter_Input(t *testing.T) {
	t.Run("normal/simple", func(t *testing.T) {
		env := testhelper.SetupEnv(t, filepath.Join("helloworld", "helloworld.proto"), "helloworld", "Greeter")

		inputter := &PromptInputter{newPromptInputter(&mockPrompt{}, env)}

		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestType)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		assert.Equal(t, `name:"foo" message:"foo"`, msg.String())
	})

	t.Run("normal/nested_message", func(t *testing.T) {
		env := testhelper.SetupEnv(t, filepath.Join("nested_message", "library.proto"), "library", "Library")

		inputter := &PromptInputter{newPromptInputter(&mockPrompt{}, env)}

		rpc, err := env.RPC("BorrowBook")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestType)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		assert.Equal(t, `person:<name:"foo"> book:<title:"foo" author:"foo">`, msg.String())
	})
}

func Test_resolveMessageDependency(t *testing.T) {
	env := testhelper.SetupEnv(t, filepath.Join("nested_message", "library.proto"), "library", "Library")

	msg, err := env.Message("Book")
	require.NoError(t, err)

	dep := messageDependency{}
	resolveMessageDependency(msg.Desc, dep, map[string]bool{})

	assert.Len(t, dep, 1)
}
