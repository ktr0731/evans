package mode

import (
	"context"
	"sort"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill/proto"
	"github.com/ktr0731/evans/logger"
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

	usecase.Inject(
		usecase.Dependencies{
			InteractiveFiller: proto.NewInteractiveFiller(prompt.New(), cfg.REPL.InputPromptFormat),
			GRPCClient:        gRPCClient,
			ResourcePresenter: table.NewPresenter(),
		},
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
		history := tidyUpHistory(replPrompt.GetCommandHistory(), cfg.REPL.HistorySize)
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

func tidyUpHistory(h []string, maxHistorySize int) []string {
	m := make(map[string]int)
	for i := range h {
		m[h[i]] = i
	}
	s := make([]int, 0, len(m))
	for _, i := range m {
		s = append(s, i)
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	history := make([]string, 0, len(h))
	for _, i := range s {
		history = append(history, h[i])
	}
	if len(history) > maxHistorySize {
		history = history[len(history)-maxHistorySize:]
	}
	return history
}
