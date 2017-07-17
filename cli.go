package main

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
	"github.com/lycoris0731/go-grpc-client/lib"

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
		ui: &UI{
			Reader:    os.Stdin,
			Writer:    os.Stdout,
			ErrWriter: os.Stderr,
		},
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

	_, err := lib.ParseFile(args[0])
	if err != nil {
		c.Error(err)
	}

	return 0
}
