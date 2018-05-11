package e2e

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	cmd "github.com/ktr0731/evans/tests/e2e/repl"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
)

func TestREPL(t *testing.T) {
	defer helper.NewServer(t).Start().Stop()

	t.Run("from stdin", func(t *testing.T) {
		cases := []struct {
			args   string
			code   int  // exit code, 1 when precondition failed
			hasErr bool // error was occurred in REPL, false if precondition failed
		}{
			{args: "", code: 1},
			{args: "--package helloworld", code: 1},
			{args: "--service Greeter", code: 1},
			{args: "testdata/helloworld.proto", hasErr: true},
			{args: "--package helloworld --service Greeter", code: 1},
			{args: "--package helloworld testdata/helloworld.proto", hasErr: true},
			{args: "--service Greeter testdata/helloworld.proto", code: 1},
			{args: "--package foo testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service foo testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter testdata/helloworld.proto"},
		}

		rh := newREPLHelper([]string{"--silent", "--repl"})

		for i, c := range cases {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				defer rh.reset()

				out, eout := new(bytes.Buffer), new(bytes.Buffer)
				rh.w = out
				rh.ew = eout

				rh.registerInput(
					cmd.Call("SayHello", "maho"),
				)

				code := rh.run(strings.Split(c.args, " "))
				assert.Equal(t, c.code, code)

				if c.hasErr {
					assert.NotEmpty(t, eout.String())
				}
				// normal case
				if c.code == 0 && !c.hasErr {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
				}
			})
		}
	})
}
