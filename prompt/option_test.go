package prompt

import (
	"testing"
)

func TestOption(t *testing.T) {
	expected := opt{
		commandHistory: []string{"foo", "bar"},
	}
	var opt opt
	opts := []Option{
		WithCommandHistory(expected.commandHistory),
	}
	for _, o := range opts {
		o(&opt)
	}

	if len(expected.commandHistory) != len(opt.commandHistory) {
		t.Errorf("expected: %s, but got %s", expected.commandHistory, opt.commandHistory)
	}
}
