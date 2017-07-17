package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lycoris0731/evans/lib/parser"
	"github.com/peterh/liner"
	"github.com/pkg/errors"
)

var (
	CmdQuit = "quit"
	CmdExit = "exit"
)

var (
	ErrUnknownCommand   = errors.New("unknown command")
	ErrUnknownTarget    = errors.New("unknown target")
	ErrUnknownPackage   = errors.New("unknown package")
	ErrUnknownService   = errors.New("unknown service")
	ErrArgumentRequired = errors.New("argument required")
)

type REPL struct {
	ui     *replUI
	config *Config
	env    *Env
	liner  *liner.State
	state  *State
}

type replUI struct {
	*UI
}

type State struct {
	currentPackage string
	currentService string
}

func (r *REPL) usePackage(name string) error {
	for _, p := range r.env.desc.GetPackages() {
		if name == p {
			r.state.currentPackage = name
			return nil
		}
	}
	return ErrUnknownPackage
}

func (r *REPL) useService(name string) error {
	for _, svc := range r.env.desc.GetServices() {
		if name == svc.Name {
			r.state.currentService = name
			return nil
		}
	}
	return ErrUnknownService
}

type Config struct {
	Port int
}

type Env struct {
	desc *parser.FileDescriptorSet
}

func NewREPL(config *Config, env *Env) *REPL {
	return &REPL{
		ui: &replUI{
			UI: NewUI(),
		},
		config: config,
		env:    env,
		liner:  liner.NewLiner(),
		state:  &State{},
	}
}

func (r *REPL) Read() (string, error) {
	prompt := fmt.Sprintf("127.0.0.1:%d> ", r.config.Port)
	if r.state.currentPackage != "" {
		dsn := r.state.currentPackage
		if r.state.currentService != "" {
			dsn += "." + r.state.currentService
		}
		prompt = fmt.Sprintf("%s@%s", dsn, prompt)
	}

	l, err := r.liner.Prompt(prompt)
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
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "target")
		}

		switch part[1] {
		case "svc", "service", "services":
			return r.env.desc.GetServices().String(), nil

		case "package", "packages":
			return r.env.desc.GetPackages().String(), nil

		default:
			return "", errors.Wrap(ErrUnknownTarget, part[1])
		}
	case "package":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "package name")
		}

		if err := r.usePackage(part[1]); err != nil {
			return "", err
		}

	case "service":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "service name")
		}

		if err := r.useService(part[1]); err != nil {
			return "", err
		}

	case "call":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
		}

		var svc, rpc string
		if r.state.currentService == "" {
			splitted := strings.Split(part[1], ".")
			if len(splitted) < 2 {
				return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
			}
			svc, rpc = splitted[0], splitted[1]
		} else {
			svc = r.state.currentService
			rpc = part[1]
		}
		fmt.Println(svc, rpc)

	default:
		return "", errors.Wrap(ErrUnknownCommand, part[0])
	}
	return "", nil
}

func (r *REPL) Print(text string) {
	fmt.Fprintf(r.ui.Writer, "%s\n", text)
}

func (r *REPL) Error(err error) {
	fmt.Fprintln(r.ui.ErrWriter, color.RedString(err.Error()))
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
			r.Error(err)
		} else {
			r.Print(result)
		}
	}
	return nil
}

func (r *REPL) Close() error {
	return r.liner.Close()
}
