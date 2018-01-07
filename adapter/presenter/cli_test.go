package presenter

import (
	"io/ioutil"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIPresenter(t *testing.T) {
	presenter := NewJSONCLIPresenter()

	t.Run("Call", func(t *testing.T) {
		env := testhelper.SetupEnv(t, "helloworld.proto", "helloworld", "Greeter")

		msg, err := env.Message("HelloRequest")
		require.NoError(t, err)

		dmsg := dynamic.NewMessage(msg.Desc)
		dmsg.SetField(dmsg.FindFieldDescriptorByName("name"), "makise")
		dmsg.SetField(dmsg.FindFieldDescriptorByName("message"), "kurisu")

		out, err := presenter.Call(dmsg)
		require.NoError(t, err)

		b, err := ioutil.ReadAll(out)
		require.NoError(t, err)

		assert.Equal(t, `{"name":"makise","message":"kurisu"}`, string(b))
	})
}
