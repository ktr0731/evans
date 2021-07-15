package proto

import (
	"runtime"
	"testing"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/prompt"
)

type stubPrompt struct {
	prompt.Prompt

	t *testing.T

	input []string

	idx       int
	selection []string
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

	idx := p.idx
	sel := p.selection[0]
	p.idx++
	p.selection = p.selection[1:]

	return idx, sel, nil
}

func (p *stubPrompt) SetPrefix(string) {}

func (p *stubPrompt) SetPrefixColor(prompt.Color) {}

func TestInteractiveFiller(t *testing.T) {
	b := builder.NewMessage("Message")
	b.AddField(builder.NewField("a", builder.FieldTypeMessage(builder.NewMessage("SubMessage"))).SetRepeated())
	b.AddField(builder.NewField("b", builder.FieldTypeEnum(
		builder.NewEnum("Enum").
			AddValue(builder.NewEnumValue("foo")).
			AddValue(builder.NewEnumValue("bar")),
	)))
	b.AddField(builder.NewField("c", builder.FieldTypeDouble()))
	b.AddField(builder.NewField("d", builder.FieldTypeFloat()))
	b.AddField(builder.NewField("e", builder.FieldTypeInt64()))
	b.AddField(builder.NewField("f", builder.FieldTypeSFixed64()))
	b.AddField(builder.NewField("g", builder.FieldTypeSInt64()))
	b.AddField(builder.NewField("h", builder.FieldTypeUInt64()))
	b.AddField(builder.NewField("i", builder.FieldTypeFixed64()))
	b.AddField(builder.NewField("j", builder.FieldTypeInt32()))
	b.AddField(builder.NewField("k", builder.FieldTypeSFixed32()))
	b.AddField(builder.NewField("l", builder.FieldTypeSInt32()))
	b.AddField(builder.NewField("m", builder.FieldTypeUInt32()))
	b.AddField(builder.NewField("n", builder.FieldTypeFixed32()))
	b.AddField(builder.NewField("o", builder.FieldTypeBool()))
	b.AddField(builder.NewField("p", builder.FieldTypeString()))
	b.AddField(builder.NewField("q", builder.FieldTypeBytes()))
	b.AddField(builder.NewField("r", builder.FieldTypeBytes()))
	b.AddField(builder.NewField("s", builder.FieldTypeBytes()))
	m, err := b.Build()
	if err != nil {
		t.Fatalf("Build should not return an error, but got '%s'", err)
	}

	msg := dynamic.NewMessage(m)
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
		selection: []string{"0", "1", "1"},
	}
	f := NewInteractiveFiller(p, "")
	if err := f.Fill(msg, fill.InteractiveFillerOpts{BytesFromFile: true}); err != nil {
		t.Errorf("should not return an error, but got '%s'", err)
	}

	if runtime.GOOS == "windows" {
		t.Skip()
		return
	}

	const want = `{"a":[{}],"b":2,"c":1.1,"d":1.2,"e":"1","f":"2","g":"3","h":"4","i":"5","j":6,"k":7,"l":8,"m":9,"n":10,"o":true,"p":"foo","q":"YmFy","r":"YmF6","s":"Ly8gUGFja2FnZSBwcm90byBwcm92aWRlcyBhIGZpbGxlciBpbXBsZW1lbnRhdGlvbiBmb3IgUHJvdG9jb2wgQnVmZmVycy4KcGFja2FnZSBwcm90bwo="}`

	marshaler := jsonpb.Marshaler{EmitDefaults: true}
	got, err := marshaler.MarshalToString(msg)
	if err != nil {
		t.Fatalf("MarshalToString should not return an error, but got '%s'", err)
	}

	if want != got {
		t.Errorf("want: %s\ngot: %s", want, got)
	}
}
