package mode

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/present"
	"github.com/ktr0731/evans/present/json"
	"github.com/ktr0731/evans/present/name"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/go-multierror"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

// DefaultCLIReader is the reader that is read for inputting request values. It is exported for E2E testing.
var DefaultCLIReader io.Reader = os.Stdin

// CLIInvoker represents an invokable function for CLI mode.
type CLIInvoker func(context.Context) error

// NewCallCLIInvoker returns an CLIInvoker implementation for calling RPCs.
// If filePath is empty, the invoker tries to read input from stdin.
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

func NewListCLIInvoker(ui cui.UI, fqn, format string) CLIInvoker {
	return func(context.Context) error {
		var presenter present.Presenter
		switch format {
		case "name":
			presenter = name.NewPresenter()
		case "json":
			presenter = json.NewPresenter()
		default:
			presenter = name.NewPresenter()
		}
		usecase.InjectPartially(usecase.Dependencies{ResourcePresenter: presenter})

		pkgs := make(map[string]struct{})
		for _, p := range usecase.ListPackages() {
			pkgs[p] = struct{}{}
		}
		// singlePkg := len(pkgs) == 1
		sp := strings.Split(fqn, ".")

		out, err := func() (string, error) {
			switch {
			case len(sp) == 1 && sp[0] == "": // Unspecified.
				var svcs []string
				for _, pkg := range usecase.ListPackages() {
					if err := usecase.UsePackage(pkg); err != nil {
						return "", errors.Wrapf(err, "failed to use package '%s'", pkg)
					}
					svc, err := usecase.FormatServices(
						&usecase.FormatServicesParams{FullyQualifiedName: true},
					)
					if err != nil {
						return "", errors.Wrap(err, "failed to list services")
					}
					svcs = append(svcs, svc)
				}
				return strings.Join(svcs, "\n"), nil
			// case len(sp) == 1 && singlePkg: // Package name or service name.
			// 	if _, ok := pkgs[sp[0]]; ok {
			// 		svc, err := usecase.FormatServices()
			// 		if err != nil {
			// 			return "", errors.Wrap(err, "failed to format services")
			// 		}
			// 		return svc, nil
			// 	}
			//
			// 	// Check service name.
			// 	svcs, err := usecase.ListServices(false)
			// 	if err != nil {
			// 		return "", errors.Wrap(err, "failed to list services")
			// 	}
			// 	for _, s := range svcs {
			// 		if sp[0] == s {
			// 			rpcs, err := usecase.FormatRPCs()
			// 			if err != nil {
			// 				return "", errors.Wrap(err, "failed to format methods")
			// 			}
			// 			return rpcs, nil
			// 		}
			// 	}
			//
			// 	// Check message name.
			// 	panic("TODO")
			//
			case len(sp) == 1: // Package name.
				svc, err := usecase.FormatServices(nil)
				if err != nil {
					return "", errors.Wrap(err, "failed to list services")
				}
				return svc, nil
			}
			return "", errors.Errorf("unknown fully-qualified name '%s'", fqn)
		}()
		if err != nil {
			return err
		}
		ui.Output(out)
		return nil
	}
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

	if err := setDefault(cfg, spec); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return invoker(ctx)
}

// IsCLIMode returns whether Evans is launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return file != "" || (!isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()))
}
