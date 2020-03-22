// Package format provides formatting APIs for display processed result application did.
package format

import (
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ResponseFormatter provides formatting feature for gRPC response.
type ResponseFormatter struct {
	format map[string]struct{}
	ResponseFormatterInterface
}

func (f *ResponseFormatter) Format(s *status.Status, header, trailer metadata.MD, v interface{}) error {
	f.FormatHeader(header)
	if err := f.FormatMessage(v); err != nil {
		return err
	}
	f.FormatTrailer(s, trailer)
	return nil
}

func (f *ResponseFormatter) FormatHeader(header metadata.MD) {
	if (has(f.format, "all") || has(f.format, "header")) && header.Len() != 0 {
		f.ResponseFormatterInterface.FormatHeader(header)
	}
}

func (f *ResponseFormatter) FormatMessage(v interface{}) error {
	if (!has(f.format, "message") && !has(f.format, "all")) || v == nil {
		return nil
	}
	return f.ResponseFormatterInterface.FormatMessage(v)
}

func (f *ResponseFormatter) FormatTrailer(status *status.Status, trailer metadata.MD) {
	if (has(f.format, "all") || has(f.format, "trailer")) && trailer.Len() != 0 {
		f.ResponseFormatterInterface.FormatTrailer(trailer)
	}

	if has(f.format, "all") || has(f.format, "status") {
		f.ResponseFormatterInterface.FormatStatus(status)
	}
}

func NewResponseFormatter(f ResponseFormatterInterface, format map[string]struct{}) *ResponseFormatter {
	return &ResponseFormatter{
		ResponseFormatterInterface: f,
		format:                     format,
	}
}

type ResponseFormatterInterface interface {
	// FormatHeader formats the response header.
	FormatHeader(header metadata.MD)
	// FormatMessage formats the response message (body).
	FormatMessage(v interface{}) error
	// FormatStatus formats the response status.
	FormatStatus(status *status.Status)
	// FormatTrailer formats the response trailer.
	FormatTrailer(trailer metadata.MD)
	// Done indicates all response information is formatted.
	// The client of ResponseFormatter should call it at the end.
	Done() error
}

func has(m map[string]struct{}, k string) bool {
	_, ok := m[k]
	return ok
}
