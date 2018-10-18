package repl

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/adapter/gateway"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
	shellstring "github.com/ktr0731/go-shellstring"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var (
	ErrUnknownCommand   = errors.New("unknown command")
	ErrArgumentRequired = errors.New("argument required")
	ErrUnknownTarget    = errors.New("unknown target")
)

// TODO: define cli mode scoped config

var (
	// DefaultReader is used for e2e testing.
	DefaultReader io.Reader = os.Stdin
)

func Run(cfg *config.Config, ui cui.UI) error {
	in := DefaultReader

	p, err := di.NewREPLInteractorParams(cfg, in)
	if err != nil {
		return err
	}
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()
	defer p.Cleanup(closeCtx)

	interactor := usecase.NewInteractor(p)

	env, err := di.Env(cfg)
	if err != nil {
		return err
	}

	r := New(cfg.REPL, env, ui, interactor)
	if err := r.Start(); err != nil {
		return err
	}
	return nil
}

type REPL struct {
	ui     cui.UI
	config *config.REPL
	env    env.Environment
	prompt gateway.Prompter
	cmds   map[string]Commander

	// exitCh receives exit signal from executor or
	// goroutine which wrapping Run method.
	exitCh chan struct{}
}

func New(config *config.REPL, env env.Environment, ui cui.UI, inputPort port.InputPort) *REPL {
	cmds := map[string]Commander{
		"call":    &callCommand{inputPort},
		"desc":    &descCommand{inputPort},
		"package": &packageCommand{inputPort},
		"service": &serviceCommand{inputPort},
		"show":    &showCommand{inputPort},
		"header":  &headerCommand{inputPort},
	}

	repl := &REPL{
		ui:     ui,
		config: config,
		env:    env,
		cmds:   cmds,
		exitCh: make(chan struct{}, 2), // for goroutines which manage quit command and CTRL+D
	}

	executor := &executor{repl: repl}
	completer := &completer{cmds: cmds, env: env}

	repl.prompt = gateway.NewRealPrompter(
		executor.execute,
		completer.complete,

		prompt.OptionSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.White),
		prompt.OptionDescriptionTextColor(prompt.Black),

		prompt.OptionSelectedSuggestionBGColor(prompt.DarkBlue),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionSelectedDescriptionBGColor(prompt.Blue),
		prompt.OptionSelectedDescriptionTextColor(prompt.Black),
	)

	repl.prompt.SetPrefix(repl.getPrompt())

	return repl
}

func (r *REPL) eval(l string) (io.Reader, error) {
	// trim quote
	// e.g. key='foo' is interpreted to `foo`
	//      key='foo bar' is `foo bar`
	//      key='"foo bar"' is `"foo bar"`
	//      key=foo bar is also `foo bar`
	part, err := shellstring.Parse(l)
	if err != nil {
		return nil, err
	}

	if part[0] == "help" {
		return strings.NewReader(r.help(r.cmds) + "\n"), nil
	}

	cmd, ok := r.cmds[part[0]]
	if !ok {
		return nil, ErrUnknownCommand
	}

	var args []string
	if len(part) != 1 {
		if part[1] == "-h" || part[1] == "--help" {
			return strings.NewReader(cmd.Help() + "\n"), nil
		}
		args = part[1:]
	}

	if err := cmd.Validate(args); err != nil {
		return nil, err
	}
	return cmd.Run(args)
}

func (r *REPL) Start() error {
	if r.config.ShowSplashText {
		r.printSplash(r.config.SplashTextPath)
		defer r.ui.InfoPrintln("Good Bye :)")
	}

	go func() {
		r.prompt.Run()
		r.exitCh <- struct{}{}
	}()

	// wait until input exit or quit command or
	// above goroutine finished.
	<-r.exitCh

	return nil
}

func (r *REPL) help(cmds map[string]Commander) string {
	var maxLen int
	// slice of [name, synopsis]
	text := make([][]string, len(cmds))
	for name, cmd := range cmds {
		text = append(text, []string{name, cmd.Synopsis()})
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	var cmdText string
	for name, cmd := range cmds {
		cmdText += fmt.Sprintf("  %-"+strconv.Itoa(maxLen)+"s    %s\n", name, cmd.Synopsis())
	}
	msg := fmt.Sprintf(`
Available commands:
%s
Show more details:
  <command> --help
`, cmdText)
	return strings.TrimRight(msg, "\n")
}

func (r *REPL) getPrompt() string {
	p := fmt.Sprintf("%s:%s> ", r.config.Server.Host, r.config.Server.Port)
	if dsn := r.env.DSN(); dsn != "" {
		p = fmt.Sprintf("%s@%s", dsn, p)
	}
	return p
}

const defaultSplashText = `
  ______
 |  ____|
 | |__    __   __   __ _   _ __    ___
 |  __|   \ \ / /  / _. | | '_ \  / __|
 | |____   \ V /  | (_| | | | | | \__ \
 |______|   \_/    \__,_| |_| |_| |___/

 more expressive universal gRPC client

`

func (r *REPL) printSplash(p string) {
	if p == "" {
		r.ui.Println(defaultSplashText)
		return
	}

	var abs string
	if strings.HasPrefix(p, "~/") {
		home, err := homedir.Dir()
		if err == nil {
			abs = filepath.Join(home, strings.TrimPrefix(p, "~/"))
		}
	} else {
		abs, _ = filepath.Abs(p)
	}
	if abs == "" {
		return
	}

	_, err := os.Stat(abs)
	if !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(abs)
		if err == nil {
			r.ui.Println(string(b))
		}
	}
}
