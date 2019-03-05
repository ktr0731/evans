package inputter

import (
	"fmt"
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity/testentity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPrompt_Input(t *testing.T) {
	const prefixFormat = ">"

	t.Run("normal/simple", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "helloworld.proto", "helloworld", "Greeter")

		p := helper.NewMockPrompt([]string{"rin", "shima"}, nil)
		inputter := newPromptInputter(p, prefixFormat, env)

		rpc, err := env.RPC("SayHello")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestMessage())
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `name:"rin" message:"shima"`, msg.String())
	})

	t.Run("normal/nested_message", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "nested.proto", "library", "Library")

		p := helper.NewMockPrompt([]string{"eriri", "spencer", "sawamura"}, nil)
		inputter := newPromptInputter(p, prefixFormat, env)

		rpc, err := env.RPC("BorrowBook")
		require.NoError(t, err)

		dmsg, err := inputter.Input(rpc.RequestMessage())
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `person:<name:"eriri"> book:<title:"spencer" author:"sawamura">`, msg.String())
	})

	t.Run("normal/enum", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		p := helper.NewMockPrompt(nil, []string{"PHILOSOPHY"})
		inputter := newPromptInputter(p, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "enum.proto")
		packages, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := packages[0].Messages[0]

		dmsg, err := inputter.Input(m)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `type:PHILOSOPHY`, msg.String())
	})

	t.Run("error/enum:invalid enum name", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "enum.proto", "library", "")

		p := helper.NewMockPrompt(nil, []string{"kumiko"})
		inputter := newPromptInputter(p, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "enum.proto")
		packages, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := packages[0].Messages[0]

		_, err = inputter.Input(m)
		e := errors.Cause(err)
		require.Equal(t, ErrUnknownEnumName, e)
	})

	t.Run("normal/oneof", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "oneof.proto", "shop", "")

		p := helper.NewMockPrompt([]string{"utaha", "kasumigaoka"}, []string{"book"})
		inputter := newPromptInputter(p, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "oneof.proto")
		packages, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)

		m := packages[0].Messages[2]
		require.Equal(t, m.Name(), "BorrowRequest")
		require.Len(t, m.Fields(), 1)

		dmsg, err := inputter.Input(m)
		require.NoError(t, err)

		msg, ok := dmsg.(*dynamic.Message)
		require.True(t, ok)

		require.Equal(t, `book:<title:"utaha" author:"kasumigaoka">`, msg.String())
	})

	t.Run("error/oneof:invalid oneof field name", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "oneof.proto", "shop", "")

		p := helper.NewMockPrompt([]string{"bar"}, []string{"Book"})
		inputter := newPromptInputter(p, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "oneof.proto")
		packages, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := packages[0].Messages[2]

		_, err = inputter.Input(m)

		e := errors.Cause(err)
		require.Equal(t, ErrUnknownOneofFieldName, e)
	})

	t.Run("normal/repeated", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "repeated.proto", "helloworld", "")

		p := helper.NewMockRepeatedPrompt([][]string{
			{"foo", "", "bar", "", ""},
		}, nil)

		cleanup := injectNewPrompt(p)
		defer cleanup()

		inputter := newPromptInputter(p, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "repeated.proto")
		packages, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := packages[0].Messages[0]
		require.Equal(t, "HelloRequest", m.Name())

		msg, err := inputter.Input(m)
		require.NoError(t, err)

		require.Equal(t, `name:"foo" name:"bar"`, msg.String())
	})

	t.Run("normal/map", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "map.proto", "example", "")

		prompt := helper.NewMockRepeatedPrompt([][]string{
			{"foo", "", "bar", "", ""},
		}, nil)

		cleanup := injectNewPrompt(prompt)
		defer cleanup()

		inputter := newPromptInputter(prompt, prefixFormat, env)

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

		prompt := helper.NewMockRepeatedPrompt([][]string{
			{"key", "", "val1", "3", "", ""},
		}, nil)

		cleanup := injectNewPrompt(prompt)
		defer cleanup()

		inputter := newPromptInputter(prompt, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "map.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[2]
		require.Equal(t, "MessageRequest", m.Name())

		msg, err := inputter.Input(m)
		require.NoError(t, err)

		require.Equal(t, `foo:<key:"key" value:<fuga:"val1" piyo:3>>`, msg.String())
	})

	t.Run("normal/circulated", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "circulated.proto", "example", "ExampleService")

		prompt := helper.NewMockRepeatedPrompt(
			[][]string{{"filter1"}, {"and1"}, {"or1"}, {"or2"}, {"10", "1"}},
			[][]string{{"dig down"}, {"dig down"}, {"finish"}, {"finish"}, {"finish"}, {"dig down"}, {"finish"}, {"finish"}, {"dig down"}, {"finish"}, {"finish"}, {"finish"}})

		cleanup := injectNewPrompt(prompt)
		defer cleanup()

		inputter := newPromptInputter(prompt, prefixFormat, env)

		descs := testhelper.ReadProtoAsFileDescriptors(t, "circulated.proto")
		p, err := protobuf.ToEntitiesFrom(descs)
		require.NoError(t, err)
		m := p[0].Messages[1]
		require.Equal(t, "FooRequest", m.Name())

		msg, err := inputter.Input(m)
		require.NoError(t, err)

		// Expected:
		//
		// {
		//   "filters": {
		//     "name": "filter1",
		//     "and": [
		//       {
		//         "name": "and1"
		//       }
		//     ],
		//     "or": [
		//       {
		//         "name": "or1"
		//       },
		//       {
		//         "name": "or2"
		//       }
		//     ]
		//   },
		//   "page": 10,
		//   "limit": 1
		// }
		require.Equal(t, `filters:<name:"filter1" and:<name:"and1"> or:<name:"or1"> or:<name:"or2">> page:10 limit:1`, msg.String())
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

func injectNewPrompt(p prompt.Prompt) func() {
	old := prompt.New
	prompt.New = func(func(string), func(goprompt.Document) []goprompt.Suggest, ...goprompt.Option) prompt.Prompt {
		return p
	}
	return func() {
		prompt.New = old
	}
}
