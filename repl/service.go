package repl

import (
	"github.com/ktr0731/evans/env"
	"github.com/pkg/errors"
)

type ServiceCommand struct {
	env *env.Env
}

func (c *ServiceCommand) Help() string {
	return "Usage: service <service name>"
}

func (c *ServiceCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "service name")
	}
	return nil
}

func (c *ServiceCommand) Run(args []string) (string, error) {
	if err := c.env.UseService(args[0]); err != nil {
		return "", err
	}
	return "", nil
}
