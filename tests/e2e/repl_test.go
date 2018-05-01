package e2e

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newREPL(t *testing.T, ui controller.UI) *controller.REPL {
	return controller.NewREPL(helper.TestConfig(), entity.NewEnv(), ui)
}

func TestREPL(t *testing.T) {
	in := strings.NewReader(`{ "name": "maho" }`)

	defer helper.NewServer(t).Start().Stop()

	controller.NewREPL(helper.TestConfig())
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
			{args: "--package helloworld --service Greeter --call SayHello testdata/helloworld.proto"},
		}

		for _, c := range cases {
			out := new(bytes.Buffer)
			ui := controller.NewUI(in, out, ioutil.Discard)

			code := newCLI(t, ui).Run(strings.Split(c.args, " "))
			require.Equal(t, c.code, code)

			if c.code == 0 {
				assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
			}
		}
	})
}
