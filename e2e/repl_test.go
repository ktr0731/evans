package e2e_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/ktr0731/evans/app"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/prompt"
)

func TestE2E_REPL(t *testing.T) {
	commonFlags := []string{"--silent"}

	cases := map[string]struct {
		input []any

		// Common flags all sub-commands can have.
		commonFlags string
		// Space separated arguments text.
		args string

		// The server enables TLS.
		tls bool

		// The server enables gRPC reflection.
		reflection bool

		// The server uses gRPC-Web protocol.
		web bool

		// Register a service that has no package.
		registerEmptyPackageService bool

		// The exit code we expected.
		expectedCode int

		// skipGolden skips golden file testing.
		skipGolden bool

		// hasErr checks whether REPL wrote some errors to UI.ErrWriter.
		hasErr bool
	}{
		// Common.

		"corresponding symbol is missing": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show '"},
			skipGolden:  true,
			hasErr:      true,
		},

		// RPC calls.

		"call --help": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call --help"},
		},
		"call Unary by selecting package and service": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package api", "service Example", "call Unary", "kaguya"},
		},
		"call Unary by selecting package and service with enriched output": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package api", "service Example", "call --enrich Unary", "kaguya"},
		},
		"call UnaryHeaderTrailerFailure by selecting package and service with enriched output": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package api", "service Example", "call --enrich UnaryHeaderTrailerFailure", "kaguya"},
			hasErr:      true,
		},
		"call Unary by selecting only service": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"service Example", "call Unary", "kaguya"},
		},
		"call Unary by selecting only fully-qualified service": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"service api.Example", "call Unary", "kaguya"},
		},
		"call Unary by selecting only service (empty package)": {
			registerEmptyPackageService: true,
			commonFlags:                 "--proto testdata/empty_package.proto",
			input:                       []any{"service EmptyPackageService", "call Unary", "kaguya"},
		},
		"call Unary by specifying --service": {
			commonFlags: "--service Example --proto testdata/test.proto",
			input:       []any{"call Unary", "kaguya"},
		},
		"call Unary with --emit-defaults": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package api", "service Example", "call --emit-defaults Unary", ""},
		},
		"call ClientStreaming": {
			commonFlags: "--proto testdata/test.proto",
			// io.EOF means end of inputting.
			input: []any{"call ClientStreaming", "kaguya", "chika", "miko", io.EOF},
		},
		"call BidiStreaming": {
			commonFlags: "--proto testdata/test.proto",
			// io.EOF means end of inputting.
			input: []any{"call BidiStreaming", "kaguya", "chika", "miko", io.EOF},
		},
		"call UnaryMessage": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryMessage", "kaguya", "shinomiya"},
		},
		"call UnaryRepeated": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryRepeated", "miyuki", "kaguya", "chika", "yu", io.EOF},
		},
		"call UnaryRepeated (specify default value)": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryRepeated", "", io.EOF},
		},
		"call UnarySelf": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnarySelf", "ohana", "matsumae", "ohana", "nako", "oshimizu", "nakochi", io.EOF, "minko", "tsurugi", "minchi", io.EOF, io.EOF},
		},
		"call UnaryMap": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryMap", "key1", "val1", "key2", "val2", io.EOF},
			skipGolden:  true,
		},
		"call UnaryOneof": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryOneof", 0, "ai", "hayasaka"},
		},
		"call UnaryEnum": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryEnum", 0},
		},
		"call UnaryBytes": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryBytes", "44KE44Gv44KK5L+644Gu6Z2S5pil44Op44OW44Kz44Oh44Gv44G+44Gh44GM44Gj44Gm44GE44KL44CC"},
		},
		"call UnaryBytes (fallback)": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryBytes", "\\u3084\\u306f\\u308a\\u4ffa\\u306e\\u9752\\u6625\\u30e9\\u30d6\\u30b3\\u30e1\\u306f\\u307e\\u3061\\u304c\\u3063\\u3066\\u3044\\u308b\\u3002"},
		},
		"call UnaryBytes --bytes-as-base64": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryBytes --bytes-as-base64", "44KE44Gv44KK5L+644Gu6Z2S5pil44Op44OW44Kz44Oh44Gv44G+44Gh44GM44Gj44Gm44GE44KL44CC"},
		},
		"call UnaryBytes --bytes-as-quoted-literals": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryBytes --bytes-as-quoted-literals", "\\u3084\\u306f\\u308a\\u4ffa\\u306e\\u9752\\u6625\\u30e9\\u30d6\\u30b3\\u30e1\\u306f\\u307e\\u3061\\u304c\\u3063\\u3066\\u3044\\u308b\\u3002"},
		},
		"call UnaryRepeatedEnum": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryRepeatedEnum", 0, 0, 1, io.EOF},
		},
		"call Unary with an invalid flag": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package api", "service Example", "call -foo Unary", "kaguya"},
			skipGolden:  true,
			hasErr:      true,
		},
		"call UnaryEcho": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call --dig-manually UnaryEcho", 0, "kaguya", "shinomiya"},
		},
		"call UnaryEcho with empty request": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call --dig-manually UnaryEcho", 1},
		},
		"call UnaryEcho with empty name request": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call --dig-manually UnaryEcho", 0, prompt.ErrAbort},
		},

		// call (gRPC-Web)

		"call client streaming RPC against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			web:         true,
			input:       []any{"call ClientStreaming", "oumae", "kousaka", "kawashima", "kato", io.EOF},
		},
		"call server streaming RPC against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			web:         true,
			input:       []any{"call ServerStreaming", "violet"},
		},
		"call bidi streaming RPC against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			web:         true,
			input:       []any{"call BidiStreaming", "oumae", "kousaka", "kawashima", "kato", io.EOF},
		},

		// show command.

		"show --help": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show --help"},
		},
		"show package": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show package"},
		},
		"show package with empty package": {
			commonFlags:                 "--proto testdata/test.proto",
			registerEmptyPackageService: true,
			input:                       []any{"show package"},
		},
		"show service": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show service"},
		},
		"show message": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show message"},
		},
		"show rpc": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show rpc"},
		},
		"show an invalid target": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"show foo"},
			skipGolden:  true,
			hasErr:      true,
		},

		// package command.

		"package --help": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package --help"},
		},
		"select a package": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package api"},
			skipGolden:  true,
		},
		"specify an invalid package name": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"package foo"},
			skipGolden:  true,
			hasErr:      true,
		},

		// service command.

		"service --help": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"service --help"},
		},
		"select a service": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"service Example"},
			skipGolden:  true,
		},
		"specify an invalid service name": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"service foo"},
			skipGolden:  true,
			hasErr:      true,
		},

		// header command.
		"header help": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header -h"},
		},
		"add a header": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header mizore=yoroizuka", "show header"},
		},
		"add a header with --raw flag": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header -r touma=youko,kazusa", "show header"},
		},
		"add two headers": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header mizore=yoroizuka nozomi=kasaki", "show header"},
		},
		"add two values to a key": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header touma=youko", "header touma=kazusa", "show header"},
		},
		"add two values in one command": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header touma=youko,kazusa", "show header"},
		},
		"remove a header": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header grpc-client", "show header"},
		},
		"header with an invalid flag": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header -foo touma=youko"},
			skipGolden:  true,
			hasErr:      true,
		},
		"header with unusable chars": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header sh!nonome=nano"},
			skipGolden:  true,
			hasErr:      true,
		},
		"header with unusable chars and --raw": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"header --raw sh!nonome=nano"},
			skipGolden:  true,
			hasErr:      true,
		},

		// desc command.

		"desc simple message": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"desc SimpleRequest"},
		},
		"desc a fully-qualified message": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"desc api.SimpleRequest"},
		},
		"desc simple message in empty package": {
			commonFlags: "--proto testdata/empty_package.proto",
			input:       []any{"desc SimpleRequest"},
		},
		"desc a repeated message": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"desc UnaryRepeatedMessageRequest"},
		},
		"desc a map": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"desc UnaryMapMessageRequest"},
		},
		"desc an invalid message": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"desc foo"},
			skipGolden:  true,
			hasErr:      true,
		},

		// exit and quit command.

		"exit --help": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"exit --help"},
		},
		"quit executes exit": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"quit"},
			skipGolden:  true,
		},

		// special keys.

		"ctrl-c skips the rest of fields if there are no message type fields": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call Unary", prompt.ErrAbort},
		},
		"ctrl-c skips the rest of the current message": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryMessage", "mumei", prompt.ErrAbort},
		},
		"ctrl-c skips the rest of the current message and exits the repeated field": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call UnaryRepeatedMessage", "kanade", "hisaishi", "kumiko", prompt.ErrAbort, io.EOF},
		},
		"ctrl-c is also enabled in streaming RPCs": {
			commonFlags: "--proto testdata/test.proto",
			input:       []any{"call BidiStreaming", "kanade", "ririka", prompt.ErrAbort, io.EOF},
		},
	}
	oldNewPrompt := prompt.New
	defer func() {
		prompt.New = oldNewPrompt
	}()

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			stopServer, port := startServer(t, c.tls, c.reflection, c.web, c.registerEmptyPackageService)
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
			if c.commonFlags != "" {
				args = append(args, strings.Split(c.commonFlags, " ")...)
			}
			args = append(args, "repl") // Sub-command name.
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
				if ew.String() != "" {
					t.Errorf("expected REPL didn't write errors to ew, but got '%s'", ew.String())
				}
			}
		})
	}
}
