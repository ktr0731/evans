package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnum(t *testing.T) {
	d := parseFile(t, "enum.proto")
	enum := d.GetEnumTypes()
	assert.Len(t, enum, 1)

	e := newEnum(enum[0])
	assert.Equal(t, e.Name(), "BookType")

	expected := []string{"EARLY", "PHILOSOPHY", "HISTORY", "SCIENCE"}
	for i, v := range e.Values {
		assert.Equal(t, expected[i], v.Name())
		assert.Equal(t, int32(i), v.Number())
	}
}
