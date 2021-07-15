package repl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ktr0731/evans/format"
	"github.com/ktr0731/evans/format/curl"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/usecase"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var (
	errArgumentRequired = errors.New("argument required")
)

type commander interface {
	// Help returns a short help message.
	Help() string

	// Synopsis returns the usage of the command.
	Synopsis() string

	// FlagSet returns a flagset.
	// If the command doesn't have a flagset, the second returned value is false.
	// Note that FlagSet is for read-only, so don't modify it.
	FlagSet() (*pflag.FlagSet, bool)

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

func (c *packageCommand) FlagSet() (*pflag.FlagSet, bool) {
	return nil, false
}

func (c *packageCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *packageCommand) Run(_ io.Writer, args []string) error {
	pkgName := args[0]
	err := usecase.UsePackage(pkgName)
	if errors.Is(err, idl.ErrUnknownPackageName) {
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

func (c *serviceCommand) FlagSet() (*pflag.FlagSet, bool) {
	return nil, false
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
		return errors.Errorf("unknown service name '%s'", args[0])
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

func (c *showCommand) FlagSet() (*pflag.FlagSet, bool) {
	return nil, false
}

func (c *showCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *showCommand) Run(w io.Writer, args []string) error {
	target := args[0]

	var f func() (string, error)

	switch strings.ToLower(target) {
	case "p", "package", "packages":
		f = usecase.FormatPackages
	case "s", "svc", "service", "services":
		f = usecase.FormatServicesOld
	case "m", "msg", "message", "messages":
		f = usecase.FormatMessages
	case "a", "r", "rpc", "api":
		f = usecase.FormatMethods
	case "h", "header", "headers":
		f = usecase.FormatHeaders
	default:
		return errors.Errorf("unknown target '%s'", target)
	}

	out, err := f()
	if err != nil {
		return errors.Wrap(err, "failed to format")
	}
	if _, err := io.WriteString(w, out); err != nil {
		return errors.Wrap(err, "failed to write formatted output to w")
	}

	return nil
}

type callCommand struct {
	enrich, digManually, bytesFromFile, emitDefaults, repeatCall, addRepeatedManually bool
}

func (c *callCommand) FlagSet() (*pflag.FlagSet, bool) {
	fs := pflag.NewFlagSet("call", pflag.ContinueOnError)
	fs.Usage = func() {} // Disable help output when an error occurred.
	fs.BoolVar(&c.enrich, "enrich", false, "enrich response output includes header, message, trailer and status")
	fs.BoolVar(&c.digManually, "dig-manually", false, "prompt asks whether to dig down if it encountered to a message field")
	fs.BoolVar(&c.bytesFromFile, "bytes-from-file", false, "interpret TYPE_BYTES input as a relative path to a file")
	fs.BoolVar(&c.emitDefaults, "emit-defaults", false, "render fields with default values")
	fs.BoolVarP(&c.repeatCall, "repeat", "r", false, "repeat previous unary or server streaming request (if exists)")
	fs.BoolVar(&c.addRepeatedManually, "add-repeated-manually", false, "prompt asks whether to add a value if it encountered to a repeated field")
	return fs, true
}

func (c *callCommand) Synopsis() string {
	return "call a RPC"
}

func (c *callCommand) Help() string {
	var buf bytes.Buffer
	fs, _ := c.FlagSet()
	fs.SetOutput(&buf)
	fs.PrintDefaults()
	return fmt.Sprintf(`usage: call <method name>

Options:
%s`, strings.TrimRightFunc(buf.String(), unicode.IsSpace))
}

func (c *callCommand) Validate(args []string) error {
	if len(args) < 1 {
		return errArgumentRequired
	}
	return nil
}

func (c *callCommand) Run(w io.Writer, args []string) error {
	usecase.InjectPartially(
		usecase.Dependencies{
			ResponseFormatter: format.NewResponseFormatter(curl.NewResponseFormatter(w, c.emitDefaults), c.enrich),
		},
	)

	// here we create the request context
	// we also add the call command flags here
	err := usecase.CallRPCInteractively(context.Background(), w, args[0], c.digManually, c.bytesFromFile, c.repeatCall, c.addRepeatedManually)
	if errors.Is(err, io.EOF) {
		return errors.New("inputting canceled")
	}
	return err
}

type headerCommand struct {
	raw bool
}

func (c *headerCommand) FlagSet() (*pflag.FlagSet, bool) {
	fs := pflag.NewFlagSet("header", pflag.ContinueOnError)
	fs.Usage = func() {} // Disable help output when an error occurred.
	fs.BoolVarP(&c.raw, "raw", "r", false, "treat the value as a raw string")
	return fs, true
}

func (c *headerCommand) Synopsis() string {
	return "set/unset headers to each request. if header value is empty, the header is removed."
}

func (c *headerCommand) Help() string {
	var buf bytes.Buffer
	fs, _ := c.FlagSet()
	fs.SetOutput(&buf)
	fs.PrintDefaults()
	return fmt.Sprintf(`usage: header [options ...] <key>=<value>[, <key>=<value>...]

Options:
%s`, strings.TrimRightFunc(buf.String(), unicode.IsSpace))
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

		if c.raw {
			if err := headers.Add(sp[0], sp[1]); err != nil {
				return errors.Wrapf(err, "failed to add a header '%s=%s'", sp[0], sp[1])
			}
			return nil
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

func (c *exitCommand) FlagSet() (*pflag.FlagSet, bool) {
	return nil, false
}

func (c *exitCommand) Validate([]string) error { return nil }

func (c *exitCommand) Run(io.Writer, []string) error {
	return io.EOF
}
