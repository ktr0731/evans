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

	goprompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/adapter/prompt"
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
	if cfg.REPL.ColoredOutput {
		ui = cui.NewColored(ui)
	}

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

	r := newEnv(cfg.REPL, cfg.Server, env, ui, interactor)
	if err := r.start(); err != nil {
		return err
	}
	return nil
}

type repl struct {
	ui           cui.UI
	config       *config.REPL
	serverConfig *config.Server
	env          env.Environment
	// TODO: REPL must not depend to c-bata/go-prompt.
	prompt prompt.Prompt
	cmds   map[string]commander

	// exitCh receives exit signal from executor or
	// goroutine which wrapping Run method.
	exitCh chan struct{}
}

func newEnv(config *config.REPL, serverConfig *config.Server, env env.Environment, ui cui.UI, inputPort port.InputPort) *repl {
	cmds := map[string]commander{
		"call":    &callCommand{inputPort},
		"desc":    &descCommand{inputPort},
		"package": &packageCommand{inputPort},
		"service": &serviceCommand{inputPort},
		"show":    &showCommand{inputPort},
		"header":  &headerCommand{inputPort},
	}

	repl := &repl{
		ui:           ui,
		config:       config,
		serverConfig: serverConfig,
		env:          env,
		cmds:         cmds,
		exitCh:       make(chan struct{}, 2), // for goroutines which manage quit command and CTRL+D
	}

	executor := &executor{repl: repl}
	completer := &completer{cmds: cmds, env: env}

	repl.prompt = prompt.New(
		executor.execute,
		completer.complete,

		goprompt.OptionSuggestionBGColor(goprompt.LightGray),
		goprompt.OptionSuggestionTextColor(goprompt.Black),
		goprompt.OptionDescriptionBGColor(goprompt.White),
		goprompt.OptionDescriptionTextColor(goprompt.Black),

		goprompt.OptionSelectedSuggestionBGColor(goprompt.DarkBlue),
		goprompt.OptionSelectedSuggestionTextColor(goprompt.Black),
		goprompt.OptionSelectedDescriptionBGColor(goprompt.Blue),
		goprompt.OptionSelectedDescriptionTextColor(goprompt.Black),
	)

	repl.prompt.SetPrefix(repl.getPrompt())

	return repl
}

func (r *repl) eval(l string) (io.Reader, error) {
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

func (r *repl) start() error {
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

func (r *repl) help(cmds map[string]commander) string {
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

func (r *repl) getPrompt() string {
	p := fmt.Sprintf("%s:%s> ", r.serverConfig.Host, r.serverConfig.Port)
	if dsn := r.env.DSN(); dsn != "" {
		p = fmt.Sprintf("%s@%s", dsn, p)
	}
	return p
}

func (r *repl) printSplash(p string) {
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
