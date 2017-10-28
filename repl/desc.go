package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type descCommand struct {
	env *env.Env
}

func (c *descCommand) Synopsis() string {
	return "describe the structure of selected message"
}

func (c *descCommand) Help() string {
	return "usage: desc <message name>"
}

func (c *descCommand) Validate(args []string) error {
	if len(args) < 2 {
		return errors.Wrap(ErrArgumentRequired, "message name")
	}
	return nil
}

func (c *descCommand) Run(args []string) (string, error) {
	msg, err := c.env.GetMessage(args[0])
	if err != nil {
		return "", err
	}
	return msg.String(), nil
}
