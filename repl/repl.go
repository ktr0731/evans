// Package repl provides a REPL environment for REPL mode.
package repl

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/prompt"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/go-shellstring"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// REPL represents a REPL mechanism.
type REPL struct {
	cfg       *config.REPL
	serverCfg *config.Server
	prompt    prompt.Prompt
	ui        *cui.UI

	cmds    map[string]commander
	aliases map[string]string
}

var commands = map[string]commander{
	"call":    &callCommand{},
	"service": &serviceCommand{},
	"header":  &headerCommand{},
	"package": &packageCommand{},
	"show":    &showCommand{},
	"exit":    &exitCommand{},

	// Depends to Protocol Buffers.
	"desc": &descCommand{},
}

// New instantiates a new REPL instance. New always calls p.SetPrefix for display the server addr.
// New may return an error if some of passed arguments are invalid.
func New(cfg *config.Config, p prompt.Prompt, ui *cui.UI, pkgName, svcName string) (*REPL, error) {
	cmds := commands
	// Each value must be a key of cmds.
	aliases := map[string]string{
		"quit": "exit",
	}

	p.SetCompleter(&completer{cmds: cmds})
	p.SetCommandHistory([]string{"foo", "bar"})

	var result error
	if pkgName != "" {
		if err := usecase.UsePackage(pkgName); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if svcName != "" {
		if err := usecase.UseService(svcName); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if result != nil {
		return nil, errors.Wrap(result, "failed to instantiate a new REPL")
	}
	r := &REPL{
		cfg:       cfg.REPL,
		serverCfg: cfg.Server,
		prompt:    p,
		ui:        ui,
		cmds:      cmds,
		aliases:   aliases,
	}

	return r, nil
}

// Run starts the read-eval-print-loop.
func (r *REPL) Run(ctx context.Context) error {
	defer r.cleanup(ctx)
	if !r.cfg.Silent {
		r.printSplash(r.cfg.SplashTextPath)
		defer r.ui.Info("Good Bye :)")
	}

	for {
		r.prompt.SetPrefix(r.makePrefix())

		in, err := r.prompt.Input()
		if err == io.EOF {
			return nil
		}

		in = strings.TrimSpace(in)
		if in == "" {
			continue
		}

		part, err := shellstring.Parse(in)
		if err != nil {
			return nil
		}

		err = r.runCommand(part[0], part[1:])
		if err == io.EOF {
			return nil
		}
		if err != nil {
			r.ui.Error(fmt.Sprintf("command %s: %s", part[0], err))
		}

		r.ui.Output("") // Break line.
	}
}

func (r *REPL) cleanup(ctx context.Context) {
}

func (r *REPL) runCommand(cmdName string, args []string) error {
	if cmdName == "help" {
		fmt.Fprintln(r.ui.Writer, r.helpText())
		return nil
	}

	cmd, ok := r.cmds[cmdName]
	if !ok {
		// Check whether cmdName is an alias for a command.
		if alias, ok := r.aliases[cmdName]; ok {
			cmd = r.cmds[alias]
		} else {
			return errors.New("unknown command")
		}
	}

	if len(args) != 0 {
		if args[0] == "-h" || args[0] == "--help" {
			fmt.Fprintln(r.ui.Writer, cmd.Help())
			return nil
		}
	}

	if err := cmd.Validate(args); err != nil {
		return err
	}
	if err := cmd.Run(r.ui.Writer, args); err != nil {
		return err
	}

	return nil
}

func (r *REPL) printSplash(p string) {
	if p == "" {
		r.ui.Output(defaultSplashText)
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
			r.ui.Output(string(b))
		}
	}
}

func (r *REPL) makePrefix() string {
	p := fmt.Sprintf("%s:%s> ", r.serverCfg.Host, r.serverCfg.Port)
	dsn := usecase.GetDomainSourceName()
	if dsn != "" {
		p = fmt.Sprintf("%s@%s", dsn, p)
	}
	return p
}

func (r *REPL) helpText() string {
	var maxLen int
	// slice of [name, synopsis]
	text := make([][]string, 0, len(r.cmds))

	for name, cmd := range r.cmds {
		text = append(text, []string{name, cmd.Synopsis()})
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	sort.Slice(text, func(i, j int) bool {
		return text[i][0] < text[j][0]
	})

	var cmdText string
	for _, t := range text {
		cmdText += fmt.Sprintf("  %-"+strconv.Itoa(maxLen)+"s    %s\n", t[0], t[1])
	}
	msg := fmt.Sprintf(`
Available commands:
%s
Show more details:
  <command> --help
`, cmdText)
	return strings.TrimRight(msg, "\n")
}
