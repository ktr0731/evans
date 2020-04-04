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

func (f *formatter) FormatMessage(v interface{}) error {
	f.FormatMessageCalled = true
	return nil
}

func (f *formatter) FormatStatus(status *status.Status) {
	f.FormatStatusCalled = true
}

func (f *formatter) FormatTrailer(trailer metadata.MD) {
	f.FormatTrailerCalled = true
}

func (f *formatter) Done() error {
	return nil
}

func TestResponseFormatter(t *testing.T) {
	cases := map[string]struct {
		format map[string]struct{}
	}{
		"all":     {format: map[string]struct{}{"all": {}}},
		"header":  {format: map[string]struct{}{"header": {}}},
		"message": {format: map[string]struct{}{"message": {}}},
		"trailer": {format: map[string]struct{}{"trailer": {}}},
		"status":  {format: map[string]struct{}{"status": {}}},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			impl := &formatter{}
			f := NewResponseFormatter(impl, c.format)
			f.FormatHeader(metadata.Pairs("key", "val"))
			if err := f.FormatMessage(struct{}{}); err != nil {
				t.Fatalf("FormatMessage should no return an error, but got '%s'", err)
			}
			f.FormatTrailer(status.New(codes.Internal, "internal error"), metadata.Pairs("key", "val"))
			if err := f.Done(); err != nil {
				t.Fatalf("Done should not return an error, but got '%s'", err)
			}

			res := map[string]bool{
				"all":     impl.FormatHeaderCalled && impl.FormatMessageCalled && impl.FormatTrailerCalled && impl.FormatStatusCalled,
				"header":  impl.FormatHeaderCalled,
				"message": impl.FormatMessageCalled,
				"trailer": impl.FormatTrailerCalled,
				"status":  impl.FormatStatusCalled,
			}

			for k := range c.format {
				if called, ok := res[k]; ok && !called {
					t.Errorf("Format method associated with %s should be called", k)
				}
			}

			t.Run("Format", func(t *testing.T) {
				impl := &formatter{}
				f := NewResponseFormatter(impl, c.format)
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

				FormatRes := map[string]bool{
					"all":     impl.FormatHeaderCalled && impl.FormatMessageCalled && impl.FormatTrailerCalled && impl.FormatStatusCalled,
					"header":  impl.FormatHeaderCalled,
					"message": impl.FormatMessageCalled,
					"trailer": impl.FormatTrailerCalled,
					"status":  impl.FormatStatusCalled,
				}

				if diff := cmp.Diff(res, FormatRes); diff != "" {
					t.Errorf("two results should be equal:\n%s", diff)
				}
			})
		})
	}
}
