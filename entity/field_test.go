package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	t.Run("repeated", func(t *testing.T) {
		d := parseFile(t, "repeated.proto")
		msgs := d.GetMessageTypes()
		assert.Len(t, msgs, 4)

		m := newMessage(msgs[0])

		fields := m.Fields
		assert.Len(t, fields, 1)

		assert.True(t, fields[0].IsRepeated())
	})
}
