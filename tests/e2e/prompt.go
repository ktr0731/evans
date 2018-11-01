package e2e

import (
	goprompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/tests/helper"
)

// SetPrompt replaces NewPrompt var by newPrompt which is prompter injected.
// SetPrompt returns cleanup func as the result.
// caller must call cleanup after each tests.
func SetPrompt(pmt *helper.MockPrompt) func() {
	old := prompt.New
	prompt.New = func(executor func(string), completer func(goprompt.Document) []goprompt.Suggest, opt ...goprompt.Option) prompt.Prompt {
		if executor == nil {
			executor = func(in string) {
				return
			}
		}
		if completer == nil {
			completer = func(d goprompt.Document) []goprompt.Suggest {
				return nil
			}
		}
		pmt.Executor = executor
		pmt.Completer = completer
		p := pmt
		return p
	}
	return func() {
		prompt.New = old
	}
}
