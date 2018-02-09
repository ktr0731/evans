package gateway

import (
	"fmt"
	"testing"

	prompt "github.com/c-bata/go-prompt"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity/testentity"
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

type mockRepeatedPrompt struct {
	*mockPrompt

	cnt          int
	inputOutputs []string
}

func (p *mockRepeatedPrompt) Input() string {
	p.cnt++
	return p.inputOutputs[p.cnt-1]
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

	t.Run("normal/repeated", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "repeated.proto", "helloworld", "")

		prompt := &mockRepeatedPrompt{mockPrompt: &mockPrompt{}, inputOutputs: []string{"foo", "", "bar", "", ""}}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "repeated.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[0]
		require.Equal(t, "HelloRequest", m.Name())

		msg, err := inputter.Input(m)
		require.NoError(t, err)

		require.Equal(t, `name:"foo" name:"bar"`, msg.String())
	})

	t.Run("normal/map", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "map.proto", "example", "")

		prompt := &mockRepeatedPrompt{mockPrompt: &mockPrompt{}, inputOutputs: []string{"foo", "", "bar", "", ""}}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "map.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[0]
		require.Equal(t, "PrimitiveRequest", m.Name())

		msg, err := inputter.Input(m)
		require.NoError(t, err)

		require.Equal(t, `foo:<key:"foo" value:"bar">`, msg.String())
	})

	t.Run("normal/map val is message", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "map.proto", "example", "")

		prompt := &mockRepeatedPrompt{mockPrompt: &mockPrompt{}, inputOutputs: []string{"key", "", "val1", "3", "", ""}}
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "map.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[2]
		require.Equal(t, "MessageRequest", m.Name())

		msg, err := inputter.Input(m)
		require.NoError(t, err)

		require.Equal(t, `foo:<key:"key" value:<fuga:"val1" piyo:3>>`, msg.String())
	})
}

func Test_makePrefix(t *testing.T) {
	prefix := "{ancestor}{name} ({type})"
	f := testentity.NewFld()
	t.Run("primitive", func(t *testing.T) {
		expected := fmt.Sprintf("%s (%s)", f.FieldName(), f.PBType())
		actual := makePrefix(prefix, f, nil, false)

		require.Equal(t, expected, actual)
	})

	t.Run("nested", func(t *testing.T) {
		expected := fmt.Sprintf("Foo::Bar::%s (%s)", f.FieldName(), f.PBType())
		actual := makePrefix(prefix, f, []string{"Foo", "Bar"}, false)
		require.Equal(t, expected, actual)
	})

	t.Run("repeated (field)", func(t *testing.T) {
		expected := fmt.Sprintf("Foo::Bar::%s <repeated> (%s)", f.FieldName(), f.PBType())
		f.FIsRepeated = true
		actual := makePrefix(prefix, f, []string{"Foo", "Bar"}, false)
		require.Equal(t, expected, actual, false)
	})
}
