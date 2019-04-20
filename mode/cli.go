package mode

import (
	"context"
	"io"
	"os"

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
func RunAsCLIMode(cfg *config.Config, call, file string, ui cui.UI) error {
	if call == "" {
		return errors.New("flag --call must not be empty")
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

	var injectResult error
	gRPCClient, err := newGRPCClient(cfg)
	if err != nil {
		injectResult = multierror.Append(injectResult, err)
	} else {
		defer gRPCClient.Close(context.Background()) // TODO: ctx
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
		fill.NewSilentFiller(in),
		gRPCClient,
		json.NewPresenter(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: signal handling
	if cfg.Default.Package == "" && len(spec.PackageNames()) == 1 {
		cfg.Default.Package = spec.PackageNames()[0]
	}
	if err := usecase.UsePackage(cfg.Default.Package); err != nil {
		return errors.Wrapf(err, "failed to set '%s' as the default package", cfg.Default.Package)
	}

	if cfg.Default.Service == "" {
		svcNames, err := spec.ServiceNames(cfg.Default.Package)
		if err != nil {
			return errors.Wrapf(err, "failed to list services belong to package '%s'", cfg.Default.Package)
		}
		if len(svcNames) == 1 {
			cfg.Default.Service = svcNames[0]
		}
	}

	if err := usecase.UseService(cfg.Default.Service); err != nil {
		return errors.Wrapf(err, "failed to est '%s' as the default service", cfg.Default.Service)
	}

	for k, v := range cfg.Request.Header {
		for _, vv := range v {
			usecase.AddHeader(k, vv)
		}
	}

	err = usecase.CallRPC(ctx, ui.Writer(), call)
	if err != nil {
		return errors.Wrapf(err, "failed to call RPC '%s'", call)
	}

	return nil
}

// IsCLIMode returns whether Evans is launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return file != "" || (!isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()))
}
