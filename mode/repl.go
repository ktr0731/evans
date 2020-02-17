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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

	wroteHeader, wroteMessage bool
}

func newCurlLikeResponsePresenter(w io.Writer, format map[string]struct{}) *curlLikeResponsePresenter {
	return &curlLikeResponsePresenter{
		w:      w,
		format: format,
		json:   json.NewPresenter("  "),
	}
}

func (p *curlLikeResponsePresenter) Format(s *status.Status, header, trailer metadata.MD, v interface{}) error {
	p.FormatHeader(header)
	if err := p.FormatMessage(v); err != nil {
		return err
	}
	p.FormatTrailer(s, trailer)
	return nil
}

func (p *curlLikeResponsePresenter) FormatHeader(header metadata.MD) {
	if has(p.format, "header") {
		if header.Len() == 0 {
			return
		}

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

		p.wroteHeader = true
	}
}

func (p *curlLikeResponsePresenter) FormatMessage(v interface{}) error {
	if !has(p.format, "message") {
		return nil
	}

	if v == nil {
		return nil
	}

	if p.wroteHeader {
		fmt.Fprintf(p.w, "\n")
	}
	msg, err := p.json.Format(v)
	if err != nil {
		return err
	}
	fmt.Fprintf(p.w, "%s\n", msg)

	p.wroteMessage = true

	return nil
}

func (p *curlLikeResponsePresenter) FormatTrailer(status *status.Status, trailer metadata.MD) {
	var wroteTrailer bool
	if has(p.format, "trailer") && trailer.Len() != 0 {
		if p.wroteHeader || p.wroteMessage {
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

		wroteTrailer = true
	}

	if has(p.format, "status") {
		if p.wroteHeader || p.wroteMessage || wroteTrailer {
			fmt.Fprintf(p.w, "\n")
		}
		fmt.Fprintf(p.w, "code: %s, number: %d, message: %q\n", status.Code().String(), status.Code(), status.Message())
	}
}

func (p *curlLikeResponsePresenter) Done() error {
	return nil
}

// jsonResponsePresenter is a formatter that formats *usecase.GRPCResponse into a JSON object.
type jsonResponsePresenter struct {
	w io.Writer
	s struct {
		Status *struct {
			Code    string `json:"code"`
			Number  uint32 `json:"number"`
			Message string `json:"message"`
		} `json:"status,omitempty"`
		Header   *metadata.MD  `json:"header,omitempty"`
		Messages []interface{} `json:"messages,omitempty"`
		Trailer  *metadata.MD  `json:"trailer,omitempty"`
	}
	p      present.Presenter
	format map[string]struct{}
}

func newJSONResponsePresenter(w io.Writer, format map[string]struct{}) *jsonResponsePresenter {
	return &jsonResponsePresenter{w: w, p: json.NewPresenter("  "), format: format}
}

func (p *jsonResponsePresenter) Format(s *status.Status, header, trailer metadata.MD, v interface{}) error {
	p.FormatHeader(header)
	_ = p.FormatMessage(v)
	p.FormatTrailer(s, trailer)
	p.Done()
	return nil
}

func (p *jsonResponsePresenter) FormatHeader(header metadata.MD) {
	if has(p.format, "header") {
		p.s.Header = &header
	}
}

func (p *jsonResponsePresenter) FormatMessage(v interface{}) error {
	if has(p.format, "message") {
		p.s.Messages = append(p.s.Messages, v)
	}
	return nil
}

func (p *jsonResponsePresenter) FormatTrailer(s *status.Status, trailer metadata.MD) {
	if has(p.format, "status") {
		p.s.Status = &struct {
			Code    string `json:"code"`
			Number  uint32 `json:"number"`
			Message string `json:"message"`
		}{s.Code().String(), uint32(s.Code()), s.Message()}
	}
	if has(p.format, "trailer") {
		p.s.Trailer = &trailer
	}
}

func (p *jsonResponsePresenter) Done() error {
	s, err := p.p.Format(p.s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(p.w, s+"\n")
	return err
}

func has(m map[string]struct{}, k string) bool {
	_, ok := m[k]
	return ok
}
