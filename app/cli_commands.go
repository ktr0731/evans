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
		RunE: runFunc(flags, func(_ *cobra.Command, cfg *mergedConfig, args []string) error {
			if len(args) == 0 {
				return errors.New("method is required")
			}
			if err := mode.RunAsCLIMode(cfg.Config, args[0], cfg.file, ui, "call"); err != nil {
				return errors.Wrap(err, "failed to run CLI mode")
			}
			return nil
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	f := cmd.LocalFlags()
	initFlagSet(f, ui.Writer())
	f.BoolVarP(&flags.meta.help, "help", "h", false, "display help text and exit")

	cmd.SetHelpFunc(usageFunc(ui.Writer()))
	return cmd
}
