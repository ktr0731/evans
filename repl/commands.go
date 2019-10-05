package repl

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/usecase"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

var (
	errArgumentRequired = errors.New("argument required")
)

type commander interface {
	// Help returns a short help message.
	Help() string

	// Synopsis returns the usage of the command.
	Synopsis() string

	// Valdiate validates whether args satisfies preconditions for running the command.
	Validate(args []string) error

	// Run runs the command. The commander implementation writes something to w.
	// Caller must check no errors by calling Validate before call Run.
	Run(w io.Writer, args []string) error
}

type packageCommand struct{}

func (c *packageCommand) Synopsis() string {
	return "set a package as the currently selected package"
}

func (c *packageCommand) Help() string {
	return "usage: package <package name>"
}

func (c *packageCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *packageCommand) Run(_ io.Writer, args []string) error {
	err := usecase.UsePackage(args[0])
	if errors.Cause(err) == idl.ErrUnknownPackageName {
		return errors.Errorf("unknown package name '%s'", args[0])
	}
	return err
}

type serviceCommand struct{}

func (c *serviceCommand) Synopsis() string {
	return "set the service as the current selected service"
}

func (c *serviceCommand) Help() string {
	return "usage: service <service name>"
}

func (c *serviceCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *serviceCommand) Run(_ io.Writer, args []string) error {
	err := usecase.UseService(args[0])
	switch errors.Cause(err) {
	case idl.ErrPackageUnselected:
		return errors.New("package unselected. please execute 'package' command at the first")
	case idl.ErrUnknownServiceName:
		return errors.Errorf("unknown package name '%s'", args[0])
	}
	return err
}

type showCommand struct{}

func (c *showCommand) Synopsis() string {
	return "show package, service or RPC names"
}

func (c *showCommand) Help() string {
	return "usage: show <package | service | message | rpc | header>"
}

func (c *showCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *showCommand) Run(w io.Writer, args []string) error {
	target := args[0]

	table := tablewriter.NewWriter(w)

	var rows [][]string
	switch strings.ToLower(target) {
	case "p", "package", "packages":
		// pkgs := usecase.ListPackages()
		// table.SetHeader([]string{"package"})
		// for _, pkg := range pkgs {
		// 	rows = append(rows, []string{pkg})
		// }
		out, err := usecase.FormatPackages()
		if err != nil {
			return errors.Wrap(err, "failed to format packages")
		}
		io.WriteString(w, out)
		return nil
	case "s", "svc", "service", "services":
		svcs, err := usecase.ListServices()
		if err != nil {
			return errors.Wrap(err, "failed to list services belong to the package")
		}
		table.SetHeader([]string{"service", "rpc", "request type", "response type"})
		for _, svc := range svcs {
			rpcs, err := usecase.ListRPCs(svc)
			if err != nil {
				panic(fmt.Sprintf("ListRPCs must not return an error, but got '%s'", err))
			}
			for _, rpc := range rpcs {
				rows = append(rows, []string{svc, rpc.Name, rpc.RequestType.Name, rpc.ResponseType.Name})
			}
		}
	case "m", "msg", "message", "messages":
		svcs, err := usecase.ListServices()
		if err != nil {
			return errors.Wrap(err, "failed to list services belong to the package")
		}
		table.SetHeader([]string{"message"})
		encountered := make(map[string]interface{})
		for _, svc := range svcs {
			rpcs, err := usecase.ListRPCs(svc)
			if err != nil {
				panic(fmt.Sprintf("ListRPCs must not return an error, but got '%s'", err))
			}
			for _, rpc := range rpcs {
				if _, found := encountered[rpc.RequestType.Name]; !found {
					rows = append(rows, []string{rpc.RequestType.Name})
					encountered[rpc.RequestType.Name] = nil
				}
				if _, found := encountered[rpc.ResponseType.Name]; !found {
					rows = append(rows, []string{rpc.ResponseType.Name})
					encountered[rpc.ResponseType.Name] = nil
				}
			}
		}
	case "a", "r", "rpc", "api":
		svcs, err := usecase.ListServices()
		if err != nil {
			return errors.Wrap(err, "failed to list services belong to the package")
		}
		table.SetHeader([]string{"RPC", "request type", "response type"})
		for _, svc := range svcs {
			rpcs, err := usecase.ListRPCs(svc)
			if err != nil {
				panic(fmt.Sprintf("ListRPCs must not return an error, but got '%s'", err))
			}
			for _, rpc := range rpcs {
				rows = append(rows, []string{rpc.Name, rpc.RequestType.Name, rpc.ResponseType.Name})
			}
		}
	case "h", "header", "headers":
		headers := usecase.ListHeaders()
		table.SetHeader([]string{"key", "val"})
		for k, v := range headers {
			for _, vv := range v {
				rows = append(rows, []string{k, vv})
			}
		}
	default:
		return errors.Errorf("unknown target '%s'", target)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	table.AppendBulk(rows)
	table.Render()
	return nil
}

type callCommand struct{}

func (c *callCommand) Synopsis() string {
	return "call a RPC"
}

func (c *callCommand) Help() string {
	return "usage: call <RPC name>"
}

func (c *callCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *callCommand) Run(w io.Writer, args []string) error {
	err := usecase.CallRPC(context.Background(), w, args[0])
	if err == io.EOF {
		return errors.New("inputting canceled")
	}
	return err
}

type headerCommand struct{}

func (c *headerCommand) Synopsis() string {
	return "set/unset headers to each request. if header value is empty, the header is removed."
}

func (c *headerCommand) Help() string {
	return "usage: header <key>=<value>[, <key>=<value>...]"
}

func (c *headerCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *headerCommand) Run(_ io.Writer, args []string) error {
	headers := usecase.ListHeaders()
	for _, h := range args {
		sp := strings.SplitN(h, "=", 2)

		// Remove the key.
		if len(sp) == 1 || sp[1] == "" {
			headers.Remove(sp[0])
			continue
		}

		for _, v := range strings.Split(sp[1], ",") {
			if err := headers.Add(sp[0], v); err != nil {
				return errors.Wrapf(err, "failed to add a header '%s=%s'", sp[0], v)
			}
		}
	}
	return nil
}

type exitCommand struct{}

func (c *exitCommand) Synopsis() string {
	return "exit current REPL"
}

func (c *exitCommand) Help() string {
	return "usage: exit"
}

func (c *exitCommand) Validate([]string) error { return nil }

func (c *exitCommand) Run(io.Writer, []string) error {
	return io.EOF
}
