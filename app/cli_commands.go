package app

import (
	"strings"

	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/mode"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newCLICallCommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "call [options ...] <method>",
		Aliases: []string{"c"},
		Short:   "call a RPC",
		Long:    `call invokes a RPC based on the passed method name.`,
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			args := cmd.Flags().Args()
			if len(args) == 0 {
				return errors.New("method is required")
			}
			invoker, err := mode.NewCallCLIInvoker(ui, args[0], cfg.file, cfg.Config.Request.Header)
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

	bindCLICallFlags(cmd.Flags(), flags, ui.Writer())

	cmd.SetHelpFunc(usageFunc(ui.Writer()))
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

	cmd.SetHelpFunc(usageFunc(ui.Writer()))
	return cmd
}
