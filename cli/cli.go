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

func newUI() *UI {
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
		ui: newUI(),
		options: &Options{
			Port: 50051,
		},
		config: config.Get(),
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
		if err := config.Edit(); err != nil {
			c.Error(err)
			return 1
		}
		return 0
	}

	if err := checkPrecondition(c.options); err != nil {
		c.Error(err)
		return 1
	}

	env, err := setupEnv(c.config.Env, c.options)
	if err != nil {
		c.Error(err)
		return 1
	}

	r := repl.NewREPL(c.config.REPL, env, repl.NewBasicUI())
	defer r.Close()

	if err := r.Start(); err != nil {
		c.Error(err)
		return 1
	}

	return 0
}

func checkPrecondition(opt *Options) error {
	if len(opt.Proto) == 0 {
		return errors.New("invalid argument")
	}
	return nil
}

func setupEnv(config *config.Env, opt *Options) (*env.Env, error) {
	// TODO: 複数の path に対応する
	desc, err := parser.ParseFile(opt.Proto, []string{})
	if err != nil {
		return nil, err
	}

	env, err := env.New(desc, config)
	if err != nil {
		return nil, err
	}

	if opt.Package != "" {
		if err := env.UsePackage(opt.Package); err != nil {
			return nil, errors.Wrapf(err, "file %s", strings.Join(opt.Proto, ", "))
		}
	}
	if opt.Service != "" {
		if err := env.UseService(opt.Service); err != nil {
			return nil, errors.Wrapf(err, "file %s", strings.Join(opt.Proto, ", "))
		}
	}
	return env, nil
}
