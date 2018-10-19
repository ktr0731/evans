package presenter

import (
	"io/ioutil"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doCall(t *testing.T, presenter *JSONPresenter) (res string) {
	descs := testhelper.ReadProtoAsFileDescriptors(t, "helloworld.proto")
	msg := testhelper.FindMessage(t, "HelloRequest", descs)

	dmsg := dynamic.NewMessage(msg)
	dmsg.SetField(dmsg.FindFieldDescriptorByName("name"), "makise")
	dmsg.SetField(dmsg.FindFieldDescriptorByName("message"), "kurisu")

	out, err := presenter.Call(dmsg)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(out)
	require.NoError(t, err)

	return string(b)
}

func TestJSONPresenter(t *testing.T) {
	presenter := NewJSON()

	res := doCall(t, presenter)
	assert.Equal(t, `{"name":"makise","message":"kurisu"}`+"\n", res)
}

func TestJSONPresenterWithIndent(t *testing.T) {
	presenter := NewJSONWithIndent()

	res := doCall(t, presenter)
	assert.Equal(t, `{
  "name": "makise",
  "message": "kurisu"
}`+"\n", res)
}
