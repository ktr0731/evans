package protobuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnum(t *testing.T) {
	d := parseFile(t, []string{"enum.proto"}, nil)
	require.Len(t, d, 1)

	enum := d[0].GetEnumTypes()
	require.Len(t, enum, 1)

	e := newEnum(enum[0])
	require.Equal(t, e.Name(), "BookType")

	expected := []string{"EARLY", "PHILOSOPHY", "HISTORY", "SCIENCE"}
	for i, v := range e.Values() {
		require.Equal(t, expected[i], v.Name())
		require.Equal(t, int32(i), v.Number())
	}
}
