package main

import (
	"fmt"

	arg "github.com/alexflint/go-arg"

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

type CLI struct {
	meta *Meta
	ui   *UI
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
	}
}

func (c *CLI) Error(err error) {
	fmt.Fprintln(c.ui.ErrWriter, err)
}

type Options struct {
	Port int `arg:"-p,help:gRPC port"`
}

func (c *CLI) Run(args []string) int {
	opts := Options{
		Port: 50051,
	}

	arg.MustParse(&opts)

	return 0
}
