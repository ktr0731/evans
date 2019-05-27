// Package app provides the entrypoint for Evans.
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/mode"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

// App is the root component for running the application.
type App struct {
	cui cui.UI

	flagSet *pflag.FlagSet

	cfg *mergedConfig
}

// New instantiates a new App instance. If cui is nil, the default UI will be used.
// Note that cui is also used for the REPL UI if the mode is REPL mode.
func New(cui cui.UI) *App {
	return &App{
		cui: cui,
	}
}

// Run starts the application. The return value means the exit code.
func (a *App) Run(args []string) int {
	err := a.run(args)
	if err == nil {
		return 0
	}

	switch err := err.(type) {
	case *config.ValidationError:
		a.printUsage()
		a.cui.Error(fmt.Sprintf("evans: %s", err.Err))
	default:
		a.cui.Error(fmt.Sprintf("evans: %s", err))
	}
	return 1
}

func (a *App) run(args []string) error {
	flags, err := a.parseFlags(args)
	if err != nil {
		return err
	}
	if err := flags.validate(); err != nil {
		return errors.Wrap(err, "invalid flag condition")
	}

	if flags.meta.verbose {
		logger.SetOutput(os.Stderr)
	}

	switch {
	case flags.meta.edit:
		if err := config.Edit(); err != nil {
			return errors.Wrap(err, "failed to edit the project local config file")
		}
		return nil
	case flags.meta.editGlobal:
		if err := config.EditGlobal(); err != nil {
			return errors.Wrap(err, "failed to edit the global config file")
		}
		return nil
	case flags.meta.version:
		a.printVersion()
		return nil
	case flags.meta.help:
		a.printUsage()
		return nil
	}

	cfg, err := mergeConfig(a.flagSet, flags)
	if err != nil {
		if err, ok := err.(*config.ValidationError); ok {
			return err
		}
		return errors.Wrap(err, "failed to merge command line flags and config files")
	}
	a.cfg = cfg

	if cfg.REPL.ColoredOutput {
		a.cui = cui.NewColored(a.cui)
	}

	isCLIMode := (cfg.cli || mode.IsCLIMode(cfg.file))
	if cfg.repl || !isCLIMode {
		baseCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		eg, ctx := errgroup.WithContext(baseCtx)
		// Run update checker asynchronously.
		eg.Go(func() error {
			return checkUpdate(ctx, a.cfg.Config)
		})

		if a.cfg.Config.Meta.AutoUpdate {
			eg.Go(func() error {
				return processUpdate(ctx, a.cfg.Config, a.cui.Writer())
			})
		} else if err := processUpdate(ctx, a.cfg.Config, a.cui.Writer()); err != nil {
			return errors.Wrap(err, "failed to update Evans")
		}

		if err := mode.RunAsREPLMode(a.cfg.Config, a.cui); err != nil {
			return errors.Wrap(err, "failed to run REPL mode")
		}

		// Always call cancel func because it is hope to abort update checking if REPL mode is finished
		// before update checking. If update checking is finished before REPL mode, cancel do nothing.
		cancel()
		if err := eg.Wait(); err != nil {
			return errors.Wrap(err, "failed to check application update")
		}
	} else if err := mode.RunAsCLIMode(a.cfg.Config, a.cfg.call, a.cfg.file, a.cui); err != nil {
		return errors.Wrap(err, "failed to run CLI mode")
	}

	return nil
}

// printUsage shows the command usage text to cui.Writer and exit. Do not call it before calling parseFlags.
func (a *App) printUsage() {
	a.flagSet.Usage()
}

// printVersion shows the version of the command and exit.
func (a *App) printVersion() {
	a.cui.Output(fmt.Sprintf("%s %s", meta.AppName, meta.Version.String()))
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

func mergeConfig(fs *pflag.FlagSet, flags *flags) (*mergedConfig, error) {
	cfg, err := config.Get(fs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}
	protos := fs.Args()
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
