package gateway

import (
	"testing"

	prompt "github.com/c-bata/go-prompt"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockPrompt struct {
	inputOutput string

	selectOutputString string
	selectOutputError  error
}

func (p *mockPrompt) Input() string {
	return p.inputOutput
}

func (p *mockPrompt) setExpectedInput(s string) {
	p.inputOutput = s
}

func (p *mockPrompt) Select(msg string, opts []string) (string, error) {
	return p.selectOutputString, p.selectOutputError
}

func (p *mockPrompt) setExpectedSelect(s string, err error) {
	p.selectOutputString = s
	p.selectOutputError = err
}

func (p *mockPrompt) SetPrefix(_ string) error {
	return nil
}

func (p *mockPrompt) SetPrefixColor(_ prompt.Color) error {
	return nil
}

func TestPrompt_Input(t *testing.T) {
	t.Run("normal/simple", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "helloworld.proto", "helloworld", "Greeter")

		prompt := &mockPrompt{}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)

		prompt.setExpectedInput("foo")
		dmsg, err := inputter.Input(rpc.RequestMessage())
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `name:"foo" message:"foo"`, msg.String())
	})

	t.Run("normal/nested_message", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "nested.proto", "library", "Library")

		prompt := &mockPrompt{}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		rpc, err := env.RPC("BorrowBook")
		require.NoError(t, err)

		prompt.setExpectedInput("foo")
		dmsg, err := inputter.Input(rpc.RequestMessage())
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `person:<name:"foo"> book:<title:"foo" author:"foo">`, msg.String())
	})

	t.Run("normal/enum", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		prompt := &mockPrompt{}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "enum.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[0]

		prompt.setExpectedSelect("PHILOSOPHY", nil)

		dmsg, err := inputter.Input(m)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `type:PHILOSOPHY`, msg.String())
	})

	t.Run("error/enum:invalid enum name", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		prompt := &mockPrompt{}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "enum.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[0]

		prompt.setExpectedSelect("kumiko", nil)

		_, err = inputter.Input(m)
		e := errors.Cause(err)
		require.Equal(t, ErrUnknownEnumName, e)
	})

	t.Run("normal/oneof", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "oneof.proto", "shop", "")

		prompt := &mockPrompt{}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "oneof.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)

		m := p[0].Messages[2]
		require.Equal(t, m.Name(), "BorrowRequest")
		require.Len(t, m.Fields(), 1)

		prompt.setExpectedInput("bar")
		prompt.setExpectedSelect("book", nil)

		dmsg, err := inputter.Input(m)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `book:<title:"bar" author:"bar">`, msg.String())
	})

	t.Run("error/oneof:invalid oneof field name", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "oneof.proto", "shop", "")

		prompt := &mockPrompt{}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "oneof.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[2]

		prompt.setExpectedInput("bar")
		prompt.setExpectedSelect("Book", nil)

		_, err = inputter.Input(m)

		e := errors.Cause(err)
		require.Equal(t, ErrUnknownOneofFieldName, e)
	})
}

// func Test_makePrefix(t *testing.T) {
// 	descs := testhelper.ReadProtoAsFileDescriptors(t, "nested.proto")
//
// 	prefix := "{ancestor}{name} ({type})"
//
// 	t.Run("primitive", func(t *testing.T) {
// 		m := testhelper.FindMessage(t, "Person", descs)
//
// 		name := m.GetFields()[0]
//
// 		expected := "name (TYPE_STRING)"
// 		actual := makePrefix(prefix, name, true)
//
// 		require.Equal(t, expected, actual)
// 	})
//
// 	t.Run("nested", func(t *testing.T) {
// 		m := testhelper.FindMessage(t, "BorrowBookRequest", descs)
//
// 		expected := []string{
// 			"Person::name (TYPE_STRING)",
// 			"Book::title (TYPE_STRING)",
// 			"Book::author (TYPE_STRING)",
// 		}
//
// 		personMsg := m.GetFields()
// 		for i, m := range personMsg {
// 			for j, f := range m.GetMessageType().GetFields() {
// 				actual := makePrefix(prefix, f, false)
// 				require.Equal(t, expected[i+j], actual)
// 			}
// 		}
// 	})
// }
