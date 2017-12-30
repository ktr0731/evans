package repl

import (
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

type packageCommand struct {
	inputPort port.InputPort
}

func (c *packageCommand) Synopsis() string {
	return "set the package as the current selected package"
}

func (c *packageCommand) Help() string {
	return "usage: package <package name>"
}

func (c *packageCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "package name")
	}
	return nil
}

func (c *packageCommand) Run(args []string) (string, error) {
	params := &port.PackageParams{args[0]}
	_, err := c.inputPort.Package(params)
	if err != nil {
		return "", errors.Wrapf(err, "package: %s", args[0])
	}
	return "", nil
}
