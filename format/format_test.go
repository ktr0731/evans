package format

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type formatter struct {
	FormatHeaderCalled, FormatMessageCalled, FormatStatusCalled, FormatTrailerCalled bool
}

func (f *formatter) FormatHeader(header metadata.MD) {
	f.FormatHeaderCalled = true
}

func (f *formatter) FormatMessage(v any) error {
	f.FormatMessageCalled = true
	return nil
}

func (f *formatter) FormatStatus(status *status.Status) error {
	f.FormatStatusCalled = true
	return nil
}

func (f *formatter) FormatTrailer(trailer metadata.MD) {
	f.FormatTrailerCalled = true
}

func (f *formatter) Done() error {
	return nil
}

func TestResponseFormatter(t *testing.T) {
	cases := map[string]struct {
		enrich bool
	}{
		"enrich=true":  {true},
		"enrich=false": {false},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			impl := &formatter{}
			f := NewResponseFormatter(impl, c.enrich)
			f.FormatHeader(metadata.Pairs("key", "val"))
			if err := f.FormatMessage(struct{}{}); err != nil {
				t.Fatalf("FormatMessage should not return an error, but got '%s'", err)
			}
			if err := f.FormatTrailer(status.New(codes.Internal, "internal error"), metadata.Pairs("key", "val")); err != nil {
				t.Fatalf("FormatTrailer should not return an error, but got '%s'", err)
			}
			if err := f.Done(); err != nil {
				t.Fatalf("Done should not return an error, but got '%s'", err)
			}

			res := map[bool]bool{
				true:  impl.FormatHeaderCalled && impl.FormatMessageCalled && impl.FormatTrailerCalled && impl.FormatStatusCalled,
				false: impl.FormatMessageCalled,
			}

			if called, ok := res[c.enrich]; ok && !called {
				t.Errorf("expected true, but false")
			}

			t.Run("Format", func(t *testing.T) {
				impl := &formatter{}
				f := NewResponseFormatter(impl, c.enrich)
				err := f.Format(
					status.New(codes.Internal, "internal error"),
					metadata.Pairs("key", "val"),
					metadata.Pairs("key", "val"),
					struct{}{},
				)
				if err != nil {
					t.Fatalf("Format should not return an error, but got '%s'", err)
				}
				if err := f.Done(); err != nil {
					t.Fatalf("Done should not return an error, but got '%s'", err)
				}

				FormatRes := map[bool]bool{
					true:  impl.FormatHeaderCalled && impl.FormatMessageCalled && impl.FormatTrailerCalled && impl.FormatStatusCalled,
					false: impl.FormatMessageCalled,
				}

				if diff := cmp.Diff(res, FormatRes); diff != "" {
					t.Errorf("two results should be equal:\n%s", diff)
				}
			})
		})
	}
}
