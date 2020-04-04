package mode

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_tidyUpHistory(t *testing.T) {
	cases := map[string]struct {
		history     []string
		historySize int
		expected    []string
	}{
		"empty": {
			history:     nil,
			historySize: 100,
			expected:    []string{},
		},
		"simple": {
			history:     []string{"foo", "bar"},
			historySize: 100,
			expected:    []string{"foo", "bar"},
		},
		"remove duplicated items": {
			history:     []string{"foo", "bar", "foo", "baz"},
			historySize: 100,
			expected:    []string{"bar", "foo", "baz"},
		},
		"over history size": {
			history:     []string{"foo", "bar", "baz"},
			historySize: 2,
			expected:    []string{"bar", "baz"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := tidyUpHistory(c.history, c.historySize)
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("-want, +got\n%s", diff)
			}
		})
	}
}
