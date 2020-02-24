package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/logger"
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
	f func(cmd *cobra.Command, cfg *mergedConfig) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
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
			printVersion(cmd.OutOrStdout())
			return nil

			// Help is processed by cobra.
		}

		// For backward-compatibility.
		var protos []string
		if isRootCommand := cmd.Parent() == nil; isRootCommand {
			protos = args
		}
		// Pass Flags instead of LocalFlags because the config is merged with common and local flags.
		cfg, err := mergeConfig(cmd.Flags(), flags, protos)
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
		Use: "evans [global options ...] <command>",
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) (err error) {
			if cfg.REPL.ColoredOutput {
				ui = cui.NewColored(ui)
			}

			defer func() {
				if err == nil {
					ui.Warn("evans: deprecated usage, please use sub-commands. see `evans -h` for more details.")
				}
			}()

			isCLIMode := (cfg.cli || mode.IsCLIMode(cfg.file))
			if cfg.repl || !isCLIMode {
				return runREPLCommand(cfg, ui)
			}
			invoker, err := mode.NewCallCLIInvoker(ui, cfg.call, cfg.file, cfg.Config.Request.Header, map[string]struct{}{"message": struct{}{}}, "")
			if err != nil {
				return err
			}
			if err := mode.RunAsCLIMode(cfg.Config, invoker); err != nil {
				return errors.Wrap(err, "failed to run CLI mode")
			}

			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().SortFlags = false
	bindFlags(cmd.PersistentFlags(), flags, ui.Writer())
	cmd.SetHelpFunc(usageFunc(ui.Writer(), nil))
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
	f.StringSliceVar(&flags.common.path, "path", nil, "comma-separated proto file paths")
	f.StringSliceVar(&flags.common.proto, "proto", nil, "comma-separated proto file names")
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
	// Hidden is enabled only the root command (see printOptions).
	for _, name := range []string{"repl", "cli", "call", "package", "service", "file"} {
		if err := f.MarkHidden(name); err != nil {
			panic(fmt.Sprintf("failed to mark %s as hidden: %s", name, err))
		}
	}
}

func newCLICommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli",
		Short: "CLI mode",
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			// For backward-compatibility.
			// If the method is specified by passing --call option, use it.
			call := cfg.call
			if call == "" {
				args := cmd.Flags().Args()
				if len(args) == 0 {
					return errors.New("method is required")
				}
				call = args[0]
			}
			invoker, err := mode.NewCallCLIInvoker(ui, call, cfg.file, cfg.Config.Request.Header, map[string]struct{}{"message": struct{}{}}, "")
			if err != nil {
				return err
			}
			if err := mode.RunAsCLIMode(cfg.Config, invoker); err != nil {
				return errors.Wrap(err, "failed to run CLI mode")
			}
			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	f := cmd.Flags()
	initFlagSet(f, ui.Writer())
	cmd.SetHelpFunc(usageFunc(ui.Writer(), nil))
	cmd.AddCommand(
		newCLICallCommand(flags, ui),
		newCLIListCommand(flags, ui),
		newCLIDescribeCommand(flags, ui),
	)
	return cmd
}

func newREPLCommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repl [options ...]",
		Short: "REPL mode",
		RunE: runFunc(flags, func(_ *cobra.Command, cfg *mergedConfig) error {
			return runREPLCommand(cfg, ui)
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	initFlagSet(cmd.Flags(), ui.Writer())
	cmd.SetHelpFunc(usageFunc(ui.Writer(), []string{"package", "service"}))
	return cmd
}

func runREPLCommand(cfg *mergedConfig, ui cui.UI) error {
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

	respFormat := map[string]struct{}{"message": struct{}{}}
	if cfg.verbose {
		for _, k := range []string{"header", "trailer", "status"} {
			respFormat[k] = struct{}{}
		}
	}

	if err := mode.RunAsREPLMode(cfg.Config, ui, cache, respFormat); err != nil {
		return errors.Wrap(err, "failed to run REPL mode")
	}

	// Always call cancel func because it is hope to abort update checking if REPL mode is finished
	// before update checking. If update checking is finished before REPL mode, cancel do nothing.
	cancel()
	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to check application update")
	}
	return nil
}

func initFlagSet(f *pflag.FlagSet, w io.Writer) {
	f.SortFlags = false
	f.SetOutput(w)
}

func printOptions(w io.Writer, cmd *cobra.Command, inheritedFlags []string) {
	_, err := io.WriteString(w, "Options:\n")
	if err != nil {
		logger.Printf("failed to write string: %s", err)
	}
	tw := tabwriter.NewWriter(w, 0, 8, 8, ' ', tabwriter.TabIndent)
	var hasHelp bool
	f := cmd.LocalFlags()
	isRootCommand := cmd.Parent() == nil
	printFlag := func(f *pflag.Flag) {
		if isRootCommand && f.Hidden {
			return
		}
		if f.Name == "help" {
			hasHelp = true
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
		fmt.Fprintf(tw, "        %s\t%s\n", cmd, usage)
	}
	f.VisitAll(printFlag)

	inf := cmd.InheritedFlags()
	for _, name := range inheritedFlags {
		f := inf.Lookup(name)
		if f == nil {
			continue
		}
		printFlag(f)
	}

	// Always show --help text.
	if !hasHelp {
		cmd := "--help, -h"
		usage := `display help text and exit (default "false")`
		fmt.Fprintf(tw, "        %s\t%s\n", cmd, usage)
	}
	tw.Flush()
}

// usage is the generator for usage output.
func usageFunc(out io.Writer, inheritedFlags []string) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, _ []string) {
		rcmd := cmd
		shortUsages := []string{rcmd.Use}
		for {
			rcmd = rcmd.Parent()
			if rcmd == nil {
				break
			}
			u := strings.TrimSuffix(rcmd.Use, " <command>")
			shortUsages = append([]string{u}, shortUsages...)
		}

		printVersion(out)
		fmt.Fprint(out, "\n")
		var buf bytes.Buffer
		printOptions(&buf, cmd, inheritedFlags)
		fmt.Fprintf(out, "Usage: %s\n\n", strings.Join(shortUsages, " "))
		if cmd.Long != "" {
			fmt.Fprint(out, cmd.Long)
			fmt.Fprint(out, "\n\n")
		}
		if cmd.Example != "" {
			fmt.Fprint(out, "Examples:\n")
			fmt.Fprint(out, cmd.Example)
			fmt.Fprint(out, "\n\n")
		}
		fmt.Fprint(out, buf.String())
		fmt.Fprint(out, "\n")

		if len(cmd.Commands()) > 0 {
			fmt.Fprintf(out, "Available Commands:\n")
			w := tabwriter.NewWriter(out, 0, 8, 8, ' ', tabwriter.TabIndent)
			for _, c := range cmd.Commands() {
				// Ignore help command.
				if c.Name() == "help" {
					continue
				}
				cmdAndAliases := append([]string{c.Name()}, c.Aliases...)
				fmt.Fprintf(w, "        %s\t%s\n", strings.Join(cmdAndAliases, ", "), c.Short)
			}
			w.Flush()
			fmt.Fprintf(out, "\n")
		}
	}
}
