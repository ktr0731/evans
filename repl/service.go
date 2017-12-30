package repl

import (
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

type serviceCommand struct {
	inputPort port.InputPort
}

func (c *serviceCommand) Synopsis() string {
	return "set the service as the current selected service"
}

func (c *serviceCommand) Help() string {
	return "usage: service <service name>"
}

func (c *serviceCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "service name")
	}
	return nil
}

func (c *serviceCommand) Run(args []string) (string, error) {
	_, err := c.inputPort.Call(nil)
	if err != nil {
		return "", err
	}
	return "", nil
}
