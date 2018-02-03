package controller

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
	"github.com/pkg/errors"
)

type Commander interface {
	Help() string
	Synopsis() string
	Validate(args []string) error
	Run(args []string) (string, error)
}

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
	res, err := c.inputPort.Describe(params)
	if err != nil {
		return "", err
	}
	return read(res)
}

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
	res, err := c.inputPort.Package(params)
	if err != nil {
		return "", errors.Wrapf(err, "package: %s", args[0])
	}
	return read(res)
}

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
	params := &port.ServiceParams{args[0]}
	res, err := c.inputPort.Service(params)
	if err != nil {
		return "", err
	}
	return read(res)
}

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
	res, err := c.inputPort.Show(params)
	if err != nil {
		return "", err
	}
	return read(res)
}

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
	res, err := c.inputPort.Call(params)
	if err != nil {
		return "", err
	}
	return read(res)
}

type headerCommand struct {
	inputPort port.InputPort
}

func (c *headerCommand) Synopsis() string {
	return "set/unset headers to each request. if header value is empty, the header is removed."
}

func (c *headerCommand) Help() string {
	return "usage: header <key>=<value>[, <key>=<value>...]"
}

func (c *headerCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errors.Wrap(ErrArgumentRequired, "<key>=<value> or <key>")
	}
	return nil
}

func (c *headerCommand) Run(args []string) (string, error) {
	headers := []*entity.Header{}
	for _, h := range args {
		sp := strings.SplitN(h, "=", 2)
		header := &entity.Header{
			Key: sp[0],
		}
		for _, r := range sp[0] {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '.' {
				return "", fmt.Errorf("invalid char in key: %c", r)
			}
		}
		// delete the key
		if len(sp) == 1 || sp[1] == "" {
			header.NeedToDelete = true
		} else {
			header.Val = sp[1]
		}
		headers = append(headers, header)
	}
	params := &port.HeaderParams{headers}
	res, err := c.inputPort.Header(params)
	if err != nil {
		return "", err
	}
	return read(res)
}

func read(r io.Reader) (string, error) {
	if r == nil {
		return "", nil
	}

	b := new(bytes.Buffer)
	_, err := b.ReadFrom(r)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
