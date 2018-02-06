package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestField(t *testing.T) {
	t.Run("repeated", func(t *testing.T) {
		d := parseFile(t, "repeated.proto")
		msgs := d.GetMessageTypes()
		require.Len(t, msgs, 4)

		m := newMessage(msgs[0])

		fields := m.Fields()
		require.Len(t, fields, 1)
	})
}
