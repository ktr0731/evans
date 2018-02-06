package gateway

import (
	"bytes"
	"testing"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/stretchr/testify/require"
)

func TestJSONFileInputter(t *testing.T) {
	descs := testhelper.ReadProtoAsFileDescriptors(t, "helloworld.proto")
	m := testhelper.FindMessage(t, "HelloRequest", descs)

	jsonInput := `{
	"name": "ktr",
	"message": "hi"
}`

	msg := dynamic.NewMessage(m)
	err := msg.TrySetField(msg.FindFieldDescriptorByName("name"), "ktr")
	require.NoError(t, err)
	err = msg.TrySetField(msg.FindFieldDescriptorByName("message"), "hi")
	require.NoError(t, err)

	in := bytes.NewReader([]byte(jsonInput))
	inputter := NewJSONFileInputter(in)
	actual, err := inputter.Input(m)
	require.NoError(t, err)

	require.Exactly(t, actual, msg)
}
