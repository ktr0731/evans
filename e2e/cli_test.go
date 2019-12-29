package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/app"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/mode"
)

func TestE2E_CLI(t *testing.T) {
	commonFlags := []string{"--verbose"}

	cases := map[string]struct {
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

		// beforeTest set up a testcase specific environment.
		// If beforeTest is nil, it is ignored.
		// beforeTest may return a function named afterTest that cleans up
		// the environment. If afterTest is nil, it is ignored.
		beforeTest func(t *testing.T) (afterTest func(t *testing.T))

		// assertTest checks whether the output is expected.
		// If nil, it will be ignored.
		assertTest func(t *testing.T, output string)

		// The output we expected. It is ignored if expectedCode isn't 0.
		expectedOut string

		// The exit code we expected.
		expectedCode int

		// Each output is formatted by flatten() for remove break lines.
		// But, if it is prefer to show as it is, you can it by specifying
		// unflatten to true.
		unflatten bool
	}{
		"print usage text to the Writer (common flag)": {
			commonFlags: "--help",
			expectedOut: expectedCLIUsageOut,
			unflatten:   true,
		},
		"print usage text to the Writer": {
			args:        "--help",
			expectedOut: expectedCLIUsageOut,
			unflatten:   true,
		},
		"print version text to the Writer (common flag)": {
			commonFlags: "--version",
			expectedOut: fmt.Sprintf("evans %s\n", meta.Version),
			unflatten:   true,
		},
		"print version text to the Writer": {
			args:        "--version",
			expectedOut: fmt.Sprintf("evans %s\n", meta.Version),
			unflatten:   true,
		},
		"cannot specify both of --cli and --repl": {
			args:         "--cli --repl",
			expectedCode: 1,
		},
		"cannot specify both of --tls and --web": {
			args:         "--web --tls testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch without proto files and reflection": {
			args:         "",
			expectedCode: 1,
		},

		// CLI mode

		"cannot launch CLI mode because proto files didn't be passed": {
			commonFlags:  "--package api --service Example",
			args:         "--call Unary --file testdata/unary_call.in",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --package is invalid value": {
			commonFlags:  "--package foo --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --service is invalid value": {
			commonFlags:  "--package api --service Foo",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --call is missing": {
			commonFlags:  "--package api --service Example",
			args:         "--file testdata/unary_call.in testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --call is invalid value": {
			commonFlags:  "--package api --service Example",
			args:         "--call Foo --file testdata/unary_call.in testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch CLI mode because the path of --file is invalid path": {
			commonFlags:  "--package api --service Example",
			args:         "--call Unary --file foo testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch CLI mode because the path of --file is invalid input": {
			commonFlags:  "--package api --service Example",
			args:         "--call Unary --file testdata/invalid.in testdata/test.proto",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --header didn't have value": {
			commonFlags:  "--header foo --package api --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			expectedCode: 1,
		},
		"call unary RPC with an input file by CLI mode": {
			commonFlags: "--package api --service Example",
			args:        "--call Unary --file testdata/unary_call.in testdata/test.proto",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC without package name because the size of packages is 1": {
			commonFlags: "--service Example --call Unary",
			args:        "--file testdata/unary_call.in testdata/test.proto",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC without package and service name because the size of packages and services are 1": {
			args:        "--call Unary --file testdata/unary_call.in testdata/test.proto",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with an input reader by CLI mode": {
			commonFlags: "--package api --service Example",
			args:        "--call Unary testdata/test.proto",
			beforeTest: func(t *testing.T) func(*testing.T) {
				old := mode.DefaultCLIReader
				mode.DefaultCLIReader = strings.NewReader(`{"name": "oumae"}`)
				return func(t *testing.T) {
					mode.DefaultCLIReader = old
				}
			},
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call client streaming RPC by CLI mode": {
			commonFlags: "--package api --service Example",
			args:        "--call ClientStreaming --file testdata/client_streaming.in testdata/test.proto",
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"call server streaming RPC by CLI mode": {
			commonFlags: "--package api --service Example",
			args:        "--call ServerStreaming --file testdata/server_streaming.in testdata/test.proto",
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"call bidi streaming RPC by CLI mode": {
			commonFlags: "--package api --service Example",
			args:        "--call BidiStreaming --file testdata/bidi_streaming.in testdata/test.proto",
			assertTest: func(t *testing.T, output string) {
				dec := json.NewDecoder(strings.NewReader(output))
				for {
					var iface interface{}
					err := dec.Decode(&iface)
					if err == io.EOF {
						return
					}
					if err != nil {
						t.Errorf("expected no errors, but got '%s'", err)
					}
				}
			},
		},
		"call unary RPC with an input file and custom headers by CLI mode": {
			commonFlags: "--header ogiso=setsuna --header touma=kazusa,youko --header sound=of=destiny --package api --service Example",
			args:        "--call UnaryHeader --file testdata/unary_header.in testdata/test.proto",
			assertTest: func(t *testing.T, output string) {
				expectedStrings := []string{
					"key = ogiso",
					"val = setsuna",
					"key = touma",
					"val = kazusa, youko",
					"key = sound",
					"val = of=destiny",
				}
				for _, s := range expectedStrings {
					if !strings.Contains(output, s) {
						t.Errorf("expected to contain '%s', but missing in '%s'", s, output)
					}
				}
			},
		},

		// CLI mode with reflection

		"cannot launch CLI mode with reflection because --call is missing": {
			commonFlags:  "--reflection --package api --service Example",
			args:         "testdata/test.proto --file testdata/unary_call.in",
			reflection:   true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with reflection because server didn't enable reflection": {
			commonFlags:  "--reflection --package api --service Example",
			args:         "--call Unary --file testdata/unary_call.in",
			reflection:   false,
			expectedCode: 1,
		},
		"call unary RPC by CLI mode with reflection with an input file": {
			commonFlags: "--reflection --package api --service Example",
			args:        "--call Unary --file testdata/unary_call.in",
			reflection:  true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},

		// CLI mode with TLS

		"cannot launch CLI mode with TLS because the server didn't enable TLS": {
			commonFlags:  "--tls --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:          false,
			expectedCode: 1,
		},
		"cannot launch CLI mode with TLS because the client didn't enable TLS": {
			commonFlags:  "--service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with TLS because cannot validate certs for 127.0.0.1 (default value)": {
			commonFlags:  "--tls --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with TLS because signed authority is unknown": {
			commonFlags:  "--tls --host localhost --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:          true,
			expectedCode: 1,
		},
		"call unary RPC with TLS by CLI mode": {
			commonFlags: "--tls --host localhost --cacert testdata/rootCA.pem --service Example",
			args:        "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"cannot launch CLI mode with TLS and reflection by CLI mode because server didn't enable TLS": {
			commonFlags:  "--tls -r --host localhost --cacert testdata/rootCA.pem --service Example",
			args:         "--call Unary --file testdata/unary_call.in",
			tls:          false,
			reflection:   true,
			expectedCode: 1,
		},
		"call unary RPC with TLS and reflection by CLI mode": {
			commonFlags: "--tls -r --host localhost --cacert testdata/rootCA.pem --service Example",
			args:        "--call Unary --file testdata/unary_call.in",
			tls:         true,
			reflection:  true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with TLS and --servername by CLI mode": {
			commonFlags: "--tls --servername localhost --cacert testdata/rootCA.pem --service Example",
			args:        "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"cannot launch CLI mode with mutual TLS auth because --certkey is missing": {
			commonFlags:  "--tls --host localhost --cacert testdata/rootCA.pem --cert testdata/localhost.pem --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with mutual TLS auth because --cert is missing": {
			commonFlags:  "--tls --host localhost --cacert testdata/rootCA.pem --certkey testdata/localhost-key.pem --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:          true,
			expectedCode: 1,
		},
		"call unary RPC with mutual TLS auth by CLI mode": {
			commonFlags: "--tls --host localhost --cacert testdata/rootCA.pem --cert testdata/localhost.pem --certkey testdata/localhost-key.pem --service Example",
			args:        "--call Unary --file testdata/unary_call.in testdata/test.proto",
			tls:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},

		// CLI mode with gRPC-Web

		"cannot send a request to gRPC-Web server because the server didn't enable gRPC-Web": {
			commonFlags:  "--web --package api --service Example",
			args:         "--call Unary --file testdata/unary_call.in testdata/test.proto testdata/test.proto",
			web:          false,
			expectedCode: 1,
		},
		"call unary RPC with an input file by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --package api --service Example",
			args:        "--call Unary --file testdata/unary_call.in testdata/test.proto testdata/test.proto",
			web:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with an input file by CLI mode and reflection against to gRPC-Web server": {
			commonFlags: "--web -r --service Example",
			args:        "--call Unary --file testdata/unary_call.in",
			web:         true,
			reflection:  true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call client streaming RPC by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --service Example",
			args:        "--call ClientStreaming --file testdata/client_streaming.in testdata/test.proto",
			web:         true,
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"call server streaming RPC by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --service Example",
			args:        "--call ServerStreaming --file testdata/server_streaming.in testdata/test.proto",
			web:         true,
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"call bidi streaming RPC by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --service Example",
			args:        "--call BidiStreaming --file testdata/bidi_streaming.in testdata/test.proto",
			web:         true,
			assertTest: func(t *testing.T, output string) {
				dec := json.NewDecoder(strings.NewReader(output))
				for {
					var iface interface{}
					err := dec.Decode(&iface)
					if err == io.EOF {
						return
					}
					if err != nil {
						t.Errorf("expected no errors, but got '%s'", err)
					}
				}
			},
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			stopServer, port := startServer(t, c.tls, c.reflection, c.web)
			defer stopServer()

			outBuf, eoutBuf := new(bytes.Buffer), new(bytes.Buffer)
			cui := cui.New(cui.Writer(outBuf), cui.ErrWriter(eoutBuf))

			args := commonFlags
			args = append([]string{"--port", port}, args...)
			if c.commonFlags != "" {
				args = append(args, strings.Split(c.commonFlags, " ")...)
			}
			args = append(args, "cli") // Sub-command name.
			if c.args != "" {
				args = append(args, strings.Split(c.args, " ")...)
			}

			if c.beforeTest != nil {
				afterTest := c.beforeTest(t)
				if afterTest != nil {
					defer afterTest(t)
				}
			}

			a := app.New(cui)
			code := a.Run(args)
			if code != c.expectedCode {
				t.Errorf("unexpected code returned: expected = %d, actual = %d", c.expectedCode, code)
			}

			actual := outBuf.String()
			if !c.unflatten {
				actual = flatten(actual)
			}

			if c.expectedCode == 0 {
				if c.assertTest != nil {
					c.assertTest(t, actual)
				}
				if c.expectedOut != "" && actual != c.expectedOut {
					t.Errorf("unexpected output:\n%s", cmp.Diff(c.expectedOut, actual))
				}
				if eoutBuf.String() != "" {
					t.Errorf("expected code is 0, but got an error message: '%s'", eoutBuf.String())
				}
			}
		})
	}
}

var expectedCLIUsageOut = fmt.Sprintf(`evans %s

Usage: evans [--help] [--version] [options ...] [PROTO [PROTO ...]]

Positional arguments:
        PROTO                          .proto files

Options:
        --call string            call specified RPC by CLI mode
        --file, -f string        a script file that will be executed by (used only CLI mode)
        --help, -h               display help text and exit (default "false")

`, meta.Version)
