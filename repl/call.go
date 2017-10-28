package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type callCommand struct {
	env *env.Env
}

func (c *callCommand) Synopsis() string {
	return "Call a RPC with interactively input"
}

func (c *callCommand) Help() string {
	return "Usage: call <RPC name>"
}

func (c *callCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "service or RPC name")
	}
	return nil
}

func (c *callCommand) Run(args []string) (string, error) {
	return c.env.Call(args[0])
}
