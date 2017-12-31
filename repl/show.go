package repl

import (
	"strings"

	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

type showCommand struct {
	inputPort port.InputPort
}

func (c *showCommand) Synopsis() string {
	return "show package, service or RPC names"
}

func (c *showCommand) Help() string {
	return "usage: show <package | service | message | rpc>"
}

func (c *showCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "target type (package, service, message)")
	}
	return nil
}

func (c *showCommand) Run(args []string) (string, error) {
	target := args[0]

	params := &port.ShowParams{}
	switch strings.ToLower(target) {
	case "p", "package", "packages":
		params.Type = port.ShowTypePackage
	case "s", "svc", "service", "services":
		params.Type = port.ShowTypeService
	case "m", "msg", "message", "messages":
		params.Type = port.ShowTypeMessage
	case "a", "r", "rpc", "api":
		params.Type = port.ShowTypeRPC
	default:
		return "", errors.Wrap(ErrUnknownTarget, target)
	}
	_, err := c.inputPort.Show(params)
	if err != nil {
		return "", err
	}
	return "", nil
}
