package prompt_test

import (
	"testing"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/stretchr/testify/assert"
)

func TestPrompt(t *testing.T) {
	cases := map[string]struct {
		executor  func(string)
		completer func(goprompt.Document) []goprompt.Suggest
		assert    func(t *testing.T, p prompt.Prompt)
	}{
		"normal": {
			assert: func(t *testing.T, p prompt.Prompt) {
				assert.Panics(t,
					func() {
						p.Run()
					},
					"Run must panic if executor is nil",
				)
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			prompt.New(c.executor, c.completer)
		})
	}
}
