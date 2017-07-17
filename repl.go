package main

import (
	"fmt"
	"io"
	"os"

	"github.com/peterh/liner"
	"github.com/pkg/errors"
)

var (
	CmdQuit = "quit"
	CmdExit = "exit"
)

type REPL struct {
	ui     *UI
	config *Config
	liner  *liner.State
}

type Config struct {
	Port int
}

func NewREPL(config *Config) *REPL {
	return &REPL{
		ui:     NewUI(),
		config: config,
		liner:  liner.NewLiner(),
	}
}

func (r *REPL) Read() (string, error) {
	return r.liner.Prompt(fmt.Sprintf("127.0.0.1:%d> ", r.config.Port))
}

func (r *REPL) Response(text string) {
	fmt.Fprintf(r.ui.Writer, "%s\n", text)
}

func (r *REPL) Start() error {
	defer func() {
		r.Response("Bye!")
		if err := r.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

REPL:
	for {
		cmd, err := r.Read()
		if err == io.EOF {
			fmt.Println()
			break REPL
		} else if err != nil {
			return errors.Wrap(err, "failed to read line")
		}

		switch cmd {
		case CmdQuit, CmdExit:
			break REPL
		default:
			r.Response(cmd)
		}
	}
	return nil
}

func (r *REPL) Close() error {
	return r.liner.Close()
}
