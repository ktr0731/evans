// +build e2e

// prompt_inputter_e2e.go provides public prompt APIs for e2e testing

package gateway

import (
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
)

// NewPromptForE2E exports newPrompt to other packages.
func NewPromptForE2E(cfg *config.Config, env entity.Environment, prompt Prompter) *Prompt {
	return newPrompt(prompt, cfg, env)
}
