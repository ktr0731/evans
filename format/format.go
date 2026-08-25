// Package format provides formatting APIs for display processed result application did.
package format

import (
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ResponseFormatter provides formatting feature for gRPC response.
type ResponseFormatter struct {
	enrich bool

	impl ResponseFormatterInterface
}

func (f *ResponseFormatter) Format(s *status.Status, header, trailer metadata.MD, v interface{}) error {
	f.FormatHeader(header)
	if err := f.FormatMessage(v); err != nil {
		return err
	}
	if err := f.FormatTrailer(s, trailer); err != nil {
		return err
	}
	return nil
}

func (f *ResponseFormatter) FormatHeader(header metadata.MD) {
	if f.enrich {
		f.impl.FormatHeader(header)
	}
}

func (f *ResponseFormatter) FormatMessage(v interface{}) error {
	if v == nil {
		return nil
	}
	return f.impl.FormatMessage(v)
}

func (f *ResponseFormatter) FormatTrailer(status *status.Status, trailer metadata.MD) error {
	if f.enrich {
		f.impl.FormatTrailer(trailer)
		if err := f.impl.FormatStatus(status); err != nil {
			return err
		}
	}
	return nil
}

// FormatDuration formats the elapsed time of the RPC. It is only emitted for
// enriched output, alongside the header, trailer and status.
func (f *ResponseFormatter) FormatDuration(d time.Duration) {
	if f.enrich {
		f.impl.FormatDuration(d)
	}
}

func (f *ResponseFormatter) Done() error {
	return f.impl.Done()
}

// NewResponseFormatter formats gRPC response with a specific formatter.
// If enrich is false, the formatter prints only messages.
// Or else, it prints all includes headers, messages, trailers and status.
func NewResponseFormatter(f ResponseFormatterInterface, enrich bool) *ResponseFormatter {
	return &ResponseFormatter{impl: f, enrich: enrich}
}

// ResponseFormatterInterface is an interface for formatting gRPC response.
type ResponseFormatterInterface interface {
	// FormatHeader formats the response header.
	FormatHeader(header metadata.MD)
	// FormatMessage formats the response message (body).
	FormatMessage(v interface{}) error
	// FormatStatus formats the response status.
	FormatStatus(status *status.Status) error
	// FormatTrailer formats the response trailer.
	FormatTrailer(trailer metadata.MD)
	// FormatDuration formats the elapsed time of the RPC.
	FormatDuration(d time.Duration)
	// Done indicates all response information is formatted.
	// The client of ResponseFormatter should call it at the end.
	Done() error
}
