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
	cmds   map[string]Commander
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
	call := &CallCommand{env}
	desc := &DescCommand{env}
	pkg := &PackageCommand{env}
	svc := &ServiceCommand{env}
	repl := &REPL{
		ui:     ui,
		config: config,
		env:    env,
		cmds: map[string]Commander{
			"show": &ShowCommand{env},

			"c":    call,
			"call": call,

			"d":        desc,
			"desc":     desc,
			"describe": desc,

			"p":       pkg,
			"package": pkg,

			"s":       svc,
			"svc":     svc,
			"service": svc,
		},
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

	cmd, ok := r.cmds[part[0]]
	if !ok {
		return "", ErrUnknownCommand
	}

	var args []string
	if len(part) != 1 {
		if part[1] == "-h" || part[1] == "--help" {
			return cmd.Help(), nil
		}
		args = part
	}

	return exec(cmd, args)
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

func exec(cmd Commander, args []string) (string, error) {
	if err := cmd.Validate(args); err != nil {
		return "", err
	}
	return cmd.Run(args)
}
