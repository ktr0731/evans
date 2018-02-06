package gateway

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ktr0731/evans/adapter/internal/testhelper"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/stretchr/testify/require"
)

func TestJSONFileInputter(t *testing.T) {
	d := testhelper.ReadProtoAsFileDescriptors(t, "helloworld.proto")
	p, err := protobuf.ToEntitiesFrom(d)
	require.NoError(t, err)

	m := p[0].Messages[0]

	jsonInput := `{
	"name": "ktr",
	"message": "hi"
}`

	setter := protobuf.NewMessageSetter(m)
	err = setter.SetField(m.Fields()[0], "ktr")
	require.NoError(t, err)
	err = setter.SetField(m.Fields()[1], "hi")
	require.NoError(t, err)

	in := bytes.NewReader([]byte(jsonInput))
	inputter := NewJSONFileInputter(in)
	res, err := inputter.Input(m)
	require.NoError(t, err)

	marshaler := jsonpb.Marshaler{}
	actual, err := marshaler.MarshalToString(res)
	require.NoError(t, err)

	require.Exactly(t, actual, `{"name":"ktr","message":"hi"}`)
}
