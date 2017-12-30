package repl

import (
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

type callCommand struct {
	inputPort port.InputPort
}

func (c *callCommand) Synopsis() string {
	return "call a RPC with interactively input"
}

func (c *callCommand) Help() string {
	return "usage: call <RPC name>"
}

func (c *callCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "service or RPC name")
	}
	return nil
}

func (c *callCommand) Run(args []string) (string, error) {
	params := &port.CallParams{args[0]}
	_, err := c.inputPort.Call(params)
	if err != nil {
		return "", err
	}
	return "", nil
}
