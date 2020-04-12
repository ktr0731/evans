// Package json provides a JSON formatter implementation.
package json

import (
	"io"

	"github.com/ktr0731/evans/format"
	"github.com/ktr0731/evans/present"
	"github.com/ktr0731/evans/present/json"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// responseFormatter is a formatter that formats *usecase.GRPCResponse into a JSON object.
type responseFormatter struct {
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
	p present.Presenter
}

func NewResponseFormatter(w io.Writer) format.ResponseFormatterInterface {
	return &responseFormatter{w: w, p: json.NewPresenter("  ")}
}

func (p *responseFormatter) FormatHeader(header metadata.MD) {
	p.s.Header = &header
}

func (p *responseFormatter) FormatMessage(v interface{}) error {
	p.s.Messages = append(p.s.Messages, v)
	return nil
}

func (p *responseFormatter) FormatTrailer(trailer metadata.MD) {
	p.s.Trailer = &trailer
}

func (p *responseFormatter) FormatStatus(s *status.Status) {
	p.s.Status = &struct {
		Code    string `json:"code"`
		Number  uint32 `json:"number"`
		Message string `json:"message"`
	}{s.Code().String(), uint32(s.Code()), s.Message()}
}

func (p *responseFormatter) Done() error {
	s, err := p.p.Format(p.s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(p.w, s+"\n")
	return err
}
