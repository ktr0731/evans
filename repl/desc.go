package repl

import (
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

type descCommand struct {
	inputPort port.InputPort
}

func (c *descCommand) Synopsis() string {
	return "describe the structure of selected message"
}

func (c *descCommand) Help() string {
	return "usage: desc <message name>"
}

func (c *descCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "message name")
	}
	return nil
}

func (c *descCommand) Run(args []string) (string, error) {
	params := &port.DescribeParams{args[0]}
	_, err := c.inputPort.Describe(params)
	if err != nil {
		return "", err
	}
	return "", nil
}
