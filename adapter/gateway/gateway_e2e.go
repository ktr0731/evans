// +build e2e

// prompt_inputter_e2e.go provides public prompt APIs for e2e testing

package gateway

import (
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
)

// SetPromptForE2E replaces NewPrompt var by newPrompt which is prompter injected.
// SetPromptForE2E returns cleanup func as the result.
// caller must call cleanup after each tests.
func SetPromptForE2E(prompt Prompter) func() {
	old := NewPrompt
	NewPrompt = func(cfg *config.Config, env entity.Environment) *Prompt {
		return newPrompt(prompt, cfg, env)
	}
	return func() {
		NewPrompt = old
	}
}
