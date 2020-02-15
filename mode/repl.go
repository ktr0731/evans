package mode

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/fill/proto"
	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/present"
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
		usecase.Dependencies{
			Spec:              spec,
			Filler:            proto.NewInteractiveFiller(prompt.New(), cfg.REPL.InputPromptFormat),
			GRPCClient:        gRPCClient,
			ResponsePresenter: newCurlLikeResponsePresenter(nil),
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

type curlLikeResponsePresenter struct {
	format map[string]struct{}
	json   present.Presenter
}

func newCurlLikeResponsePresenter(format map[string]struct{}) *curlLikeResponsePresenter {
	return &curlLikeResponsePresenter{
		format: format,
		json:   json.NewPresenter("  "),
	}
}

func (p *curlLikeResponsePresenter) Format(res *usecase.GRPCResponse) (string, error) {
	specified := func(k string) bool {
		_, ok := p.format[k]
		return ok
	}
	var b strings.Builder
	if specified("status") {
		fmt.Fprintf(&b, "%d %s\n", res.Status, res.Status.String())
	}
	if specified("header") {
		var s []string
		for k, v := range res.Header {
			for _, vv := range v {
				s = append(s, fmt.Sprintf("%s: %s", k, vv))
			}
		}
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		fmt.Fprintf(&b, "%s\n", strings.Join(s, "\n"))
	}

	if specified("message") {
		fmt.Fprintf(&b, "\n")
		msg, err := p.json.Format(res.Message)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s", msg)
	}

	if specified("trailer") {
		fmt.Fprintf(&b, "\n\n")
		var s []string
		for k, v := range res.Trailer {
			for _, vv := range v {
				s = append(s, fmt.Sprintf("%s: %s", k, vv))
			}
		}
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		fmt.Fprintf(&b, "%s", strings.Join(s, "\n"))
	}
	return strings.TrimSpace(b.String()), nil
}

// jsonResponsePresenter is a formatter that formats *usecase.GRPCResponse into a JSON object.
type jsonResponsePresenter struct {
	p present.Presenter
}

func newJSONResponsePresenter() *jsonResponsePresenter {
	return &jsonResponsePresenter{p: json.NewPresenter("  ")}
}

func (p *jsonResponsePresenter) Format(res *usecase.GRPCResponse) (string, error) {
	return p.p.Format(res)
}
