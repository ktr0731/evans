package e2e

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ktr0731/evans/di"
	cmd "github.com/ktr0731/evans/tests/e2e/repl"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestREPL(t *testing.T) {
	defer helper.NewServer(t, false).Start().Stop()

	t.Run("from stdin", func(t *testing.T) {
		cases := []struct {
			args          string
			code          int  // exit code, 1 when precondition failed
			hasErr        bool // error was occurred in REPL, false if precondition failed
			useReflection bool
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
			{args: "--reflection", code: 1, useReflection: true},
			{args: "--reflection --service Greeter", code: 0, useReflection: true},
		}

		rh := newREPLHelper([]string{"--silent", "--repl"})

		cleanup := func() {
			rh.reset()
			di.Reset()
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				defer cleanup()

				out, eout := new(bytes.Buffer), new(bytes.Buffer)
				rh.w = out
				rh.ew = eout

				rh.registerInput(
					cmd.Call("SayHello", "maho"),
				)

				args := strings.Split(c.args, " ")
				args = append(args, "--repl")
				code := rh.run(args)
				assert.Equal(t, c.code, code, eout.String())

				if c.hasErr {
					assert.NotEmpty(t, eout.String(), eout.String())
				}
				// normal case
				if c.code == 0 && !c.hasErr {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()), eout.String())
				}
			})
		}
	})
}
