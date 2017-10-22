package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/env"
	"github.com/peterh/liner"
	"github.com/pkg/errors"
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
	config *config.REPL
	env    *env.Env
	liner  *liner.State
}

type UI struct {
	Reader            io.Reader
	Writer, ErrWriter io.Writer
	prompt            string
}

func NewBasicUI() *UI {
	return &UI{
		Reader:    os.Stdin,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

func NewREPL(config *config.REPL, env *env.Env, ui *UI) *REPL {
	repl := &REPL{
		ui:     ui,
		config: config,
		env:    env,
	}
	l := liner.NewLiner()
	l.SetCompleter(repl.GetCompletion)
	repl.liner = l

	return repl
}

func (r *REPL) Read() (string, error) {
	prompt := fmt.Sprintf("%s:%s> ", r.config.Server.Host, r.config.Server.Port)
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

	// TODO: 型定義して条件分岐を殺す
	switch part[0] {
	case "show":
		if len(part) < 2 || len(part[1]) == 0 {
			return "", errors.Wrap(ErrArgumentRequired, "target type (package, service, message)")
		}
		if part[1] == "-h" || part[1] == "--help" {
			return "Usage: show <package | service | message | rpc>", nil
		}
		return show(r.env, part[1])

	case "c", "call":
		if len(part) < 2 || len(part[1]) == 0 {
			return "", errors.Wrap(ErrArgumentRequired, "service or RPC name")
		}
		if part[1] == "-h" || part[1] == "--help" {
			return "Usage: call <RPC name>", nil
		}
		return r.env.Call(part[1])

	case "d", "desc", "describe":
		if len(part) < 2 || len(part[1]) == 0 {
			return "", errors.Wrap(ErrArgumentRequired, "message name")
		}
		if part[1] == "-h" || part[1] == "--help" {
			return "Usage: desc <message name>", nil
		}
		return describe(r.env, part[1])

	case "p", "package":
		if len(part) < 2 || len(part[1]) == 0 {
			return "", errors.Wrap(ErrArgumentRequired, "package name")
		}

		if part[1] == "-h" || part[1] == "--help" {
			return "Usage: package <package name>", nil
		}

		if err := r.env.UsePackage(part[1]); err != nil {
			return "", errors.Wrapf(err, "file %s", part[1])
		}

	case "s", "svc", "service":
		if len(part) < 2 || len(part[1]) == 0 {
			return "", errors.Wrap(ErrArgumentRequired, "service name")
		}

		if part[1] == "-h" || part[1] == "--help" {
			return "Usage: service <service name>", nil
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
	for {
		l, err := r.Read()
		if err == io.EOF || l == "quit" || l == "exit" {
			if err == io.EOF {
				fmt.Println()
			}
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "failed to read line")
		}

		if len(strings.TrimSpace(l)) == 0 {
			continue
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
	r.Print("Good Bye :)")
	return r.liner.Close()
}
