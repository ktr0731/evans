package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/ktr0731/evans/app"
	"github.com/ktr0731/evans/cui"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/mode"
	"github.com/ktr0731/evans/usecase"
	"github.com/pkg/errors"
)

var replacer = regexp.MustCompile(`access-control-expose-headers: .*\n`)

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

		deprecatedUsage bool
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
		"cannot launch because proto files didn't be passed": {
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			expectedCode: 1,
		},
		"cannot launch because package name is invalid value": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in foo.Example.Unary",
			expectedCode: 1,
		},
		"cannot launch because service name is invalid value": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Foo.Unary",
			expectedCode: 1,
		},
		"cannot launch because method name is missing": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example",
			expectedCode: 1,
		},
		"cannot launch because method name is invalid value": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Foo",
			expectedCode: 1,
		},
		"cannot launch because the path of --file is invalid path": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file foo api.Example.Unary",
			expectedCode: 1,
		},
		"cannot launch because the path of --file is invalid input": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/invalid.in api.Example.Unary",
			expectedCode: 1,
		},
		"cannot launch because --header didn't have value": {
			commonFlags:  "--header foo --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			expectedCode: 1,
		},
		"call unary RPC with an input file (backward-compatibility)": {
			commonFlags:     "--package api --service Example --proto testdata/test.proto",
			cmd:             "call",
			args:            "--file testdata/unary_call.in Unary",
			deprecatedUsage: true,
			expectedOut:     `{ "message": "oumae" }`,
		},
		"call unary RPC with an input file": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			expectedOut: `{ "message": "oumae" }`,
		},
		"call fully-qualified unary RPC with an input file": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			expectedOut: `{ "message": "oumae" }`,
		},
		"call unary RPC with --call flag (backward-compatibility)": {
			commonFlags:     "--package api --service Example --proto testdata/test.proto",
			cmd:             "",
			args:            "--file testdata/unary_call.in --call Unary",
			deprecatedUsage: true,
			expectedOut:     `{ "message": "oumae" }`,
		},
		"call unary RPC without package name because the size of packages is 1 (backward-compatibility)": {
			commonFlags:     "--service Example --proto testdata/test.proto",
			cmd:             "call",
			args:            "--file testdata/unary_call.in Unary",
			deprecatedUsage: true,
			expectedOut:     `{ "message": "oumae" }`,
		},
		"call unary RPC without package and service name because the size of packages and services are 1": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in Unary",
			expectedOut: `{ "message": "oumae" }`,
		},
		"call unary RPC with an input reader": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "api.Example.Unary",
			beforeTest: func(t *testing.T) func(*testing.T) {
				old := mode.DefaultCLIReader
				mode.DefaultCLIReader = strings.NewReader(`{"name": "oumae"}`)
				return func(t *testing.T) {
					mode.DefaultCLIReader = old
				}
			},
			expectedOut: `{ "message": "oumae" }`,
		},
		"call client streaming RPC": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/client_streaming.in api.Example.ClientStreaming",
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"call server streaming RPC": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/server_streaming.in api.Example.ServerStreaming",
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"call bidi streaming RPC": {
			commonFlags: "--proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/bidi_streaming.in api.Example.BidiStreaming",
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
		"call unary RPC with an input file and custom headers": {
			commonFlags: "--header ogiso=setsuna --header touma=kazusa,youko --header sound=of=destiny --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_header.in api.Example.UnaryHeader",
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

		// call command with timeout header.

		"unary call timed out": {
			commonFlags:  "--header grpc-timeout=0s --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.example.unary",
			expectedCode: 1,
		},
		"unary call with timeout header": {
			commonFlags: "--header grpc-timeout=1s --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			expectedOut: `{ "message": "oumae" }`,
		},
		"client streaming call timed out": {
			commonFlags:  "--header grpc-timeout=0s --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/client_streaming.in api.Example.ClientStreaming",
			expectedCode: 1,
		},
		"client streaming call with timeout header": {
			commonFlags: "--header grpc-timeout=1s --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/client_streaming.in api.Example.ClientStreaming",
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"server streaming call timed out": {
			commonFlags:  "--header grpc-timeout=0s --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/server_streaming.in api.Example.ServerStreaming",
			expectedCode: 1,
		},
		"server streaming call with timeout header": {
			commonFlags: "--header grpc-timeout=1s --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/server_streaming.in api.Example.ServerStreaming",
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"bidi streaming call timed out": {
			commonFlags:  "--header grpc-timeout=0s --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/bidi_streaming.in api.Example.BidiStreaming",
			expectedCode: 1,
		},
		"bidi streaming call with timeout header": {
			commonFlags: "--header grpc-timeout=1s --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/bidi_streaming.in api.Example.BidiStreaming",
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

		// call command with reflection

		"cannot launch with reflection because method name is missing": {
			commonFlags:  "--reflection --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example",
			reflection:   true,
			expectedCode: 1,
		},
		"cannot launch with reflection because server didn't enable reflection": {
			commonFlags:  "--reflection",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			reflection:   false,
			expectedCode: 1,
		},
		"call unary RPC with reflection with an input file (backward-compatibility)": {
			commonFlags:     "--reflection --package api --service Example",
			cmd:             "call",
			args:            "--file testdata/unary_call.in Unary",
			reflection:      true,
			deprecatedUsage: true,
			expectedOut:     `{ "message": "oumae" }`,
		},
		"call unary RPC with reflection with an input file": {
			commonFlags: "--reflection",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			reflection:  true,
			expectedOut: `{ "message": "oumae" }`,
		},

		// call command with TLS

		"cannot launch with TLS because the server didn't enable TLS": {
			commonFlags:  "--tls --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          false,
			expectedCode: 1,
		},
		"cannot launch with TLS because the client didn't enable TLS": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch with TLS because cannot validate certs for 127.0.0.1 (default value)": {
			commonFlags:  "--tls --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch with TLS because signed authority is unknown": {
			commonFlags:  "--tls --host localhost --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          true,
			expectedCode: 1,
		},
		"call unary RPC with TLS": {
			commonFlags: "--tls --host localhost --cacert testdata/rootCA.pem --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			tls:         true,
			expectedOut: `{ "message": "oumae" }`,
		},
		"cannot launch with TLS and reflection because server didn't enable TLS": {
			commonFlags:  "--tls -r --host localhost --cacert testdata/rootCA.pem",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          false,
			reflection:   true,
			expectedCode: 1,
		},
		"call unary RPC with TLS and reflection": {
			commonFlags: "--tls -r --host localhost --cacert testdata/rootCA.pem",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			tls:         true,
			reflection:  true,
			expectedOut: `{ "message": "oumae" }`,
		},
		"call unary RPC with TLS and --servername": {
			commonFlags: "--tls --servername localhost --cacert testdata/rootCA.pem --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			tls:         true,
			expectedOut: `{ "message": "oumae" }`,
		},
		"cannot launch with mutual TLS auth because --certkey is missing": {
			commonFlags:  "--tls --host localhost --cacert testdata/rootCA.pem --cert testdata/localhost.pem --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          true,
			expectedCode: 1,
		},
		"cannot launch with mutual TLS auth because --cert is missing": {
			commonFlags:  "--tls --host localhost --cacert testdata/rootCA.pem --certkey testdata/localhost-key.pem --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			tls:          true,
			expectedCode: 1,
		},
		"call unary RPC with mutual TLS auth": {
			commonFlags: "--tls --host localhost --cacert testdata/rootCA.pem --cert testdata/localhost.pem --certkey testdata/localhost-key.pem --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			tls:         true,
			expectedOut: `{ "message": "oumae" }`,
		},

		// call command with gRPC-Web

		"cannot send a request to gRPC-Web server because the server didn't enable gRPC-Web": {
			commonFlags:  "--web --proto testdata/test.proto",
			cmd:          "call",
			args:         "--file testdata/unary_call.in api.Example.Unary",
			web:          false,
			expectedCode: 1,
		},
		"call unary RPC with an input file against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			web:         true,
			expectedOut: `{ "message": "oumae" }`,
		},
		"call unary RPC with an input file and reflection against to gRPC-Web server": {
			commonFlags: "--web -r",
			cmd:         "call",
			args:        "--file testdata/unary_call.in api.Example.Unary",
			web:         true,
			reflection:  true,
			expectedOut: `{ "message": "oumae" }`,
		},
		"call client streaming RPC against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/client_streaming.in api.Example.ClientStreaming",
			web:         true,
			expectedOut: `{ "message": "you sent requests 4 times (oumae, kousaka, kawashima, kato)." }`,
		},
		"call server streaming RPC against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/server_streaming.in api.Example.ServerStreaming",
			web:         true,
			expectedOut: `{ "message": "hello oumae, I greet 1 times." } { "message": "hello oumae, I greet 2 times." } { "message": "hello oumae, I greet 3 times." }`,
		},
		"call bidi streaming RPC against to gRPC-Web server": {
			commonFlags: "--web --proto testdata/test.proto",
			cmd:         "call",
			args:        "--file testdata/bidi_streaming.in api.Example.BidiStreaming",
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
		"call unary RPC with --enrich flag": {
			commonFlags:      "-r",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --enrich api.Example.UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call unary RPC with --enrich flag and JSON format": {
			commonFlags:      "-r",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --enrich --output json api.Example.UnaryHeaderTrailer",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call failure unary RPC with --enrich flag": {
			commonFlags:      "-r",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --enrich api.Example.UnaryHeaderTrailerFailure",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
			expectedCode:     1,
		},
		"call failure unary RPC with --enrich and JSON format": {
			commonFlags:      "-r",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --enrich --output json api.Example.UnaryHeaderTrailerFailure",
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
			expectedCode:     1,
		},
		"call unary RPC with --enrich flag against to gRPC-Web server": {
			commonFlags:      "--web -r",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --enrich api.Example.UnaryHeaderTrailer",
			web:              true,
			reflection:       true,
			unflatten:        true,
			assertWithGolden: true,
		},
		"call failure unary RPC with --enrich flag against to gRPC-Web server": {
			commonFlags:      "--web -r",
			cmd:              "call",
			args:             "--file testdata/unary_call.in --enrich api.Example.UnaryHeaderTrailerFailure",
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
		"cannot list because of invalid method name": {
			commonFlags:  "--proto testdata/test.proto",
			cmd:          "list",
			args:         "api.Example.UnaryFoo",
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
				eout := eoutBuf.String()
				if c.deprecatedUsage {
					// Trim "deprecated" message.
					eout = strings.ReplaceAll(eout, color.YellowString("evans: deprecated usage, please use sub-commands. see `evans -h` for more details.")+"\n", "")
				}
				if eout != "" {
					t.Errorf("expected code is 0, but got an error message: '%s'", eoutBuf.String())
				}
			}
			if c.assertTest != nil {
				c.assertTest(t, actual)
			}
			if c.assertWithGolden {
				s := outBuf.String()
				s = replacer.ReplaceAllString(s, "")
				compareWithGolden(t, s)
			}
		})
	}
}
