package protobuf

import (
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseRepeatedProto(t *testing.T) []*desc.MessageDescriptor {
	d := parseFile(t, []string{"repeated.proto"}, nil)
	require.Len(t, d, 1)

	msgs := d[0].GetMessageTypes()
	require.Len(t, msgs, 5)

	return msgs
}

func TestField(t *testing.T) {
	t.Run("repeated", func(t *testing.T) {
		msgs := parseRepeatedProto(t)
		helloRequest := newMessage(msgs[0])

		fields := helloRequest.Fields()
		require.Len(t, fields, 1)

		assert.True(t, fields[0].IsRepeated())
	})

	t.Run("repeated enum", func(t *testing.T) {
		msgs := parseRepeatedProto(t)
		repeatedEnum := newMessage(msgs[4])

		fields := repeatedEnum.Fields()
		require.Len(t, fields, 1)

		assert.True(t, fields[0].IsRepeated())
	})
}
