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
			args string
			code int // exit code, 1 if precondition failed

			hasErr bool // error was occurred in repl, false if precondition failed

			// Server config
			useReflection bool
			useWeb        bool
			specifyCA     bool
			useTLS        bool
		}{
			{args: "", code: 1}, // cannot launch repl case
			{args: "--package api", code: 1},
			{args: "--service Example", code: 1},
			{args: "--package api --service Example", code: 1},
			{args: "--package foo testdata/api.proto", code: 1},
			{args: "--package api --service foo testdata/api.proto", code: 1},
			{args: "--package api --service Example testdata/api.proto"},

			{args: "--reflection --package api --service Example", useReflection: true},
			{args: "--reflection --service Example", useReflection: true},     // Package api package is inferred.
			{args: "--reflection --service api.Example", useReflection: true}, // Specify package by --service flag.
			{args: "--reflection", useReflection: true},                       // Package api and service Example are inferred.

			{args: "--web", useWeb: true, code: 1},
			{args: "--web --package api", useWeb: true, code: 1},
			{args: "--web --service Example", useWeb: true, code: 1},
			{args: "--web --package api --service Example", useWeb: true, code: 1},
			{args: "--web --package foo --service Example testdata/api.proto", useWeb: true, code: 1},
			{args: "--web --package api --service foo testdata/api.proto", useWeb: true, code: 1},
			{args: "--web --package api --service Example testdata/api.proto", useWeb: true},

			{args: "--web --reflection --service Example", useReflection: true, useWeb: true},
			{args: "--web --reflection --service bar", useReflection: true, useWeb: true, code: 1},

			{args: "--tls --host localhost -r --service Example", useReflection: true, specifyCA: true, useTLS: true},
			{args: "--tls --cert testdata/cert/localhost.pem --certkey testdata/cert/localhost-key.pem --host localhost -r --service Example", useReflection: true, specifyCA: true, useTLS: true},
			{args: "--tls --insecure --host localhost -r --service Example", useReflection: true, specifyCA: true, useTLS: true},
			{args: "--tls --host localhost -r --service Example", useReflection: true, useTLS: true, code: 1},
		}

		rh := newREPLHelper([]string{"--silent", "--repl"})

		cleanup := func() {
			rh.reset()
			di.Reset()
		}

		for _, c := range cases {
			t.Run(c.args, func(t *testing.T) {
				srv, port := newServer(t, c.useReflection, c.useTLS, c.useWeb)
				defer srv.Serve().Stop()
				defer cleanup()

				out, eout := new(bytes.Buffer), new(bytes.Buffer)
				rh.w = out
				rh.ew = eout

				rh.registerInput(
					cmd.Call("Unary", "maho"),
				)

				args := strings.Split(c.args, " ")
				// the first test case.
				if len(args) == 1 && args[0] == "" {
					args = []string{"--port", port}
				} else {
					args = append([]string{"--port", port}, args...)
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
					assert.Equal(t, `{ "message": "hello, maho" }`, flatten(out.String()), eout.String())
				}
			})
		}
	})
}
