package mode

import (
	"context"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill/proto"
	"github.com/ktr0731/evans/present/json"
	"github.com/ktr0731/evans/prompt"
	"github.com/ktr0731/evans/repl"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/go-multierror"
	"github.com/pkg/errors"
)

func RunAsREPLMode(cfg *config.Config, ui *cui.UI) error {
	var result error
	gRPCClient, err := newGRPCClient(cfg)
	if err != nil {
		result = multierror.Append(result, err)
	} else {
		defer gRPCClient.Close(context.Background())
	}

	spec, err := newSpec(cfg, gRPCClient)
	if err != nil {
		result = multierror.Append(result, err)
	}

	usecase.Inject(
		spec,
		proto.NewInteractiveFiller(prompt.New(), cfg.REPL.InputPromptFormat),
		gRPCClient,
		json.NewPresenter(),
	)

	if result != nil {
		return result
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: remove duplication
	// TODO: signal handling
	if cfg.Default.Package == "" && len(spec.PackageNames()) == 1 {
		cfg.Default.Package = spec.PackageNames()[0]
	}

	if cfg.Default.Service == "" {
		svcNames, err := spec.ServiceNames(cfg.Default.Package)
		if err == nil && len(svcNames) == 1 {
			cfg.Default.Service = svcNames[0]
		}
	}

	for k, v := range cfg.Request.Header {
		for _, vv := range v {
			usecase.AddHeader(k, vv)
		}
	}

	// Pass empty value. See repl.New comments for more details.
	replPrompt := prompt.New()
	replPrompt.SetPrefixColor(prompt.ColorBlue)

	repl, err := repl.New(cfg, replPrompt, ui, cfg.Default.Package, cfg.Default.Service)
	if err != nil {
		return errors.Wrap(err, "failed to launch a new REPL")
	}
	repl.Run(ctx)

	return nil
}
