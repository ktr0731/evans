package app

import (
	"strings"

	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/mode"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newCLICallCommand(flags *flags, ui cui.UI) *cobra.Command {
	var (
		out          string
		enrich       bool
		emitDefaults bool
	)
	cmd := &cobra.Command{
		Use:     "call [options ...] <method>",
		Aliases: []string{"c"},
		Short:   "call a method",
		Long:    `call invokes a method based on the passed method name.`,
		Example: strings.Join([]string{
			"        $ echo '{}' | evans -r cli call api.Service.Unary # call Unary method with an empty message",
			"        $ evans -r cli call -f in.json api.Service.Unary  # call Unary method with an input file",
			"",
			"        $ evans -r cli call -f in.json --enrich --output json api.Service.Unary # enrich output with JSON format",
		}, "\n"),
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			if cfg.REPL.ColoredOutput {
				ui = cui.NewColored(ui)
			}

			args := cmd.Flags().Args()
			if len(args) == 0 {
				return errors.New("method is required")
			}
			invoker, err := mode.NewCallCLIInvoker(ui, args[0], &mode.CallCLIInvokerOption{
				Headers:      cfg.Config.Request.Header,
				Enrich:       enrich,
				EmitDefaults: emitDefaults,
				FilePath:     cfg.file,
				FormatType:   out,
			})
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
	f.BoolVar(&enrich, "enrich", false, `enrich response output includes header, message, trailer and status`)
	f.BoolVar(&emitDefaults, "emit-defaults", false, `render fields with default values`)
	f.StringVarP(&out, "output", "o", "curl", `output format. one of "json" or "curl". "curl" is a curl-like format.`)

	cmd.SetHelpFunc(usageFunc(ui.Writer(), []string{"file"}))
	return cmd
}

func newCLIListCommand(flags *flags, ui cui.UI) *cobra.Command {
	var (
		out string
	)
	cmd := &cobra.Command{
		Use:     "list [options ...] [fully-qualified service/method name]",
		Aliases: []string{"ls", "show"},
		Short:   "list services or methods",
		Long: `list provides listing feature against to gRPC services or methods belong to a service.
If a fully-qualified service name (in the form of <package name>.<service name>),
list lists method names belong to the service. If not, list lists all services.`,
		Example: strings.Join([]string{
			"        $ evans -r cli list             # list all services",
			"        $ evans -r cli list -o json     # list all services with JSON format",
			`        $ evans -r cli list api.Service # list all methods belong to service "api.Service"`,
		}, "\n"),
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			if cfg.REPL.ColoredOutput {
				ui = cui.NewColored(ui)
			}

			var dsn string
			args := cmd.Flags().Args()
			if len(args) > 0 {
				dsn = args[0]
			}
			invoker := mode.NewListCLIInvoker(ui, dsn, out)
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
	f.StringVarP(&out, "output", "o", "name", `output format. one of "json" or "name".`)

	cmd.SetHelpFunc(usageFunc(ui.Writer(), nil))
	return cmd
}

func newCLIDescribeCommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "desc [options ...] [symbol]",
		Aliases: []string{"describe"},
		Short:   "describe the descriptor of a symbol",
		Long: `desc shows the descriptor of the given symbol.
The symbol should be a fully-qualified name. If no symbol is passed, desc shows all descriptors of the loaded services.`,
		Example: strings.Join([]string{
			"        $ evans -r cli desc             # describe the descriptors of the loaded services",
			`        $ evans -r cli desc api.Service # describe the service descriptor of "api.Service"`,
			`        $ evans -r cli desc api.Request # describe the message descriptor of "api.Request"`,
		}, "\n"),
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			if cfg.REPL.ColoredOutput {
				ui = cui.NewColored(ui)
			}

			var fqn string
			args := cmd.Flags().Args()
			if len(args) > 0 {
				fqn = args[0]
			}
			invoker := mode.NewDescribeCLIInvoker(ui, fqn)
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
	return cmd
}
