package proto

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/bufbuild/protocompile"
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/prompt"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type stubPrompt struct {
	prompt.Prompt

	t *testing.T

	input []string

	idx       int
	selection []int
}

func (p *stubPrompt) Input() (string, error) {
	p.t.Helper()

	if len(p.input) == 0 {
		p.t.Fatal("no input")
	}

	in := p.input[0]
	p.input = p.input[1:]

	return in, nil
}

func (p *stubPrompt) Select(string, []string) (int, string, error) {
	p.t.Helper()

	if len(p.selection) == 0 {
		p.t.Fatal("no selection")
	}

	sel := p.selection[p.idx]
	p.idx++

	return sel, fmt.Sprintf("%d", sel), nil
}

func (p *stubPrompt) SetPrefix(string) {}

func (p *stubPrompt) SetPrefixColor(prompt.Color) {}

func TestInteractiveFiller(t *testing.T) {
	c := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: []string{"testdata"},
		}),
	}
	compiled, err := c.Compile(context.TODO(), "test.proto")
	if err != nil {
		t.Fatal(err)
	}

	m := compiled[0].Messages().ByName(protoreflect.Name("Message"))
	msg := dynamicpb.NewMessage(m)
	p := &stubPrompt{
		t: t,
		input: []string{
			"1.1",          // c
			"1.2",          // d
			"1",            // e
			"2",            // f
			"3",            // g
			"4",            // h
			"5",            // i
			"6",            // j
			"7",            // k
			"8",            // l
			"9",            // m
			"10",           // n
			"true",         // o
			"foo",          // p
			"bar",          // q
			"\x62\x61\x7a", // r
			"./proto.go",   // s
		},
		selection: []int{
			0, // a - yes
			1, // a - no
			1, // b - enum2
		},
	}
	f := NewInteractiveFiller(p, "")
	if err := f.Fill(msg, fill.InteractiveFillerOpts{BytesFromFile: true}); err != nil {
		t.Errorf("should not return an error, but got '%s'", err)
	}

	if runtime.GOOS == "windows" {
		t.Skip()
		return
	}

	const want = `{"a":[{}],"b":"enum2","c":1.1,"d":1.2,"e":"1","f":"2","g":"3","h":"4","i":"5","j":6,"k":7,"l":8,"m":9,"n":10,"o":true,"p":"foo","q":"YmFy","r":"YmF6","s":"Ly8gUGFja2FnZSBwcm90byBwcm92aWRlcyBhIGZpbGxlciBpbXBsZW1lbnRhdGlvbiBmb3IgUHJvdG9jb2wgQnVmZmVycy4KcGFja2FnZSBwcm90bwo="}`

	marshaler := jsonpb.Marshaler{EmitDefaults: true}
	got, err := marshaler.MarshalToString(msg)
	if err != nil {
		t.Fatalf("MarshalToString should not return an error, but got '%s'", err)
	}

	if want != got {
		t.Errorf("want: %s\ngot: %s", want, got)
	}
}

func Test_defaultValueFromKind(t *testing.T) {
	cases := map[string]struct {
		kind protoreflect.Kind

		expected protoreflect.Value
	}{
		"string": {
			kind:     protoreflect.StringKind,
			expected: protoreflect.ValueOf(""),
		},
		"double": {
			kind:     protoreflect.DoubleKind,
			expected: protoreflect.ValueOf(float64(0)),
		},
		"float": {
			kind:     protoreflect.FloatKind,
			expected: protoreflect.ValueOf(float32(0)),
		},
		"int64": {
			kind:     protoreflect.Int64Kind,
			expected: protoreflect.ValueOf(int64(0)),
		},
		"uint64": {
			kind:     protoreflect.Uint64Kind,
			expected: protoreflect.ValueOf(uint64(0)),
		},
		"int32": {
			kind:     protoreflect.Int32Kind,
			expected: protoreflect.ValueOf(int32(0)),
		},
		"uint32": {
			kind:     protoreflect.Uint32Kind,
			expected: protoreflect.ValueOf(uint32(0)),
		},
		"fixed64": {
			kind:     protoreflect.Fixed64Kind,
			expected: protoreflect.ValueOf(uint64(0)),
		},
		"fixed32": {
			kind:     protoreflect.Fixed32Kind,
			expected: protoreflect.ValueOf(uint32(0)),
		},
		"bool": {
			kind:     protoreflect.BoolKind,
			expected: protoreflect.ValueOf(false),
		},
		"bytes": {
			kind:     protoreflect.BytesKind,
			expected: protoreflect.ValueOf([]byte{}),
		},
		"sfixed64": {
			kind:     protoreflect.Sfixed64Kind,
			expected: protoreflect.ValueOf(int64(0)),
		},
		"sfixed32": {
			kind:     protoreflect.Sfixed32Kind,
			expected: protoreflect.ValueOf(int32(0)),
		},
		"sint64": {
			kind:     protoreflect.Sint64Kind,
			expected: protoreflect.ValueOf(int64(0)),
		},
		"sint32": {
			kind:     protoreflect.Sint32Kind,
			expected: protoreflect.ValueOf(int32(0)),
		},
	}

	for name, c := range cases {
		name, c := name, c
		t.Run(name, func(t *testing.T) {
			actual := defaultValueFromKind(c.kind)
			if !reflect.DeepEqual(c.expected, actual) {
				t.Errorf("expected '%v' (type = %T), but got '%v' (type = %T)",
					c.expected, c.expected, actual, actual)
			}
		})
	}
}
