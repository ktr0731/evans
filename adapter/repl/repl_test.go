package repl

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
)

var cacheDir string

func init() {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	cacheDir = dir
	os.Setenv("XDG_CACHE_HOME", cacheDir)
}

func Test_repl_cleanup(t *testing.T) {
	defer os.RemoveAll(cacheDir)

	// Note that previous command histories are cached.
	cases := []struct {
		name        string
		historySize int
		inputs      []string
		expected    []string
	}{
		{
			name:        "no inputs",
			historySize: 3,
			expected:    []string{},
		},
		{
			name:        "command history size is less than max size",
			historySize: 3,
			inputs:      []string{"miyuki", "kaguya"},
			expected:    []string{"miyuki", "kaguya"},
		},
		{
			name:        "command history size is greather than max size",
			historySize: 3,
			inputs:      []string{"miyuki", "kaguya", "chika", "yu"},
			expected:    []string{"kaguya", "chika", "yu"}, // The oldest command will be discarded.
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			prompt := helper.NewMockPrompt(c.inputs, nil)
			// Call Input for recording command history.
			for i := 0; i < len(c.inputs); i++ {
				prompt.Input()
			}
			assert.Len(t, prompt.History(), len(c.inputs), "prompt must hold Input history")

			r := &repl{
				config: &config.REPL{
					HistorySize: c.historySize,
				},
				prompt: prompt,
			}
			r.cleanup()

			newCache := cache.Get()
			assert.Equal(t, c.expected, newCache.CommandHistory)
		})
	}
}
