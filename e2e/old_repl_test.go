package e2e_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/app"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/prompt"
)

var (
	update = flag.Bool("update", false, "update goldens")
)

func init() {
	testing.Init()
	flag.Parse()
	if *update {
		os.RemoveAll(filepath.Join("testdata", "fixtures"))
		if err := os.MkdirAll(filepath.Join("testdata", "fixtures"), 0755); err != nil {
			panic(fmt.Sprintf("failed to create fixture dirs: %s", err))
		}
	}
}

func TestE2E_OldREPL(t *testing.T) {
	// In testing, Go modifies stdin, so we specify --repl explicitly.
	commonFlags := []string{"--silent", "--repl"}

	cases := map[string]struct {
		input []interface{}

		// Space separated arguments text.
		args string

		// The server enables TLS.
		tls bool

		// The server enables gRPC reflection.
		reflection bool

		// The server uses gRPC-Web protocol.
		web bool

		// The exit code we expected.
		expectedCode int

		// skipGolden skips golden file testing.
		skipGolden bool

		// hasErr checks whether REPL wrote some errors to UI.ErrWriter.
		hasErr bool
	}{
		// RPC calls.

		"call Unary by selecting package and service": {
			args:  "testdata/test.proto",
			input: []interface{}{"package api", "service Example", "call Unary", "kaguya"},
		},
		"call Unary by selecting only service": {
			args:  "testdata/test.proto",
			input: []interface{}{"service Example", "call Unary", "kaguya"},
		},
		"call Unary by specifying --service": {
			args:  "--service Example testdata/test.proto",
			input: []interface{}{"call Unary", "kaguya"},
		},
		"call ClientStreaming": {
			args: "testdata/test.proto",
			// io.EOF means end of inputting.
			input: []interface{}{"call ClientStreaming", "kaguya", "chika", "miko", io.EOF},
		},
		"call BidiStreaming": {
			args: "testdata/test.proto",
			// io.EOF means end of inputting.
			input: []interface{}{"call BidiStreaming", "kaguya", "chika", "miko", io.EOF},
		},
		"call UnaryMessage": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryMessage", "kaguya", "shinomiya"},
		},
		"call UnaryRepeated": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryRepeated", "miyuki", "kaguya", "chika", "yu", io.EOF},
		},
		"call UnarySelf": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnarySelf", "dig down", "ohana", "matsumae", "ohana", "dig down", "nako", "oshimizu", "nakochi", "finish", "dig down", "minko", "tsurugi", "minchi", "finish", "finish"},
		},
		"call UnaryMap": {
			args:       "testdata/test.proto",
			input:      []interface{}{"call UnaryMap", "key1", "val1", "key2", "val2", io.EOF},
			skipGolden: true,
		},
		"call UnaryOneof": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryOneof", "msg", "ai", "hayasaka"},
		},
		"call UnaryEnum": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryEnum", "Male"},
		},
		"call UnaryBytes": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryBytes", "\\u3084\\u306f\\u308a\\u4ffa\\u306e\\u9752\\u6625\\u30e9\\u30d6\\u30b3\\u30e1\\u306f\\u307e\\u3061\\u304c\\u3063\\u3066\\u3044\\u308b\\u3002"},
		},
		"call UnaryRepeatedEnum": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryRepeatedEnum", "Male", "Male", "Female", io.EOF},
		},

		// call (gRPC-Web)

		"call client streaming RPC against to gRPC-Web server": {
			args:  "--web testdata/test.proto",
			web:   true,
			input: []interface{}{"call ClientStreaming", "oumae", "kousaka", "kawashima", "kato", io.EOF},
		},
		"call server streaming RPC against to gRPC-Web server": {
			args:  "--web testdata/test.proto",
			web:   true,
			input: []interface{}{"call ServerStreaming", "violet"},
		},
		"call bidi streaming RPC against to gRPC-Web server": {
			args:  "--web testdata/test.proto",
			web:   true,
			input: []interface{}{"call BidiStreaming", "oumae", "kousaka", "kawashima", "kato", io.EOF},
		},

		// show command.

		"show --help": {
			args:  "testdata/test.proto",
			input: []interface{}{"show --help"},
		},
		"show package": {
			args:  "testdata/test.proto",
			input: []interface{}{"show package"},
		},
		"show service": {
			args:  "testdata/test.proto",
			input: []interface{}{"show service"},
		},
		"show message": {
			args:  "testdata/test.proto",
			input: []interface{}{"show message"},
		},
		"show rpc": {
			args:  "testdata/test.proto",
			input: []interface{}{"show rpc"},
		},
		"show an invalid target": {
			args:       "testdata/test.proto",
			input:      []interface{}{"show foo"},
			skipGolden: true,
			hasErr:     true,
		},

		// package command.

		"select a package": {
			args:       "testdata/test.proto",
			input:      []interface{}{"package api"},
			skipGolden: true,
		},
		"specify an invalid package name": {
			args:       "testdata/test.proto",
			input:      []interface{}{"package foo"},
			skipGolden: true,
			hasErr:     true,
		},

		// service command.

		"select a service": {
			args:       "testdata/test.proto",
			input:      []interface{}{"service Example"},
			skipGolden: true,
		},
		"specify an invalid service name": {
			args:       "testdata/test.proto",
			input:      []interface{}{"service foo"},
			skipGolden: true,
			hasErr:     true,
		},

		// header command.

		"header help": {
			args:  "testdata/test.proto",
			input: []interface{}{"header -h"},
		},
		"add a header": {
			args:  "testdata/test.proto",
			input: []interface{}{"header mizore=yoroizuka", "show header"},
		},
		"add a header with --raw flag": {
			args:  "testdata/test.proto",
			input: []interface{}{"header -r touma=youko,kazusa", "show header"},
		},
		"add two headers": {
			args:  "testdata/test.proto",
			input: []interface{}{"header mizore=yoroizuka nozomi=kasaki", "show header"},
		},
		"add two values to a key": {
			args:  "testdata/test.proto",
			input: []interface{}{"header touma=youko", "header touma=kazusa", "show header"},
		},
		"add two values in one command": {
			args:  "testdata/test.proto",
			input: []interface{}{"header touma=youko,kazusa", "show header"},
		},
		"remove a header": {
			args:  "testdata/test.proto",
			input: []interface{}{"header grpc-client", "show header"},
		},

		// desc command.

		"desc simple message": {
			args:  "testdata/test.proto",
			input: []interface{}{"desc SimpleRequest"},
		},
		"desc a repeated message": {
			args:  "testdata/test.proto",
			input: []interface{}{"desc UnaryRepeatedMessageRequest"},
		},
		"desc a map": {
			args:  "testdata/test.proto",
			input: []interface{}{"desc UnaryMapMessageRequest"},
		},
		"desc an invalid message": {
			args:       "testdata/test.proto",
			input:      []interface{}{"desc foo"},
			skipGolden: true,
			hasErr:     true,
		},

		// quit command.

		"quit executes exit": {
			args:       "testdata/test.proto",
			input:      []interface{}{"quit"},
			skipGolden: true,
		},

		// special keys.

		"ctrl-c skips the rest of fields if there are no message type fields": {
			args:  "testdata/test.proto",
			input: []interface{}{"call Unary", prompt.ErrAbort},
		},
		"ctrl-c skips the rest of the current message": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryMessage", "mumei", prompt.ErrAbort},
		},
		"ctrl-c skips the rest of the current message and exits the repeated field": {
			args:  "testdata/test.proto",
			input: []interface{}{"call UnaryRepeatedMessage", "kanade", "hisaishi", "kumiko", prompt.ErrAbort},
		},
		"ctrl-c is also enabled in streaming RPCs": {
			args:  "testdata/test.proto",
			input: []interface{}{"call BidiStreaming", "kanade", "ririka", prompt.ErrAbort, io.EOF},
		},
	}
	oldNewPrompt := prompt.New
	defer func() {
		prompt.New = oldNewPrompt
	}()

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			stopServer, port := startServer(t, c.tls, c.reflection, c.web)
			defer stopServer()

			stubPrompt := &stubPrompt{
				t:      t,
				Prompt: oldNewPrompt(),
				input:  append(c.input, "exit"),
			}
			prompt.New = func(...prompt.Option) prompt.Prompt {
				return stubPrompt
			}

			args := commonFlags
			args = append([]string{"--port", port}, args...)
			if c.args != "" {
				args = append(args, strings.Split(c.args, " ")...)
			}

			w, ew := new(bytes.Buffer), new(bytes.Buffer)
			ui := cui.New(cui.Writer(w), cui.ErrWriter(ew))

			a := app.New(ui)
			code := a.Run(args)
			if code != c.expectedCode {
				t.Errorf("unexpected code returned: expected = %d, actual = %d", c.expectedCode, code)
			}

			if !c.skipGolden {
				compareWithGolden(t, w.String())
			}

			if c.hasErr {
				if ew.String() == "" {
					t.Errorf("expected REPL wrote some error to ew, but empty output")
				}
			} else {
				eout := ew.String()
				// Trim "deprecated" message.
				eout = strings.Replace(eout, color.YellowString("evans: deprecated usage, please use sub-commands. see `evans -h` for more details.")+"\n", "", -1)
				if eout != "" {
					t.Errorf("expected REPL didn't write errors to ew, but got '%s'", eout)
				}
			}
		})
	}
}

