// Package curl provides a curl-like formatter implementation.
package curl

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ktr0731/evans/format"
	"github.com/ktr0731/evans/present"
	"github.com/ktr0731/evans/present/json"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type responsePresenter struct {
	w io.Writer

	json present.Presenter

	wroteHeader, wroteMessage, wroteTrailer bool
}

func NewResponseFormatter(w io.Writer) format.ResponseFormatterInterface {
	return &responsePresenter{
		w:    w,
		json: json.NewPresenter("  "),
	}
}

func (p *responsePresenter) FormatHeader(header metadata.MD) {
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

func (p *responsePresenter) FormatMessage(v interface{}) error {
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

func (p *responsePresenter) FormatTrailer(trailer metadata.MD) {
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

	p.wroteTrailer = true
}

func (p *responsePresenter) FormatStatus(status *status.Status) {
	if p.wroteHeader || p.wroteMessage || p.wroteTrailer {
		fmt.Fprintf(p.w, "\n")
	}
	fmt.Fprintf(p.w, "code = %s, number = %d, message = %q\n", status.Code().String(), status.Code(), status.Message())
}

func (p *responsePresenter) Done() error {
	return nil
}
