package e2e

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/evans/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCLI(ui controller.UI) *controller.CLI {
	return controller.NewCLI(meta.AppName, meta.Version.String(), ui)
}

func flatten(s string) string {
	s = strings.Replace(s, "\n", " ", -1)
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(" +")
	return re.ReplaceAllString(s, " ")
}

func TestCLI(t *testing.T) {
	in := strings.NewReader(`{ "name": "maho" }`)

	controller.DefaultCLIReader = in
	defer func() {
		controller.DefaultCLIReader = os.Stdin
	}()

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
			{args: "--package helloworld --service Greeter --call SayHello testdata/helloworld.proto"},
		}

		for _, c := range cases {
			out := new(bytes.Buffer)
			ui := controller.NewUI(in, out, ioutil.Discard)

			code := newCLI(ui).Run(strings.Split(c.args, " "))
			require.Equal(t, c.code, code)

			if c.code == 0 {
				assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
			}
		}
	})

	t.Run("from file", func(t *testing.T) {
		cases := []struct {
			args string
			code int
		}{
			{args: "--file testdata/in.json", code: 1},
			{args: "--file testdata/in.json testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld --service Greeter testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld --call SayHello testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld --service Greeter --call SayHello", code: 1},
			{args: "--file testdata/in.json --package helloworld --service Greeter --call SayHello testdata/helloworld.proto"},
		}

		for _, c := range cases {
			out := new(bytes.Buffer)
			ui := controller.NewUI(in, out, ioutil.Discard)

			code := newCLI(ui).Run(strings.Split(c.args, " "))
			require.Equal(t, c.code, code)

			if c.code == 0 {
				assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
			}
		}
	})
}
