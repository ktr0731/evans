package presenter

import (
	"io/ioutil"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestCLIPresenter(t *testing.T) {
	presenter := NewJSONCLIPresenter()

	t.Run("Call", func(t *testing.T) {
		descs := testhelper.ReadProtoAsFileDescriptors(t, "helloworld.proto")
		msg := testhelper.FindMessage(t, "HelloRequest", descs)

		dmsg := dynamic.NewMessage(msg)
		dmsg.SetField(dmsg.FindFieldDescriptorByName("name"), "makise")
		dmsg.SetField(dmsg.FindFieldDescriptorByName("message"), "kurisu")

		out, err := presenter.Call(dmsg)
		require.NoError(t, err)

		b, err := ioutil.ReadAll(out)
		require.NoError(t, err)

		require.Equal(t, `{"name":"makise","message":"kurisu"}`+"\n", string(b))
	})
}

func TestJSONCLIPresenterWithIndent(t *testing.T) {
	presenter := NewJSONCLIPresenterWithIndent()

	t.Run("Call", func(t *testing.T) {
		descs := testhelper.ReadProtoAsFileDescriptors(t, "helloworld.proto")
		msg := testhelper.FindMessage(t, "HelloRequest", descs)

		dmsg := dynamic.NewMessage(msg)
		dmsg.SetField(dmsg.FindFieldDescriptorByName("name"), "makise")
		dmsg.SetField(dmsg.FindFieldDescriptorByName("message"), "kurisu")

		out, err := presenter.Call(dmsg)
		require.NoError(t, err)

		b, err := ioutil.ReadAll(out)
		require.NoError(t, err)

		require.Equal(t, `{
  "name": "makise",
  "message": "kurisu"
}`+"\n", string(b))
	})
}
