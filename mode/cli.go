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

type CLIInvoker func(context.Context) error

func NewCallCLIInvoker(ui cui.UI, rpcName, filePath string, headers config.Header) (CLIInvoker, error) {
	if rpcName == "" {
		return nil, errors.New("method is required")
	}
	// TODO: parse package and service from call.
	return func(ctx context.Context) error {
		in := DefaultCLIReader
		if filePath != "" {
			f, err := os.Open(filePath)
			if err != nil {
				return errors.Wrap(err, "failed to open the script file")
			}
			defer f.Close()
			in = f
		}
		filler := fill.NewSilentFiller(in)
		usecase.InjectPartially(usecase.Dependencies{Filler: filler})

		for k, v := range headers {
			for _, vv := range v {
				usecase.AddHeader(k, vv)
			}
		}

		err := usecase.CallRPC(ctx, ui.Writer(), rpcName)
		if err != nil {
			return errors.Wrapf(err, "failed to call RPC '%s'", rpcName)
		}
		return nil
	}, nil
}

// RunAsCLIMode starts Evans as CLI mode.
func RunAsCLIMode(cfg *config.Config, invoker CLIInvoker) error {
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

	usecase.InjectPartially(
		usecase.Dependencies{
			Spec:              spec,
			GRPCClient:        gRPCClient,
			ResponsePresenter: json.NewPresenter(),
			ResourcePresenter: json.NewPresenter(),
		},
	)

	// If the spec has only one package, mark it as the default package.
	if cfg.Default.Package == "" && len(spec.PackageNames()) == 1 {
		cfg.Default.Package = spec.PackageNames()[0]
	}
	if err := usecase.UsePackage(cfg.Default.Package); err != nil {
		return errors.Wrapf(err, "failed to set '%s' as the default package", cfg.Default.Package)
	}

	// If the spec has only one service, mark it as the default service.
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
		return errors.Wrapf(err, "failed to set '%s' as the default service", cfg.Default.Service)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return invoker(ctx)
}

// IsCLIMode returns whether Evans is launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return file != "" || (!isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()))
}
