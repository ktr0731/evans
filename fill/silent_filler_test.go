package fill_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bufbuild/protocompile"
	"github.com/ktr0731/evans/fill"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestSilentFiller(t *testing.T) {
	cases := map[string]struct {
		in     string
		hasErr bool
	}{
		"normal":       {in: `{"p": "bar"}`},
		"invalid JSON": {in: `foo`, hasErr: true},
	}

	c := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: []string{filepath.Join("proto", "testdata")},
		}),
	}
	compiled, err := c.Compile(context.TODO(), "test.proto")
	if err != nil {
		t.Fatal(err)
	}

	md := compiled[0].Messages().ByName(protoreflect.Name("Message"))

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {

			f := fill.NewSilentFiller(strings.NewReader(c.in))
			i := dynamicpb.NewMessage(md)
			err := f.Fill(i)
			if c.hasErr {
				if err == nil {
					t.Errorf("Fill must return an error, but got nil")
				}
			} else if err != nil {
				t.Errorf("Fill must not return an error, but got an error: '%s'", err)
			}
		})
	}
}