type stubPrompt struct {
	t *testing.T
	prompt.Prompt

	input []interface{}
}

func (p *stubPrompt) Input() (string, error) {
	if len(p.input) == 0 {
		p.t.Fatal("p.input is empty, but testing is continued yet. Are you forgot to use io.EOF for finishing inputting?")
	}
	s := p.input[0]
	p.input = p.input[1:]
	switch e := s.(type) {
	case string:
		return e, nil
	case error:
		return "", e
	default:
		panic(fmt.Sprintf("expected string or error, but got '%T'", e))
	}
}

func (p *stubPrompt) Select(string, []string) (string, error) {
	return p.Input()
}

var (
	goldenPathReplacer = strings.NewReplacer(
		"/", "-",
		" ", "_",
		"=", "-",
		"'", "",
		`"`, "",
		",", "",
	)
	goldenReplacer = strings.NewReplacer(
		"\r", "", // For Windows.
	)
)

func compareWithGolden(t *testing.T, actual string) {
	name := t.Name()
	normalizeFilename := func(name string) string {
		fname := goldenPathReplacer.Replace(strings.ToLower(name)) + ".golden"
		return filepath.Join("testdata", "fixtures", fname)
	}

	fname := normalizeFilename(name)

	if *update {
		if err := ioutil.WriteFile(fname, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to update the golden file: %s", err)
		}
		return
	}

	// Load the golden file.
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatalf("failed to load a golden file: %s", err)
	}
	expected := goldenReplacer.Replace(string(b))

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("wrong result: \n%s", diff)
	}
}
