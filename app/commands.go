package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/mode"
	"github.com/ktr0731/evans/prompt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

type command struct {
	*cobra.Command

	flags *flags
	ui    cui.UI
}

// registerNewCommands registers sub-commands for new-style interface.
func (c *command) registerNewCommands() {
	c.AddCommand(
		newCLICommand(c.flags, c.ui),
		newREPLCommand(c.flags, c.ui),
	)
}

// runFunc is a common entrypoint for Run func.
func runFunc(
	flags *flags,
	f func(*cobra.Command, *mergedConfig) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := flags.validate(); err != nil {
			return errors.Wrap(err, "invalid flag condition")
		}

		if flags.meta.help {
			printUsage(cmd)
			return nil
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
			printVersion(cmd.OutOrStdout())
			return nil
		case flags.meta.help:
			printUsage(cmd)
			return nil
		}

		// Pass Flags instead of LocalFlags because the config is merged with common and local flags.
		cfg, err := mergeConfig(cmd.Flags(), flags, args)
		if err != nil {
			if err, ok := err.(*config.ValidationError); ok {
				printUsage(cmd)
				return err
			}
			return errors.Wrap(err, "failed to merge command line flags and config files")
		}

		// The entrypoint for the command.
		err = f(cmd, cfg)
		if err == nil {
			return nil
		}
		if _, ok := err.(*config.ValidationError); ok {
			printUsage(cmd)
		}
		return err
	}
}

func newOldCommand(flags *flags, ui cui.UI) *command {
	cmd := &cobra.Command{
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			if cfg.REPL.ColoredOutput {
				ui = cui.NewColored(ui)
			}

			isCLIMode := (cfg.cli || mode.IsCLIMode(cfg.file))
			if cfg.repl || !isCLIMode {
				cache, err := cache.Get()
				if err != nil {
					return errors.Wrap(err, "failed to get the cache content")
				}

				baseCtx, cancel := context.WithCancel(context.Background())
				defer cancel()
				eg, ctx := errgroup.WithContext(baseCtx)
				// Run update checker asynchronously.
				eg.Go(func() error {
					return checkUpdate(ctx, cfg.Config, cache)
				})

				if cfg.Config.Meta.AutoUpdate {
					eg.Go(func() error {
						return processUpdate(ctx, cfg.Config, ui.Writer(), cache, prompt.New())
					})
				} else if err := processUpdate(ctx, cfg.Config, ui.Writer(), cache, prompt.New()); err != nil {
					return errors.Wrap(err, "failed to update Evans")
				}

				if err := mode.RunAsREPLMode(cfg.Config, ui, cache); err != nil {
					return errors.Wrap(err, "failed to run REPL mode")
				}

				// Always call cancel func because it is hope to abort update checking if REPL mode is finished
				// before update checking. If update checking is finished before REPL mode, cancel do nothing.
				cancel()
				if err := eg.Wait(); err != nil {
					return errors.Wrap(err, "failed to check application update")
				}
			} else if err := mode.RunAsCLIMode(cfg.Config, cfg.call, cfg.file, ui); err != nil {
				return errors.Wrap(err, "failed to run CLI mode")
			}

			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	bindFlags(cmd.PersistentFlags(), flags, ui.Writer())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.PersistentFlags().Usage()
	})
	cmd.SetOut(ui.Writer())
	return &command{cmd, flags, ui}
}

