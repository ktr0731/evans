// +build e2e

package e2e

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ktr0731/evans/di"
	cmd "github.com/ktr0731/evans/tests/e2e/repl"
	"github.com/stretchr/testify/assert"
)

// TODO: add tests for other commands.
func TestREPL(t *testing.T) {
	t.Run("call", func(t *testing.T) {
		cases := []struct {
			args          string
			code          int  // exit code, 1 when precondition failed
			hasErr        bool // error was occurred in repl, false if precondition failed
			useReflection bool
			useWeb        bool

			specifyCA bool
			useTLS    bool
		}{
			{args: "", code: 1}, // cannot launch repl case
			{args: "--package helloworld", code: 1},
			{args: "--service Greeter", code: 1},
			{args: "testdata/helloworld.proto", hasErr: true},
			{args: "--package helloworld --service Greeter", code: 1},
			{args: "--package helloworld testdata/helloworld.proto", hasErr: true},
			{args: "--service Greeter testdata/helloworld.proto", code: 1},
			{args: "--package foo testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service foo testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter testdata/helloworld.proto"},

			{args: "--reflection --package helloworld", hasErr: true, useReflection: true},
			{args: "--reflection --package helloworld --service Greeter", useReflection: true},

			{args: "--web", useWeb: true, code: 1},
			{args: "--web --package helloworld", useWeb: true, code: 1},
			{args: "--web --service Greeter", useWeb: true, code: 1},
			{args: "--web testdata/helloworld.proto", useWeb: true, hasErr: true},
			{args: "--web --package helloworld --service Greeter", useWeb: true, code: 1},
			{args: "--web --package helloworld testdata/helloworld.proto", useWeb: true, hasErr: true},
			{args: "--web --service Greeter testdata/helloworld.proto", useWeb: true, code: 1},
			{args: "--web --package foo --service Greeter testdata/helloworld.proto", useWeb: true, code: 1},
			{args: "--web --package helloworld --service foo testdata/helloworld.proto", useWeb: true, code: 1},
			{args: "--web --package helloworld --service Greeter testdata/helloworld.proto", useWeb: true},

			{args: "--web --reflection --package helloworld --service Greeter", useReflection: true, useWeb: true},
			{args: "--web --reflection --package helloworld --service bar", useReflection: true, useWeb: true, code: 1},

			{args: "--tls --host localhost -r --package helloworld --service Greeter", useReflection: true, specifyCA: true, useTLS: true},
			{args: "--tls --cert testdata/cert/localhost.pem --certkey testdata/cert/localhost-key.pem --host localhost -r --package helloworld --service Greeter", useReflection: true, specifyCA: true, useTLS: true},
			{args: "--tls --insecure --host localhost -r --package helloworld --service Greeter", useReflection: true, specifyCA: true, useTLS: true},
			{args: "--tls --host localhost -r --service Greeter", useReflection: true, useTLS: true, code: 1},
		}

		rh := newREPLHelper([]string{"--silent", "--repl"})

		cleanup := func() {
			rh.reset()
			di.Reset()
		}

		for _, c := range cases {
			t.Run(c.args, func(t *testing.T) {
				srv := newServer(t, c.useReflection, c.useTLS)

				defer srv.start(c.useWeb).stop()
				defer cleanup()

				out, eout := new(bytes.Buffer), new(bytes.Buffer)
				rh.w = out
				rh.ew = eout

				rh.registerInput(
					cmd.Call("SayHello", "maho"),
				)

				args := strings.Split(c.args, " ")
				// the first test case.
				if len(args) == 1 && args[0] == "" {
					args = []string{"--port", srv.port}
				} else {
					args = append([]string{"--port", srv.port}, args...)
				}
				if c.useTLS && c.specifyCA {
					args = append([]string{"--cacert", "testdata/cert/rootCA.pem"}, args...)
				}
				code := rh.run(args)
				assert.Equal(t, c.code, code, eout.String())

				if c.hasErr {
					assert.NotEmpty(t, eout.String())
				}
				// normal case
				if c.code == 0 && !c.hasErr {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()), eout.String())
				}
			})
		}
	})
}
