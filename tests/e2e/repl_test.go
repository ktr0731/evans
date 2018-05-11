package e2e

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	cmd "github.com/ktr0731/evans/tests/e2e/repl"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREPL(t *testing.T) {
	defer helper.NewServer(t).Start().Stop()

	t.Run("from stdin", func(t *testing.T) {
		cases := []struct {
			args string
			code int
		}{
			{args: "", code: 1},
			{args: "testdata/helloworld.proto", code: 1},
			{args: "--package helloworld testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --call SayHello testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter --call SayHello", code: 1},
			{args: "--package helloworld --service Greeter testdata/helloworld.proto"},
		}

		rh := newREPLHelper([]string{"--silent", "--repl"})

		for i, c := range cases {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				defer rh.reset()

				out := new(bytes.Buffer)
				rh.w = out

				rh.registerInput(
					cmd.Call("SayHello", "maho"),
				)

				code := rh.run(strings.Split(c.args, " "))
				require.Equal(t, c.code, code)

				if c.code == 0 {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
				}
			})
		}
	})
}
