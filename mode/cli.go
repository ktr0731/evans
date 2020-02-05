package mode

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/present/json"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/go-multierror"
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

	// Common dependencies.

	var injectResult error
	gRPCClient, err := newGRPCClient(cfg)
	if err != nil {
		injectResult = multierror.Append(injectResult, err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			gRPCClient.Close(ctx)
		}()
	}

	spec, err := newSpec(cfg, gRPCClient)
	if err != nil {
		injectResult = multierror.Append(injectResult, err)
	}

	if injectResult != nil {
		return injectResult
	}

	usecase.Inject(
		spec,
		filler,
		gRPCClient,
		json.NewPresenter(),
		json.NewPresenter(),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := setDefault(cfg); err != nil {
		return err
	}

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
	}

	return invoker(ctx)
}

// IsCLIMode returns whether Evans is launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return file != "" || (!isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()))
}
