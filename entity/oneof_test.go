package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneOf(t *testing.T) {
	d := parseFile(t, "oneof.proto")
	msgs := d.GetMessageTypes()
	require.Len(t, msgs, 1)

	m := newMessage(msgs[0])
	require.Equal(t, m.name(), "Example")

	assert.Len(t, m.OneOfs, 1)

	oneof := m.OneOfs[0]
	assert.Equal(t, "oneof_example", oneof.name())
	assert.Equal(t, "", oneof.typ())
	assert.Len(t, oneof.Choices, 2)

	expected := []string{"makise", "shiina"}
	for i, v := range oneof.Choices {
		assert.Equal(t, expected[i], v.Number())
	}
}
