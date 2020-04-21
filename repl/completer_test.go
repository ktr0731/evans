package repl

import (
	"strings"
	"testing"

	"github.com/ktr0731/evans/idl/proto"
	"github.com/ktr0731/evans/usecase"
)

type dummyDocument struct {
	textBeforeCursor string
}

func (d *dummyDocument) GetWordBeforeCursor() string {
	sp := strings.Split(d.textBeforeCursor, " ")
	return sp[len(sp)-1]
}

func (d *dummyDocument) TextBeforeCursor() string {
	return d.textBeforeCursor
}

func TestCompleter(t *testing.T) {
	cmpl := newCompleter(commands)
	spec, err := proto.LoadFiles([]string{"testdata"}, []string{"test.proto"})
	if err != nil {
		t.Fatalf("LoadFiles must not return an error, but got '%s'", err)
	}
	usecase.Inject(usecase.Dependencies{Spec: spec})
	err = usecase.UsePackage("api")
	if err != nil {
		t.Fatalf("UsePackage must not return an error, but got '%s'", err)
	}
	err = usecase.UseService("Example")
	if err != nil {
		t.Fatalf("UseService must not return an error, but got '%s'", err)
	}

	cases := map[string]struct {
		text      string
		isDefault bool
		isEmpty   bool
	}{
		"empty":                   {text: "", isEmpty: true},
		"show":                    {text: "show "},
		"show returns nothing1":   {text: "show f", isEmpty: true},
		"show returns nothing2":   {text: "show package p", isEmpty: true},
		"package":                 {text: "package "},
		"package returns nothing": {text: "package api ", isEmpty: true},
		"service":                 {text: "service "},
		"service returns nothing": {text: "service Example ", isEmpty: true},
		"call":                    {text: "call "},
		"call flag1":              {text: "call -"},
		"call flag2":              {text: "call --"},
		"call flag3":              {text: "call --e"},
		"call returns nothing":    {text: "call RPC ", isEmpty: true},
		"desc":                    {text: "desc "},
		"desc returns nothing":    {text: "desc Request ", isEmpty: true},
		"header":                  {text: "header -"},
		"default":                 {text: "s", isDefault: true},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			doc := &dummyDocument{textBeforeCursor: c.text}
			suggestions := cmpl.Complete(doc)
			if c.isDefault {
				if len(suggestions) != 0 &&
					strings.HasPrefix(suggestions[len(suggestions)-1].Text, "--help") {
					t.Errorf("in default completion, it must not return --help suggestion")
				}
			} else {
				if len(suggestions) != 0 && !strings.HasPrefix(suggestions[0].Text, "--") &&
					!strings.HasPrefix(suggestions[len(suggestions)-1].Text, "--help") {
					t.Errorf("completion must return --help suggestion at the final suggestion")
				}
			}

			if c.isEmpty {
				if n := len(suggestions); n != 0 {
					t.Errorf("completion must not return any suggestions, but got %d", n)
				}
			} else {
				if n := len(suggestions); n == 0 {
					t.Errorf("completion must return some suggestions, but got no suggestions")
				}
			}
		})
	}
}
