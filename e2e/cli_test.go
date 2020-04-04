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
	"github.com/ktr0731/evans/usecase"
	"github.com/pkg/errors"
)

func TestE2E_CLI(t *testing.T) {
	commonFlags := []string{"--verbose"}

	cases := map[string]struct {
		// Common flags all sub-commands can have.
		commonFlags string
		cmd         string
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
		// assertWithGolden asserts the output with the golden file.
		assertWithGolden bool

		// The exit code we expected.
		expectedCode int

		// Each output is formatted by flatten() for remove break lines.
		// But, if it is prefer to show as it is, you can it by specifying
		// unflatten to true.
		unflatten bool
	}{
		"print usage text to the Writer (common flag)": {
			commonFlags:      "--help",
			assertWithGolden: true,
			unflatten:        true,
		},
		"print usage text to the Writer": {
			args:             "--help",
			assertWithGolden: true,
			unflatten:        true,
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

		// call command

		"print call command usage": {
			commonFlags:      "",
			cmd:              "call",
			args:             "-h",
			assertWithGolden: true,
		},
		"cannot launch CLI mode because proto files didn't be passed": {
			commonFlags:  "--package api --service Example",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --package is invalid value": {
			commonFlags:  "--package foo --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --service is invalid value": {
			commonFlags:  "--package api --service Foo --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			expectedCode: 1,
		},
		"cannot launch CLI mode because method is missing": {
			commonFlags:  "--package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in",
			expectedCode: 1,
		},
		"cannot launch CLI mode because method is invalid value": {
			commonFlags:  "--package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "Foo --file testdata/unary_call.in",
			expectedCode: 1,
		},
		"cannot launch CLI mode because the path of --file is invalid path": {
			commonFlags:  "--package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file foo Unary",
			expectedCode: 1,
		},
		"cannot launch CLI mode because the path of --file is invalid input": {
			commonFlags:  "--package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/invalid.in Unary",
			expectedCode: 1,
		},
		"cannot launch CLI mode because --header didn't have value": {
			commonFlags:  "--header foo --package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			expectedCode: 1,
		},
		"call unary RPC with an input file by CLI mode": {
			commonFlags: "--package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call fully-qualified unary RPC with an input file by CLI mode": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with --call flag (backward-compatibility)": {
			commonFlags: "--package api --service Example --proto testdata/test.proto",
			cmd:         "",
			args:        "--file testdata/unary_call.in --call Unary",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC without package name because the size of packages is 1": {
			commonFlags: "--service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC without package and service name because the size of packages and services are 1": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with an input reader by CLI mode": {
			commonFlags: "--package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "Unary",
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
			commonFlags: "--package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "ClientStreaming --file testdata/client_streaming.in",
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"call server streaming RPC by CLI mode": {
			commonFlags: "--package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "ServerStreaming --file testdata/server_streaming.in",
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"call bidi streaming RPC by CLI mode": {
			commonFlags: "--package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "BidiStreaming --file testdata/bidi_streaming.in",
			assertTest: func(t *testing.T, output string) {
				dec := json.NewDecoder(strings.NewReader(output))
				for {
					var iface interface{}
					err := dec.Decode(&iface)
					if errors.Is(err, io.EOF) {
						return
					}
					if err != nil {
						t.Errorf("expected no errors, but got '%s'", err)
					}
				}
			},
		},
		"call unary RPC with an input file and custom headers by CLI mode": {
			commonFlags: "--header ogiso=setsuna --header touma=kazusa,youko --header sound=of=destiny --package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "UnaryHeader --file testdata/unary_header.in",
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

		// call command with reflection

		"cannot launch CLI mode with reflection because method is missing": {
			commonFlags:  "--reflection --package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in",
			reflection:   true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with reflection because server didn't enable reflection": {
			commonFlags:  "--reflection --package api --service Example",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			reflection:   false,
			expectedCode: 1,
		},
		"call unary RPC by CLI mode with reflection with an input file": {
			commonFlags: "--reflection --package api --service Example",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			reflection:  true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},

		// call command with TLS

		"cannot launch CLI mode with TLS because the server didn't enable TLS": {
			commonFlags:  "--tls --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          false,
			expectedCode: 1,
		},
		"cannot launch CLI mode with TLS because the client didn't enable TLS": {
			commonFlags:  "--service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with TLS because cannot validate certs for 127.0.0.1 (default value)": {
			commonFlags:  "--tls --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with TLS because signed authority is unknown": {
			commonFlags:  "--tls --host localhost --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          true,
			expectedCode: 1,
		},
		"call unary RPC with TLS by CLI mode": {
			commonFlags: "--tls --host localhost --cacert testdata/rootCA.pem --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			tls:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"cannot launch CLI mode with TLS and reflection by CLI mode because server didn't enable TLS": {
			commonFlags:  "--tls -r --host localhost --cacert testdata/rootCA.pem --service Example",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          false,
			reflection:   true,
			expectedCode: 1,
		},
		"call unary RPC with TLS and reflection by CLI mode": {
			commonFlags: "--tls -r --host localhost --cacert testdata/rootCA.pem --service Example",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			tls:         true,
			reflection:  true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with TLS and --servername by CLI mode": {
			commonFlags: "--tls --servername localhost --cacert testdata/rootCA.pem --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			tls:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"cannot launch CLI mode with mutual TLS auth because --certkey is missing": {
			commonFlags:  "--tls --host localhost --cacert testdata/rootCA.pem --cert testdata/localhost.pem --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch CLI mode with mutual TLS auth because --cert is missing": {
			commonFlags:  "--tls --host localhost --cacert testdata/rootCA.pem --certkey testdata/localhost-key.pem --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			tls:          true,
			expectedCode: 1,
		},
		"call unary RPC with mutual TLS auth by CLI mode": {
			commonFlags: "--tls --host localhost --cacert testdata/rootCA.pem --cert testdata/localhost.pem --certkey testdata/localhost-key.pem --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			tls:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},

		// call command with gRPC-Web

		"cannot send a request to gRPC-Web server because the server didn't enable gRPC-Web": {
			commonFlags:  "--web --package api --service Example --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in Unary",
			web:          false,
			expectedCode: 1,
		},
		"call unary RPC with an input file by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --package api --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			web:         true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call unary RPC with an input file by CLI mode and reflection against to gRPC-Web server": {
			commonFlags: "--web -r --service Example",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			web:         true,
			reflection:  true,
			expectedOut: `{ "message": "hello, oumae" }`,
		},
		"call client streaming RPC by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "ClientStreaming --file testdata/client_streaming.in",
			web:         true,
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"call server streaming RPC by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "ServerStreaming --file testdata/server_streaming.in",
			web:         true,
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"call bidi streaming RPC by CLI mode against to gRPC-Web server": {
			commonFlags: "--web --service Example --proto testdata/test.proto",
			cmd:         "call",
			args:        "BidiStreaming --file testdata/bidi_streaming.in",
			web:         true,
			assertTest: func(t *testing.T, output string) {
				dec := json.NewDecoder(strings.NewReader(output))
				for {
					var iface interface{}
					err := dec.Decode(&iface)
					if errors.Is(err, io.EOF) {
						return
					}
					if err != nil {
						t.Errorf("expected no errors, but got '%s'", err)
					}
				}
			},
		},
		"call unary RPC with --print flag (all)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print all UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call unary RPC with --print flag (header)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print header UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call unary RPC with --print flag (message)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print message UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call unary RPC with --print flag (trailer)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print trailer UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call unary RPC with --print flag (status)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print status UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call unary RPC with --print flag (header,message,status)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print header,message,status UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call failure unary RPC with --print flag (all)": {
			commonFlags:      "-r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print header,message,trailer,status UnaryHeaderTrailerFailure",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
			expectedCode:     1,
		},
		// NOTE: Currently, gRPC-Web server implementation returns disorderly Access-Control-Expose-Headers.
		// So, for golden file testing, we discard headers.
		"call unary RPC with --print flag (message,trailer,status) against to gRPC-Web server": {
			commonFlags:      "--web -r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print message,trailer,status UnaryHeaderTrailer",
			web:              true,
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call failure unary RPC with --print flag (message,trailer,status) against to gRPC-Web server": {
			commonFlags:      "--web -r --service Example",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --print message,trailer,status UnaryHeaderTrailerFailure",
			web:              true,
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
			expectedCode:     1,
		},

		// list command
		"print list command usage": {
			commonFlags:      "",
			cmd:              "list",
			args:             "-h",
			assertWithGolden: true,
		},
		"list services without args": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "list",
			args:        "",
			expectedOut: `api.Example`,
		},
		"list services without args and two protos": {
			commonFlags: "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:         "list",
			args:        "",
			expectedOut: `EmptyPackageService api.Example`,
		},
		"list services with name format": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "list",
			args:        "-o name",
			expectedOut: `api.Example`,
		},
		"list services with JSON format": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "list",
			args:        "-o json",
			expectedOut: `{ "services": [ { "name": "api.Example" } ] }`,
		},
		"list methods with name format": {
			commonFlags:      "--proto testdata/test.proto",
			cmd:              "list",
			args:             "-o name api.Example",
			assertWithGolden: true,
		},
		"list methods with JSON format": {
			commonFlags:      "--proto testdata/test.proto",
			cmd:              "list",
			args:             "-o json api.Example",
			assertWithGolden: true,
		},
		"list methods that have empty name": {
			commonFlags: "--proto testdata/empty_package.proto",
			cmd:         "list",
			args:        "EmptyPackageService",
			expectedOut: `EmptyPackageService.Unary`,
		},
		"list a method with name format": {
			commonFlags: "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:         "list",
			args:        "-o name api.Example.Unary",
			expectedOut: `api.Example.Unary`,
		},
		"list a method with JSON format": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "list",
			args:        "-o json api.Example.Unary",
			expectedOut: `{ "name": "Unary", "fully_qualified_name": "api.Example.Unary", "request_type": "api.SimpleRequest", "response_type": "api.SimpleResponse" }`,
		},
		"cannot list because of invalid package name": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "list",
			args:         "Foo",
			expectedCode: 1,
		},
		"cannot list because of invalid service name": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "list",
			args:         "api.Foo",
			expectedCode: 1,
		},

		// desc command

		"print desc command usage": {
			commonFlags:      "",
			cmd:              "desc",
			args:             "-h",
			assertWithGolden: true,
		},
		"describe all service descriptors": {
			commonFlags:      "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:              "desc",
			args:             "",
			assertWithGolden: true,
		},
		"describe a service descriptor": {
			commonFlags:      "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:              "desc",
			args:             "api.Example",
			assertWithGolden: true,
		},
		"describe a method descriptor": {
			commonFlags:      "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:              "desc",
			args:             "api.Example.Unary",
			assertWithGolden: true,
		},
		"describe a message descriptor": {
			commonFlags:      "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:              "desc",
			args:             "api.SimpleRequest",
			assertWithGolden: true,
		},
		"invalid symbol": {
			commonFlags:  "--proto testdata/test.proto,testdata/empty_package.proto",
			cmd:          "desc",
			args:         "api.Foo",
			expectedCode: 1,
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			defer usecase.Clear()

			stopServer, port := startServer(t, c.tls, c.reflection, c.web, c.registerEmptyPackageService)
			defer stopServer()

			outBuf, eoutBuf := new(bytes.Buffer), new(bytes.Buffer)
			cui := cui.New(cui.Writer(outBuf), cui.ErrWriter(eoutBuf))

			args := commonFlags
			args = append([]string{"--port", port}, args...)
			if c.commonFlags != "" {
				args = append(args, strings.Split(c.commonFlags, " ")...)
			}
			args = append(args, "cli")
			if c.cmd != "" {
				args = append(args, c.cmd)
			}
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
				if c.expectedOut != "" && actual != c.expectedOut {
					t.Errorf("unexpected output:\n%s", cmp.Diff(c.expectedOut, actual))
				}
				if eoutBuf.String() != "" {
					t.Errorf("expected code is 0, but got an error message: '%s'", eoutBuf.String())
				}
			}
			if c.assertTest != nil {
				c.assertTest(t, actual)
			}
			if c.assertWithGolden {
				compareWithGolden(t, outBuf.String())
			}
		})
	}
}
