package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lycoris0731/evans/lib/parser"
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
	env    *Env
	liner  *liner.State
}

type Config struct {
	Port int
}

type Env struct {
	desc *parser.FileDescriptorSet
}

func NewREPL(config *Config, env *Env) *REPL {
	return &REPL{
		ui:     NewUI(),
		config: config,
		env:    env,
		liner:  liner.NewLiner(),
	}
}

func (r *REPL) Read() (string, error) {
	l, err := r.liner.Prompt(fmt.Sprintf("127.0.0.1:%d> ", r.config.Port))
	if err == nil {
		// TODO: 書き出し
		r.liner.AppendHistory(l)
	}
	return l, err
}

func (r *REPL) Eval(l string) (string, error) {
	part := strings.Split(l, " ")
	switch part[0] {
	case "show":
		switch part[1] {
		case "svc", "service", "services":
			return r.env.desc.GetServices().String(), nil
		}
	}
	return "", nil
}

func (r *REPL) Print(text string) {
	fmt.Fprintf(r.ui.Writer, "%s\n", text)
}

func (r *REPL) Start() error {
	defer func() {
		r.Print("Bye!")
		if err := r.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	for {
		l, err := r.Read()

		if err == io.EOF || l == CmdQuit || l == CmdExit {
			if err == io.EOF {
				fmt.Println()
			}
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to read line")
		}

		result, err := r.Eval(l)
		if err != nil {
			return errors.Wrap(err, "failed to evaluate line")
		}
		r.Print(result)
	}
	return nil
}

func (r *REPL) Close() error {
	return r.liner.Close()
}
