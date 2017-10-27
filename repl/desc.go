package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type DescCommand struct {
	env *env.Env
}

func (c *DescCommand) Synopsis() string {
	return "Describe the structure of selected message"
}

func (c *DescCommand) Help() string {
	return "Usage: desc <message name>"
}

func (c *DescCommand) Validate(args []string) error {
	if len(args) < 2 {
		return errors.Wrap(ErrArgumentRequired, "message name")
	}
	return nil
}

func (c *DescCommand) Run(args []string) (string, error) {
	msg, err := c.env.GetMessage(args[0])
	if err != nil {
		return "", err
	}
	return msg.String(), nil
}
