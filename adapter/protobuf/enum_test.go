package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnum(t *testing.T) {
	d := parseFile(t, "enum.proto")
	enum := d.GetEnumTypes()
	require.Len(t, enum, 1)

	e := newEnum(enum[0])
	require.Equal(t, e.Name(), "BookType")

	expected := []string{"EARLY", "PHILOSOPHY", "HISTORY", "SCIENCE"}
	for i, v := range e.Values() {
		require.Equal(t, expected[i], v.Name())
		require.Equal(t, int32(i), v.Number())
	}
}
