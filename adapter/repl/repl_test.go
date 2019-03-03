package repl

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_repl_cleanup(t *testing.T) {
	setup := func(t *testing.T) func() {
		t.Helper()

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		oldCachePath := cache.Path
		return func() {
			os.RemoveAll(dir)
			cache.Path = oldCachePath
		}
	}

	cases := map[string]struct {
		historySize int
		inputs      []string
		expected    []string
	}{
		"command history size is greather than max size": {
			historySize: 3,
			inputs:      []string{"miyuki", "kaguya", "chika", "yu"},
			expected:    []string{"kaguya", "chika", "yu"}, // The oldest command will be discarded.
		},
		"command history size is less than max size": {
			historySize: 3,
			inputs:      []string{"miyuki", "kaguya"},
			expected:    []string{"miyuki", "kaguya"},
		},
		// "no inputs": {
		// 	historySize: 3,
		// 	expected:    []string{},
		// },
	}

	for name, c := range cases {
		name := name
		c := c

		t.Run(name, func(t *testing.T) {
			cleanup := setup(t)
			defer cleanup()

			prompt := helper.NewMockPrompt(c.inputs, nil)
			// Call Input for recording command history.
			for i := 0; i < len(c.inputs); i++ {
				prompt.Input()
			}

			r := &repl{
				config: &config.REPL{
					HistorySize: c.historySize,
				},
				prompt: prompt,
			}
			r.cleanup()

			newCache := cache.Get()
			assert.Equal(t, newCache.CommandHistory, c.expected)
		})
	}
}
