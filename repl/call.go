package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type CallCommand struct {
	env *env.Env
}

func (c *CallCommand) Synopsis() string {
	return "Call a RPC with interactively input"
}

func (c *CallCommand) Help() string {
	return "Usage: call <RPC name>"
}

func (c *CallCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "service or RPC name")
	}
	return nil
}

func (c *CallCommand) Run(args []string) (string, error) {
	return c.env.Call(args[0])
}
