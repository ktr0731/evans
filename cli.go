package main

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/lycoris0731/evans/lib/parser"
	"github.com/lycoris0731/evans/repl"

	"io"
	"os"
)

type Meta struct {
	Title, Version string
}

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

	Port        int  `arg:"-p,help:gRPC port"`
	Interactive bool `arg:"-i,help:use interactive mode"`
}

type CLI struct {
	meta    *Meta
	ui      *UI
	options *Options
}

func NewCLI(title, version string) *CLI {
	return &CLI{
		meta: &Meta{
			Title:   title,
			Version: version,
		},
		ui: NewUI(),
		options: &Options{
			Port: 50051,
		},
	}
}

func (c *CLI) Error(err error) {
	fmt.Fprintln(c.ui.ErrWriter, err)
}

func (c *CLI) Run(args []string) int {
	arg.MustParse(c.options)

	desc, err := parser.ParseFile(args, []string{})
	if err != nil {
		c.Error(err)
		return 1
	}

	config := &repl.Config{
		Port: c.options.Port,
	}
	env := &repl.Env{
		Desc: desc,
	}
	if err := repl.NewREPL(config, env, repl.NewUI()).Start(); err != nil {
		c.Error(err)
		return 1
	}

	return 0
}
