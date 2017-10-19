package cli

import (
	"fmt"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/env"
	"github.com/ktr0731/evans/parser"
	"github.com/ktr0731/evans/repl"
	"github.com/pkg/errors"

	"io"
	"os"
)

type UI struct {
	Reader            io.Reader
	Writer, ErrWriter io.Writer
}

func NewUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

type Options struct {
	Proto []string `arg:"positional,help:.proto files"`

	Interactive bool   `arg:"-i,help:use interactive mode"`
	EditConfig  bool   `arg:"-e,help:edit config file by $EDITOR"`
	Port        int    `arg:"-p,help:gRPC port"`
	Package     string `arg:"help:default package"`
	Service     string `arg:"help:default service. evans parse package from this if --package is nothing."`
}

func (o *Options) Version() string {
	return "evans 0.1.0"
}

type CLI struct {
	ui     *UI
	config *config.Config

	parser  *arg.Parser
	options *Options
}

func NewCLI(title, version string) *CLI {
	return &CLI{
		ui: NewUI(),
		options: &Options{
			Port: 50051,
		},
	}
}

func (c *CLI) Error(err error) {
	fmt.Fprintln(c.ui.ErrWriter, err)
}

func (c *CLI) Help() {
	c.parser.WriteHelp(c.ui.Writer)
}

func (c *CLI) Usage() {
	c.parser.WriteUsage(c.ui.Writer)
}

func (c *CLI) Run(args []string) int {
	c.parser = arg.MustParse(c.options)

	if c.options.EditConfig {
		err := config.Edit()
		if err != nil {
			panic(err)
		}
		return 0
	}

	if len(c.options.Proto) == 0 {
		c.Error(errors.New("invalid argument"))
		return 1
	}

	desc, err := parser.ParseFile(c.options.Proto, []string{})
	if err != nil {
		c.Error(err)
		return 1
	}

	config := &repl.Config{
		Port: c.options.Port,
	}
	env, err := env.New(desc, config.Port)
	if err != nil {
		c.Error(err)
		return 1
	}

	if c.options.Package != "" {
		if err := env.UsePackage(c.options.Package); err != nil {
			c.Error(errors.Wrapf(err, "file %s", strings.Join(c.options.Proto, ", ")))
			return 1
		}
	}
	if c.options.Service != "" {
		if err := env.UseService(c.options.Service); err != nil {
			c.Error(errors.Wrapf(err, "file %s", strings.Join(c.options.Proto, ", ")))
			return 1
		}
	}

	r := repl.NewREPL(config, env, repl.NewUI())
	defer r.Close()

	if err := r.Start(); err != nil {
		c.Error(err)
		return 1
	}

	return 0
}
