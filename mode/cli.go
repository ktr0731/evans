package mode

import (
	"context"
	"io"
	"os"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/usecase"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

// DefaultCLIReader is the reader that is read for inputting request values. It is exported for E2E testing.
var DefaultCLIReader io.Reader = os.Stdin

// RunAsCLIMode starts Evans as CLI mode.
func RunAsCLIMode(cfg *config.Config, endpoint, file string, ui cui.UI) error {
	if endpoint == "" {
		return errors.New("method is required")
	}
	in := DefaultCLIReader
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return errors.Wrap(err, "failed to open the script file")
		}
		defer f.Close()
		in = f
	}
	filler := fill.NewSilentFiller(in)
	// TODO: parse package and service from call.

	invoker := func(ctx context.Context) error {
		for k, v := range cfg.Request.Header {
			for _, vv := range v {
				usecase.AddHeader(k, vv)
			}
		}

		err := usecase.CallRPC(ctx, ui.Writer(), endpoint)
		if err != nil {
			return errors.Wrapf(err, "failed to call RPC '%s'", endpoint)
		}
		return nil
	}, nil
}

// IsCLIMode returns whether Evans is launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return file != "" || (!isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()))
}
