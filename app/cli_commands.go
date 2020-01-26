package app

import (
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

	bindCLIFlags(cmd.LocalFlags(), flags, ui.Writer())

	cmd.SetHelpFunc(usageFunc(ui.Writer()))
	return cmd
}

func newCLIListCommand(flags *flags, ui cui.UI) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [options ...]",
		Aliases: []string{"ls", "l", "show", "s"},
		Short:   "list packages, services, methods or messages",
		RunE: runFunc(flags, func(cmd *cobra.Command, cfg *mergedConfig) error {
			var dsn string
			args := cmd.Flags().Args()
			if len(args) > 0 {
				dsn = args[0]
			}
			invoker := mode.NewListCLIInvoker(ui, dsn)
			if err := mode.RunAsCLIMode(cfg.Config, invoker); err != nil {
				return errors.Wrap(err, "failed to run CLI mode")
			}
			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	bindCLIFlags(cmd.LocalFlags(), flags, ui.Writer())

	cmd.SetHelpFunc(usageFunc(ui.Writer()))
	return cmd
}
