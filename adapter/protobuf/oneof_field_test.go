package protobuf

import (
	"testing"

	"github.com/ktr0731/evans/entity"
	"github.com/stretchr/testify/require"
)

func TestOneOfField(t *testing.T) {
	d := parseFile(t, []string{"oneof.proto"}, nil)
	require.Len(t, d, 1)

	m := newMessage(d[0].GetMessageTypes()[0])
	require.Equal(t, m.Name(), "Example")

	require.Len(t, m.Fields(), 1)

	f, ok := m.Fields()[0].(entity.OneOfField)
	require.True(t, ok)
	c := f.Choices()
	require.Len(t, c, 2)
	require.Equal(t, c[0].FieldName(), "makise")
	require.Equal(t, c[1].FieldName(), "shiina")
}
