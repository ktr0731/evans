package mode

import (
	"context"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill/proto"
	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/present/json"
	"github.com/ktr0731/evans/present/table"
	"github.com/ktr0731/evans/prompt"
	"github.com/ktr0731/evans/repl"
	"github.com/ktr0731/evans/usecase"
	"github.com/pkg/errors"
)

func RunAsREPLMode(cfg *config.Config, ui cui.UI, cache *cache.Cache) error {
	gRPCClient, err := newGRPCClient(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to instantiate a new gRPC client")
	}
	defer gRPCClient.Close(context.Background())

	spec, err := newSpec(cfg, gRPCClient)
	if err != nil {
		return errors.Wrap(err, "failed to instantiate a new spec")
	}

	usecase.Inject(
		spec,
		proto.NewInteractiveFiller(prompt.New(), cfg.REPL.InputPromptFormat),
		gRPCClient,
		json.NewPresenter(),
		table.NewPresenter(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := setDefault(cfg); err != nil {
		return err
	}

	for k, v := range cfg.Request.Header {
		for _, vv := range v {
			usecase.AddHeader(k, vv)
		}
	}

	replPrompt := prompt.New(prompt.WithCommandHistory(cache.CommandHistory))
	replPrompt.SetPrefixColor(prompt.ColorBlue)

	defer func() {
		history := make([]string, 0, len(replPrompt.GetCommandHistory()))
		encountered := map[string]interface{}{}
		for _, e := range replPrompt.GetCommandHistory() {
			if _, found := encountered[e]; found {
				continue
			}
			history = append(history, e)
			encountered[e] = nil
		}
		if len(history) > cfg.REPL.HistorySize {
			history = history[len(history)-cfg.REPL.HistorySize:]
		}
		cache.CommandHistory = history
		if err := cache.Save(); err != nil {
			logger.Printf("failed to write command history: %s", err)
		}
	}()

	repl, err := repl.New(cfg, replPrompt, ui, cfg.Default.Package, cfg.Default.Service)
	if err != nil {
		return errors.Wrap(err, "failed to launch a new REPL")
	}
	return repl.Run(ctx)
}
