package mode

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/format"
	"github.com/ktr0731/evans/format/curl"
	fmtjson "github.com/ktr0731/evans/format/json"
	"github.com/ktr0731/evans/idl"
	"github.com/ktr0731/evans/idl/proto"
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
func NewCallCLIInvoker(ui cui.UI, methodName, filePath string, headers config.Header, enrich bool, formatType string) (CLIInvoker, error) {
	if methodName == "" {
		return nil, errors.New("method is required")
	}
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
		var rfi format.ResponseFormatterInterface
		switch formatType {
		case "curl":
			rfi = curl.NewResponseFormatter(ui.Writer())
		case "json":
			rfi = fmtjson.NewResponseFormatter(ui.Writer())
		default:
			rfi = curl.NewResponseFormatter(ui.Writer())
		}
		usecase.InjectPartially(usecase.Dependencies{
			ResponseFormatter: format.NewResponseFormatter(rfi, enrich),
			Filler:            filler,
		})

		for k, v := range headers {
			for _, vv := range v {
				usecase.AddHeader(k, vv)
			}
		}

		// Try to parse methodName as a fully-qualified method name.
		// If it is valid, use its fully-qualified service.
		fqsn, mtd, err := usecase.ParseFullyQualifiedMethodName(methodName)
		if err == nil {
			pkg, svc := proto.ParseFullyQualifiedServiceName(fqsn)
			if err := usecase.UsePackage(pkg); err != nil {
				return errors.Wrapf(err, "failed to use package '%s'", pkg)
			}
			if err := usecase.UseService(svc); err != nil {
				return errors.Wrapf(err, "failed to use service '%s'", svc)
			}
			methodName = mtd
		}

		err = usecase.CallRPC(ctx, ui.Writer(), methodName)
		if err != nil {
			return errors.Wrapf(err, "failed to call RPC '%s'", methodName)
		}
		return nil
	}, nil
}

func NewListCLIInvoker(ui cui.UI, fqn, format string) CLIInvoker {
	const (
		fname = "name"
		fjson = "json"
	)
	return func(context.Context) error {
		var presenter present.Presenter
		switch format {
		case fname:
			presenter = name.NewPresenter()
		case fjson:
			presenter = json.NewPresenter("  ")
		default:
			presenter = name.NewPresenter()
		}
		usecase.InjectPartially(usecase.Dependencies{ResourcePresenter: presenter})

		commonErr := errors.Errorf("unknown fully-qualified service name or method name '%s'", fqn)
		out, err := func() (string, error) {
			if fqn == "" {
				svc, err := usecase.FormatServices()
				if err != nil {
					return "", errors.Wrap(err, "failed to list services")
				}
				return svc, nil
			}

			if isFullyQualifiedMethodName(fqn) {
				// A fully-qualified method name is passed.
				// Return as it is (same behavior as grpc_cli).
				rpc, err := usecase.FormatMethod(fqn)
				if err != nil {
					return "", errors.Wrap(err, "failed to format RPC")
				}
				return rpc, nil
			}

			// Parse as a fully-qualified service name.

			pkg, svc := proto.ParseFullyQualifiedServiceName(fqn)

			if err := usecase.UsePackage(pkg); err != nil {
				return "", commonErr // Return commonErr because UsePackage will be deprecated.
			}

			if err := usecase.UseService(svc); err != nil && errors.Is(err, idl.ErrUnknownServiceName) {
				return "", commonErr
			} else if err != nil {
				return "", errors.Wrapf(err, "failed to use service '%s'", svc)
			}

			rpcs, err := usecase.FormatMethods()
			if err != nil {
				return "", errors.Wrap(err, "failed to format RPCs")
			}
			return rpcs, nil
		}()
		if err != nil {
			return err
		}
		ui.Output(out)
		return nil
	}
}

func NewDescribeCLIInvoker(ui cui.UI, fqn string) CLIInvoker {
	return func(context.Context) error {
		var (
			out string
			err error
		)
		if fqn != "" {
			out, err = usecase.FormatDescriptor(fqn)
		} else {
			out, err = usecase.FormatServiceDescriptors()
		}
		if err != nil {
			return errors.Wrap(err, "failed to describe")
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
			ResourcePresenter: json.NewPresenter("  "),
		},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := setDefault(cfg); err != nil {
		return err
	}

	return invoker(ctx)
}

// IsCLIMode returns whether Evans is launched as CLI mode or not.
func IsCLIMode(file string) bool {
	return file != "" || (!isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()))
}

func isFullyQualifiedMethodName(s string) bool {
	_, _, err := usecase.ParseFullyQualifiedMethodName(s)
	return err == nil
}
