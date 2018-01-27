package gateway

import (
	"testing"

	prompt "github.com/c-bata/go-prompt"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPrompt struct{}

func (p *mockPrompt) Input() string {
	return "foo"
}

func (p *mockPrompt) SetPrefix(_ string) error {
	return nil
}

func (p *mockPrompt) SetPrefixColor(_ prompt.Color) error {
	return nil
}

func TestPromptInputter_Input(t *testing.T) {
	t.Run("normal/simple", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "helloworld.proto", "helloworld", "Greeter")

		inputter := &PromptInputter{newPromptInputter(&mockPrompt{}, helper.TestConfig(), env)}

		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestType)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		assert.Equal(t, `name:"foo" message:"foo"`, msg.String())
	})

	t.Run("normal/nested_message", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "nested.proto", "library", "Library")

		inputter := &PromptInputter{newPromptInputter(&mockPrompt{}, helper.TestConfig(), env)}

		rpc, err := env.RPC("BorrowBook")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestType)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		assert.Equal(t, `person:<name:"foo"> book:<title:"foo" author:"foo">`, msg.String())
	})

	t.Run("normal/enum", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		inputter := &PromptInputter{newPromptInputter(&mockPrompt{}, helper.TestConfig(), env)}

		m, err := env.Message("Book")
		require.NoError(t, err)

		dmsg, err := inputter.Input(m.Desc)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		assert.Equal(t, `person:<name:"foo"> book:<title:"foo" author:"foo">`, msg.String())
	})
}

func Test_resolveMessageDependency(t *testing.T) {
	env := testhelper.SetupEnv(t, "nested.proto", "library", "Library")

	msg, err := env.Message("Book")
	require.NoError(t, err)

	dep := messageDependency{}
	resolveMessageDependency(msg.Desc, dep, map[string]bool{})

	assert.Len(t, dep, 1)
}

func Test_makePrefix(t *testing.T) {
	env := testhelper.SetupEnv(t, "nested.proto", "library", "Library")

	prefix := "{ancestor}{name} ({type})"

	t.Run("primitive", func(t *testing.T) {
		msg, err := env.Message("Person")
		require.NoError(t, err)

		name := msg.Desc.GetFields()[0]

		expected := "name (TYPE_STRING)"
		actual := makePrefix(prefix, name, true)

		assert.Equal(t, expected, actual)
	})

	t.Run("nested", func(t *testing.T) {
		msg, err := env.Message("BorrowBookRequest")
		require.NoError(t, err)

		expected := []string{
			"Person::name (TYPE_STRING)",
			"Book::title (TYPE_STRING)",
			"Book::author (TYPE_STRING)",
		}

		personMsg := msg.Desc.GetFields()
		for i, m := range personMsg {
			for j, f := range m.GetMessageType().GetFields() {
				actual := makePrefix(prefix, f, false)
				assert.Equal(t, expected[i+j], actual)
			}
		}
	})
}