func bindFlags(f *pflag.FlagSet, flags *flags, w io.Writer) {
	initFlagSet(f, w)

	f.BoolVar(&flags.mode.repl, "repl", false, "launch Evans as REPL mode")
	f.BoolVar(&flags.mode.cli, "cli", false, "start as CLI mode")

	f.StringVar(&flags.cli.call, "call", "", "call specified RPC by CLI mode")
	f.StringVarP(&flags.cli.file, "file", "f", "", "a script file that will be executed by (used only CLI mode)")

	f.BoolVarP(&flags.repl.silent, "silent", "s", false, "hide redundant output")

	f.StringVar(&flags.common.pkg, "package", "", "default package")
	f.StringVar(&flags.common.service, "service", "", "default service")
	f.StringSliceVar(&flags.common.path, "path", nil, "proto file paths")
	f.StringVar(&flags.common.host, "host", "", "gRPC server host")
	f.StringVarP(&flags.common.port, "port", "p", "50051", "gRPC server port")
	f.Var(
		newStringToStringValue(nil, &flags.common.header),
		"header", "default headers that set to each requests (example: foo=bar)")
	f.BoolVar(&flags.common.web, "web", false, "use gRPC-Web protocol")
	f.BoolVarP(&flags.common.reflection, "reflection", "r", false, "use gRPC reflection")
	f.BoolVarP(&flags.common.tls, "tls", "t", false, "use a secure TLS connection")
	f.StringVar(&flags.common.cacert, "cacert", "", "the CA certificate file for verifying the server")
	f.StringVar(
		&flags.common.cert,
		"cert", "", "the certificate file for mutual TLS auth. it must be provided with --certkey.")
	f.StringVar(
		&flags.common.certKey,
		"certkey", "", "the private key file for mutual TLS auth. it must be provided with --cert.")
	f.StringVar(
		&flags.common.serverName,
		"servername", "", "override the server name used to verify the hostname (ignored if --tls is disabled)")

	f.BoolVarP(&flags.meta.edit, "edit", "e", false, "edit the project config file by using $EDITOR")
	f.BoolVar(&flags.meta.editGlobal, "edit-global", false, "edit the global config file by using $EDITOR")
	f.BoolVar(&flags.meta.verbose, "verbose", false, "verbose output")
	f.BoolVarP(&flags.meta.version, "version", "v", false, "display version and exit")
	f.BoolVarP(&flags.meta.help, "help", "h", false, "display help text and exit")

	// Flags used by old-style only.
	for _, name := range []string{"repl", "cli", "call", "file"} {
		if err := f.MarkHidden(name); err != nil {
			panic(fmt.Sprintf("failed to mark %s as hidden: %s", name, err))
		}
	}
}

func newCLICommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use: "cli",
		RunE: runFunc(flags, func(_ *cobra.Command, cfg *mergedConfig) error {
			if err := mode.RunAsCLIMode(cfg.Config, cfg.call, cfg.file, ui); err != nil {
				return errors.Wrap(err, "failed to run CLI mode")
			}
			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	bindCLIFlags(cmd.LocalFlags(), flags, ui.Writer())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.LocalFlags().Usage()
	})
	return cmd
}

func newREPLCommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use: "repl",
		RunE: runFunc(flags, func(_ *cobra.Command, cfg *mergedConfig) error {
			cache, err := cache.Get()
			if err != nil {
				return errors.Wrap(err, "failed to get the cache content")
			}

			baseCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			eg, ctx := errgroup.WithContext(baseCtx)
			// Run update checker asynchronously.
			eg.Go(func() error {
				return checkUpdate(ctx, cfg.Config, cache)
			})

			if cfg.Config.Meta.AutoUpdate {
				eg.Go(func() error {
					return processUpdate(ctx, cfg.Config, ui.Writer(), cache, prompt.New())
				})
			} else if err := processUpdate(ctx, cfg.Config, ui.Writer(), cache, prompt.New()); err != nil {
				return errors.Wrap(err, "failed to update Evans")
			}

			if err := mode.RunAsREPLMode(cfg.Config, ui, cache); err != nil {
				return errors.Wrap(err, "failed to run REPL mode")
			}

			// Always call cancel func because it is hope to abort update checking if REPL mode is finished
			// before update checking. If update checking is finished before REPL mode, cancel do nothing.
			cancel()
			if err := eg.Wait(); err != nil {
				return errors.Wrap(err, "failed to check application update")
			}
			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	bindREPLFlags(cmd.LocalFlags(), ui.Writer())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.LocalFlags().Usage()
	})
	return cmd
}

func bindCLIFlags(f *pflag.FlagSet, flags *flags, w io.Writer) {
	initFlagSet(f, w)
	f.StringVar(&flags.cli.call, "call", "", "call specified RPC by CLI mode")
	f.StringVarP(&flags.cli.file, "file", "f", "", "a script file that will be executed by (used only CLI mode)")
}

func bindREPLFlags(f *pflag.FlagSet, w io.Writer) {
	initFlagSet(f, w)
}

func initFlagSet(f *pflag.FlagSet, w io.Writer) {
	f.SortFlags = false
	f.SetOutput(w)
	f.Usage = usageFunc(w, f)
}

// usage is the generator for usage output.
func usageFunc(out io.Writer, f *pflag.FlagSet) func() {
	return func() {
		printVersion(out)
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 8, 8, ' ', tabwriter.TabIndent)
		f.VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			cmd := "--" + f.Name
			if f.Shorthand != "" {
				cmd += ", -" + f.Shorthand
			}
			name, _ := pflag.UnquoteUsage(f)
			if name != "" {
				cmd += " " + name
			}
			usage := f.Usage
			if f.DefValue != "" {
				usage += fmt.Sprintf(` (default "%s")`, f.DefValue)
			}
			fmt.Fprintf(w, "        %s\t%s\n", cmd, usage)
		})
		w.Flush()
		fmt.Fprintf(out, usageFormat, meta.AppName, buf.String())
	}
}
