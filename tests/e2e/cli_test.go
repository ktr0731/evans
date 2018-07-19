package e2e

import (
	"bytes"
	"io/ioutil"
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

type expected struct {
	args string
	code int
	out  string
}

func TestCLI(t *testing.T) {
	setup := func() func() {
		old := controller.DefaultCLIReader
		return func() {
			controller.DefaultCLIReader = old
		}
	}
	cleanup := setup()
	defer cleanup()

	defer helper.NewServer(t).Start().Stop()

	t.Run("from stdin", func(t *testing.T) {
		cases := []struct {
			in       string
			expected []expected
		}{
			{
				in: `{ "name": "maho" }`,
				expected: []expected{
					{args: "", code: 1},
					{args: "testdata/helloworld.proto", code: 1},
					{args: "--package helloworld testdata/helloworld.proto", code: 1},
					{args: "--package helloworld --service Greeter testdata/helloworld.proto", code: 1},
					{args: "--package helloworld --call SayHello testdata/helloworld.proto", code: 1},
					{args: "--package helloworld --service Greeter --call SayHello", code: 1},
					{args: "--package helloworld --service Greeter --call SayHello testdata/helloworld.proto", out: `{ "message": "Hello, maho!" }`},
				},
			},
		}

		for _, r := range cases {
			for _, c := range r.expected {
				in := strings.NewReader(r.in)
				controller.DefaultCLIReader = in

				out := new(bytes.Buffer)
				ui := controller.NewUI(in, out, ioutil.Discard)

				code := newCLI(ui).Run(strings.Split(c.args, " "))
				require.Equal(t, c.code, code)

				if c.code == 0 {
					assert.Equal(t, c.out, flatten(out.String()))
				}
			}
		}
	})

	t.Run("from file", func(t *testing.T) {
		cases := []struct {
			in       string
			expected []expected
		}{
			{
				in: `{ "name": "maho" }`,
				expected: []expected{
					{args: "--file testdata/in.json", code: 1},
					{args: "--file testdata/in.json testdata/helloworld.proto", code: 1},
					{args: "--file testdata/in.json --package helloworld testdata/helloworld.proto", code: 1},
					{args: "--file testdata/in.json --package helloworld --service Greeter testdata/helloworld.proto", code: 1},
					{args: "--file testdata/in.json --package helloworld --call SayHello testdata/helloworld.proto", code: 1},
					{args: "--file testdata/in.json --package helloworld --service Greeter --call SayHello", code: 1},
					{args: "--file testdata/in.json --package helloworld --service Greeter --call SayHello testdata/helloworld.proto", out: `{ "message": "Hello, maho!" }`},
				},
			},
		}

		for _, r := range cases {
			for _, c := range r.expected {
				in := strings.NewReader(r.in)
				controller.DefaultCLIReader = in

				out := new(bytes.Buffer)
				ui := controller.NewUI(in, out, ioutil.Discard)

				code := newCLI(ui).Run(strings.Split(c.args, " "))
				require.Equal(t, c.code, code)

				if c.code == 0 {
					assert.Equal(t, c.out, flatten(out.String()))
				}
			}
		}
	})
}
