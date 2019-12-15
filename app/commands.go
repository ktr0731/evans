package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/logger"
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
	c.AddCommand(newCLICommand(c.flags, c.ui))
}

func newOldCommand(flags *flags, ui cui.UI) *command {
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
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
				printVersion(ui.Writer())
				return nil
			case flags.meta.help:
				printUsage(cmd)
				return nil
			}

			cfg, err := mergeConfig(cmd.PersistentFlags(), flags, args)
			if err != nil {
				if err, ok := err.(*config.ValidationError); ok {
					return err
				}
				return errors.Wrap(err, "failed to merge command line flags and config files")
			}

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
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	bindFlags(cmd.PersistentFlags(), flags, ui.Writer())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.PersistentFlags().Usage()
	})
	return &command{cmd, flags, ui}
}

func bindFlags(f *pflag.FlagSet, flags *flags, w io.Writer) {
	f.SortFlags = false
	f.SetOutput(w)

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
		f.MarkHidden(name)
	}

	f.Usage = func() {
		out := w
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

func newCLICommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use: "cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			// if err := flags.validate(); err != nil {
			// 	return errors.Wrap(err, "invalid flag condition")
			// }
			//
			// if flags.meta.verbose {
			// 	logger.SetOutput(os.Stderr)
			// }
			//
			// switch {
			// case flags.meta.edit:
			// 	if err := config.Edit(); err != nil {
			// 		return errors.Wrap(err, "failed to edit the project local config file")
			// 	}
			// 	return nil
			// case flags.meta.editGlobal:
			// 	if err := config.EditGlobal(); err != nil {
			// 		return errors.Wrap(err, "failed to edit the global config file")
			// 	}
			// 	return nil
			// case flags.meta.version:
			// 	printVersion(ui.Writer())
			// 	return nil
			// case flags.meta.help:
			// 	printUsage(cmd)
			// 	return nil
			// }
			//
			// cfg, err := mergeConfig(cmd.PersistentFlags(), flags, args)
			// if err != nil {
			// 	if err, ok := err.(*config.ValidationError); ok {
			// 		return err
			// 	}
			// 	return errors.Wrap(err, "failed to merge command line flags and config files")
			// }
			//
			// if cfg.REPL.ColoredOutput {
			// 	ui = cui.NewColored(ui)
			// }
			//
			// isCLIMode := (cfg.cli || mode.IsCLIMode(cfg.file))
			// if cfg.repl || !isCLIMode {
			// 	cache, err := cache.Get()
			// 	if err != nil {
			// 		return errors.Wrap(err, "failed to get the cache content")
			// 	}
			//
			// 	baseCtx, cancel := context.WithCancel(context.Background())
			// 	defer cancel()
			// 	eg, ctx := errgroup.WithContext(baseCtx)
			// 	// Run update checker asynchronously.
			// 	eg.Go(func() error {
			// 		return checkUpdate(ctx, cfg.Config, cache)
			// 	})
			//
			// 	if cfg.Config.Meta.AutoUpdate {
			// 		eg.Go(func() error {
			// 			return processUpdate(ctx, cfg.Config, ui.Writer(), cache, prompt.New())
			// 		})
			// 	} else if err := processUpdate(ctx, cfg.Config, ui.Writer(), cache, prompt.New()); err != nil {
			// 		return errors.Wrap(err, "failed to update Evans")
			// 	}
			//
			// 	if err := mode.RunAsREPLMode(cfg.Config, ui, cache); err != nil {
			// 		return errors.Wrap(err, "failed to run REPL mode")
			// 	}
			//
			// 	// Always call cancel func because it is hope to abort update checking if REPL mode is finished
			// 	// before update checking. If update checking is finished before REPL mode, cancel do nothing.
			// 	cancel()
			// 	if err := eg.Wait(); err != nil {
			// 		return errors.Wrap(err, "failed to check application update")
			// 	}
			// } else if err := mode.RunAsCLIMode(cfg.Config, cfg.call, cfg.file, ui); err != nil {
			// 	return errors.Wrap(err, "failed to run CLI mode")
			// }

			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	bindCLIFlags(cmd.PersistentFlags(), flags, ui.Writer())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.PersistentFlags().Usage()
	})
	return cmd
}

func bindCLIFlags(f *pflag.FlagSet, flags *flags, w io.Writer) {
	f.SortFlags = false
	f.SetOutput(w)

	f.StringVar(&flags.cli.call, "call", "", "call specified RPC by CLI mode")
	f.StringVarP(&flags.cli.file, "file", "f", "", "a script file that will be executed by (used only CLI mode)")

	f.Usage = func() {
		out := w
		printVersion(out)
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 8, 8, ' ', tabwriter.TabIndent)
		f.VisitAll(func(f *pflag.Flag) {
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
