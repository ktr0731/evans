// Package app provides the entrypoint for Evans.
package app

import (
	"fmt"
	"io"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/usecase"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// App is the root component for running the application.
type App struct {
	cui cui.UI
	cmd *command
}

// New instantiates a new App instance. ui must not be a nil.
// Note that cui is also used for the REPL UI if the mode is REPL mode.
func New(ui cui.UI) *App {
	var flags flags
	cmd := newOldCommand(&flags, ui)
	return &App{
		cui: ui,
		cmd: cmd,
	}
}

// Run starts the application. The return value means the exit code.
func (a *App) Run(args []string) int {
	// Currently, Evans is migrating to new-style command-line interface.
	// So, there are both of old-style and new-style command-line interfaces in this version.

	a.cmd.SetArgs(args)
	for _, r := range args {
		// Hack.
		switch r {
		case "cli", "repl": // Sub commands for new-style interface.
			// If an arg named "cli" or "repl" is passed, it is regarded as a sub-command of new-style.
			a.cmd.registerNewCommands()
			a.cmd.RunE = nil
		case "-h", "--help":
			// If the help flags is passed, call registerNewCommands for display sub-command helps.
			a.cmd.registerNewCommands()
		}
	}
	err := a.cmd.Execute()
	if err == nil {
		return 0
	}

	var e interface {
		Code() usecase.ErrorCode
		Message() string
	}
	if errors.As(err, &e) {
		a.cui.Error(
			fmt.Sprintf("evans: code = %s, number = %d, message = %q", e.Code().String(), e.Code(), e.Message()),
		)
		return 1
	}

	a.cui.Error(fmt.Sprintf("evans: %s", err))
	return 1
}

// printUsage shows the command usage text to cui.Writer and exit. Do not call it before calling parseFlags.
func printUsage(cmd interface{ Help() error }) {
	_ = cmd.Help() // Help never return errors.
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "%s %s\n", meta.AppName, meta.Version.String())
}

// mergedConfig represents the conclusive config. Common config items are stored to *config.Config.
// Flags that can be specified by command line only are represented as fields.
type mergedConfig struct {
	*config.Config

	// The RPC name that want to call.
	call string
	// The input that is used for CLI mode. Empty if the input is stdin.
	file string
	// Explicit using CLI mode.
	cli bool

	// Explicit using REPL mode.
	repl bool

	// Verbose output.
	verbose bool
}

func mergeConfig(fs *pflag.FlagSet, flags *flags, protos []string) (*mergedConfig, error) {
	cfg, err := config.Get(fs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}
	cfg.Default.ProtoFile = append(cfg.Default.ProtoFile, protos...)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &mergedConfig{
		Config:  cfg,
		call:    flags.cli.call,
		file:    flags.cli.file,
		repl:    flags.mode.repl,
		cli:     flags.mode.cli,
		verbose: flags.meta.verbose,
	}, nil
}
