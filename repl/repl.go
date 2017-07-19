package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lycoris0731/evans/env"
	"github.com/peterh/liner"
	"github.com/pkg/errors"
)

var (
	CmdQuit = "quit"
	CmdExit = "exit"
)

var (
	ErrUnknownCommand   = errors.New("unknown command")
	ErrArgumentRequired = errors.New("argument required")
	ErrUnknownTarget    = errors.New("unknown target")
)

func newUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

type REPL struct {
	ui     *UI
	config *Config
	env    *env.Env
	liner  *liner.State
}

type UI struct {
	Reader            io.Reader
	Writer, ErrWriter io.Writer
	prompt            string
}

func NewUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

type Config struct {
	Port int
}

func NewREPL(config *Config, env *env.Env, ui *UI) *REPL {
	return &REPL{
		ui:     ui,
		config: config,
		env:    env,
		liner:  liner.NewLiner(),
	}
}

func (r *REPL) Read() (string, error) {
	prompt := fmt.Sprintf("127.0.0.1:%d> ", r.config.Port)
	if dsn := r.env.GetDSN(); dsn != "" {
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
			return "", errors.Wrap(ErrArgumentRequired, "target type (package, service, message)")
		}
		return show(r.env, part[1])

	case "c", "call":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
		}
		return call(r.env, part[1])

	case "d", "desc", "describe":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "message name")
		}
		return describe(r.env, part[1])

	case "p", "package":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "package name")
		}

		if err := r.env.UsePackage(part[1]); err != nil {
			return "", err
		}

	case "s", "svc", "service":
		if len(part) < 2 {
			return "", errors.Wrap(ErrArgumentRequired, "service name")
		}

		if err := r.env.UseService(part[1]); err != nil {
			return "", err
		}

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
		r.Print("Good Bye :)")
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
