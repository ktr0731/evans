// Package app provides the entrypoint for Evans.
package app

import (
	"fmt"
	"io"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/meta"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// App is the root component for running the application.
type App struct {
	cui cui.UI
	cmd *cobra.Command
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
	a.cmd.SetArgs(args)
	err := a.cmd.Execute()
	if err == nil {
		return 0
	}

	switch err := err.(type) {
	case *config.ValidationError:
		printUsage(a.cmd)
		a.cui.Error(fmt.Sprintf("evans: %s", err.Err))
	default:
		a.cui.Error(fmt.Sprintf("evans: %s", err))
	}
	return 1
}

// printUsage shows the command usage text to cui.Writer and exit. Do not call it before calling parseFlags.
func printUsage(cmd *cobra.Command) {
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
}

func mergeConfig(fs *pflag.FlagSet, flags *flags, args []string) (*mergedConfig, error) {
	cfg, err := config.Get(fs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}
	protos := args
	cfg.Default.ProtoFile = append(cfg.Default.ProtoFile, protos...)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &mergedConfig{
		Config: cfg,
		call:   flags.cli.call,
		file:   flags.cli.file,
		repl:   flags.mode.repl,
		cli:    flags.mode.cli,
	}, nil
}
