package gateway

import (
	"fmt"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity/testentity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPrompt_Input(t *testing.T) {
	t.Run("normal/simple", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "helloworld.proto", "helloworld", "Greeter")

		prompt := helper.NewMockPrompt([]string{"foo"}, nil)
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestMessage())
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `name:"foo" message:"foo"`, msg.String())
	})

	t.Run("normal/nested_message", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "nested.proto", "library", "Library")

		prompt := helper.NewMockPrompt([]string{"foo"}, nil)
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		rpc, err := env.RPC("BorrowBook")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestMessage())
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `person:<name:"foo"> book:<title:"foo" author:"foo">`, msg.String())
	})

	t.Run("normal/enum", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		prompt := helper.NewMockPrompt(nil, []string{"PHILOSOPHY"})
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "enum.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[0]

		dmsg, err := inputter.Input(m)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `type:PHILOSOPHY`, msg.String())
	})

	t.Run("error/enum:invalid enum name", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		prompt := helper.NewMockPrompt(nil, []string{"kumiko"})
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "enum.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[0]

		_, err = inputter.Input(m)
		e := errors.Cause(err)
		require.Equal(t, ErrUnknownEnumName, e)
	})

	t.Run("normal/oneof", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "oneof.proto", "shop", "")

		prompt := helper.NewMockPrompt([]string{"bar"}, []string{"book"})
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "oneof.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)

		m := p[0].Messages[2]
		require.Equal(t, m.Name(), "BorrowRequest")
		require.Len(t, m.Fields(), 1)

		dmsg, err := inputter.Input(m)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `book:<title:"bar" author:"bar">`, msg.String())
	})

	t.Run("error/oneof:invalid oneof field name", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "oneof.proto", "shop", "")

		prompt := helper.NewMockPrompt([]string{"bar"}, []string{"Book"})
		inputter := newPrompt(prompt, helper.TestConfig(), env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "oneof.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[2]

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
	var f *testentity.Fld
	backup := func(tmp testentity.Fld) func() {
		return func() {
			f = &tmp
		}
	}

	prefix := "{ancestor}{name} ({type})"
	f = testentity.NewFld()
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
		expected := fmt.Sprintf("<repeated> Foo::Bar::%s (%s)", f.FieldName(), f.PBType())
		cleanup := backup(*f)
		defer cleanup()
		f.FIsRepeated = true
		actual := makePrefix(prefix, f, []string{"Foo", "Bar"}, false)
		require.Equal(t, expected, actual, false)
	})

	t.Run("repeated (ancestor)", func(t *testing.T) {
		expected := fmt.Sprintf("<repeated> Foo::Bar::%s (%s)", f.FieldName(), f.PBType())
		actual := makePrefix(prefix, f, []string{"Foo", "Bar"}, true)
		require.Equal(t, expected, actual, false)
	})

	t.Run("repeated (both)", func(t *testing.T) {
		expected := fmt.Sprintf("<repeated> Foo::Bar::%s (%s)", f.FieldName(), f.PBType())
		cleanup := backup(*f)
		defer cleanup()
		f.FIsRepeated = true
		actual := makePrefix(prefix, f, []string{"Foo", "Bar"}, true)
		require.Equal(t, expected, actual, false)
	})
}
