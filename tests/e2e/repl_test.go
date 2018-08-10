package e2e

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ktr0731/evans/di"
	cmd "github.com/ktr0731/evans/tests/e2e/repl"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestREPL(t *testing.T) {
	t.Run("from stdin", func(t *testing.T) {
		cases := []struct {
			args          string
			code          int  // exit code, 1 when precondition failed
			hasErr        bool // error was occurred in REPL, false if precondition failed
			useReflection bool
			useWeb        bool
		}{
			{args: "", code: 1}, // cannot launch REPL case
			{args: "--package helloworld", code: 1},
			{args: "--service Greeter", code: 1},
			{args: "testdata/helloworld.proto", hasErr: true},
			{args: "--package helloworld --service Greeter", code: 1},
			{args: "--package helloworld testdata/helloworld.proto", hasErr: true},
			{args: "--service Greeter testdata/helloworld.proto", code: 1},
			{args: "--package foo testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service foo testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter testdata/helloworld.proto"},

			{args: "--reflection", hasErr: true, useReflection: true},
			{args: "--reflection --service Greeter", useReflection: true},

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

			{args: "--web --reflection --service Greeter", useReflection: true, useWeb: true, code: 1},
		}

		rh := newREPLHelper([]string{"--silent", "--repl"})

		cleanup := func() {
			rh.reset()
			di.Reset()
		}

		for _, c := range cases {
			t.Run(c.args, func(t *testing.T) {
				defer helper.NewServer(t, c.useReflection).Start(c.useWeb).Stop()
				defer cleanup()

				out, eout := new(bytes.Buffer), new(bytes.Buffer)
				rh.w = out
				rh.ew = eout

				rh.registerInput(
					cmd.Call("SayHello", "maho"),
				)

				args := strings.Split(c.args, " ")
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
