package mode

import (
	"context"
	"fmt"
	"io"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
			ResponsePresenter: newCurlLikeResponsePresenter(ui.Writer(), nil),
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
	w io.Writer

	format map[string]struct{}
	json   present.Presenter
}

func newCurlLikeResponsePresenter(w io.Writer, format map[string]struct{}) *curlLikeResponsePresenter {
	return &curlLikeResponsePresenter{
		w:      w,
		format: format,
		json:   json.NewPresenter("  "),
	}
}

func (p *curlLikeResponsePresenter) Format(s codes.Code, header, trailer metadata.MD, v interface{}) error {
	p.FormatHeader(header)
	if err := p.FormatMessage(v); err != nil {
		return err
	}
	p.FormatTrailer(s, trailer)
	return nil
}

func (p *curlLikeResponsePresenter) FormatHeader(header metadata.MD) {
	if has(p.format, "header") {
		var s []string
		for k, v := range header {
			for _, vv := range v {
				s = append(s, fmt.Sprintf("%s: %s", k, vv))
			}
		}
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		fmt.Fprintf(p.w, "%s\n", strings.Join(s, "\n"))
	}
}

func (p *curlLikeResponsePresenter) FormatMessage(v interface{}) error {
	if !has(p.format, "message") {
		return nil
	}

	if has(p.format, "header") {
		fmt.Fprintf(p.w, "\n")
	}
	msg, err := p.json.Format(v)
	if err != nil {
		return err
	}
	fmt.Fprintf(p.w, "%s\n", msg)
	return nil
}

func (p *curlLikeResponsePresenter) FormatTrailer(status codes.Code, trailer metadata.MD) {
	if has(p.format, "trailer") {
		if has(p.format, "header") || has(p.format, "message") {
			fmt.Fprintf(p.w, "\n")
		}

		var s []string
		for k, v := range trailer {
			for _, vv := range v {
				s = append(s, fmt.Sprintf("%s: %s", k, vv))
			}
		}
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		fmt.Fprintf(p.w, "%s\n", strings.Join(s, "\n"))
	}

	if has(p.format, "status") {
		if has(p.format, "trailer") {
			fmt.Fprintf(p.w, "\n")
		}
		fmt.Fprintf(p.w, "%d %s\n", status, status.String())
	}
}

func (p *curlLikeResponsePresenter) Done() error {
	return nil
}

// jsonResponsePresenter is a formatter that formats *usecase.GRPCResponse into a JSON object.
type jsonResponsePresenter struct {
	w io.Writer
	s struct {
		Status struct {
			Code       uint32 `json:"code"`
			StringCode string `json:"string_code"`
		} `json:"status"`
		Header  metadata.MD `json:"header_metadata"`
		Message interface{} `json:"message"`
		Trailer metadata.MD `json:"trailer_metadata"`
	}
	p present.Presenter
}

func newJSONResponsePresenter(w io.Writer) *jsonResponsePresenter {
	return &jsonResponsePresenter{w: w, p: json.NewPresenter("  ")}
}

func (p *jsonResponsePresenter) Format(s codes.Code, header, trailer metadata.MD, v interface{}) error {
	p.FormatHeader(header)
	_ = p.FormatMessage(v)
	p.FormatTrailer(s, trailer)
	p.Done()
	return nil
}

func (p *jsonResponsePresenter) FormatHeader(header metadata.MD) {
	p.s.Header = header
}

func (p *jsonResponsePresenter) FormatMessage(v interface{}) error {
	p.s.Message = v
	return nil
}

func (p *jsonResponsePresenter) FormatTrailer(s codes.Code, trailer metadata.MD) {
	p.s.Status = struct {
		Code       uint32 `json:"code"`
		StringCode string `json:"string_code"`
	}{uint32(s), s.String()}
	p.s.Trailer = trailer
}

func (p *jsonResponsePresenter) Done() error {
	s, err := p.p.Format(p.s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(p.w, s)
	return err
}

func has(m map[string]struct{}, k string) bool {
	_, ok := m[k]
	return ok
}
