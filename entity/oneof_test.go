package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOneOf(t *testing.T) {
	d := parseFile(t, "oneof.proto")
	msgs := d.GetMessageTypes()
	require.Len(t, msgs, 2)

	m := newMessage(msgs[0])
	require.Equal(t, m.Name(), "Example")

}
