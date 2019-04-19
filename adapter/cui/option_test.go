package cui

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	cases := map[string]struct {
		r        io.Reader
		expected io.Reader
	}{
		"normal": {
			r:        os.Stdin,
			expected: os.Stdin,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ui := &basicUI{}
			Reader(c.r)(ui)
			assert.Equal(t, c.expected, ui.reader)
		})
	}
}

func TestWriter(t *testing.T) {
	cases := map[string]struct {
		w        io.Writer
		expected io.Writer
	}{
		"normal": {
			w:        os.Stdout,
			expected: os.Stdout,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ui := &basicUI{}
			Writer(c.w)(ui)
			assert.Equal(t, c.expected, ui.writer)
		})
	}
}

func TestErrWriter(t *testing.T) {
	cases := map[string]struct {
		ew       io.Writer
		expected io.Writer
	}{
		"normal": {
			ew:       os.Stderr,
			expected: os.Stderr,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ui := &basicUI{}
			ErrWriter(c.ew)(ui)
			assert.Equal(t, c.expected, ui.errWriter)
		})
	}
}
