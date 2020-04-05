package cui

import (
	"io"
	"os"
	"testing"
)

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
		c := c
		t.Run(name, func(t *testing.T) {
			ui := &basicUI{}
			Writer(c.w)(ui)
			if c.expected != ui.writer {
				t.Errorf("expected %T, but got %T", c.expected, ui.writer)
			}
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
		c := c

		t.Run(name, func(t *testing.T) {
			ui := &basicUI{}
			ErrWriter(c.ew)(ui)
			if c.expected != ui.errWriter {
				t.Errorf("expected %T, but got %T", c.expected, ui.errWriter)
			}
		})
	}
}
