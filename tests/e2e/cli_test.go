// +build e2e

package e2e

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ktr0731/evans/adapter/cli"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/di"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func flatten(s string) string {
	s = strings.Replace(s, "\n", " ", -1)
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(" +")
	return re.ReplaceAllString(s, " ")
}

func TestCLI(t *testing.T) {
	cleanup := di.Reset

	defer func() {
		cli.DefaultReader = os.Stdin
	}()

	t.Run("from stdin", func(t *testing.T) {
		cases := []struct {
			args          string
			code          int
			useReflection bool
			useWeb        bool
			useTLS        bool
		}{
			{args: "", code: 1},
			{args: "testdata/helloworld.proto", code: 1},
			{args: "--package helloworld testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --call SayHello testdata/helloworld.proto", code: 1},
			{args: "--package helloworld --service Greeter --call SayHello", code: 1},
			{args: "--package helloworld --service Greeter --call SayHello testdata/helloworld.proto"},

			{args: "--reflection", code: 1, useReflection: true},
			{args: "--reflection --service Greeter", code: 1, useReflection: true},
			{args: "--reflection --call SayHello", code: 1, useReflection: true},
			{args: "--reflection --service Greeter --call SayHello", useReflection: true},

			{args: "--web --package helloworld --service Greeter --call SayHello testdata/helloworld.proto", useWeb: true},

			{args: "--web --reflection --package foo", useReflection: true, useWeb: true, code: 1},
			{args: "--web --reflection --service bar", useReflection: true, useWeb: true, code: 1},
			{args: "--web --reflection --service Greeter", useReflection: true, useWeb: true, code: 1},
			{args: "--web --reflection --service Greeter --call SayHello", useReflection: true, useWeb: true},

			{args: "--tls --host localhost --package helloworld --service Greeter --call SayHello testdata/helloworld.proto", useTLS: true},
		}

		for _, c := range cases {
			t.Run(c.args, func(t *testing.T) {
				srv := newServer(t, c.useReflection, c.useTLS)

				defer srv.start(c.useWeb).stop()
				defer cleanup()

				in := strings.NewReader(`{ "name": "maho" }`)
				cli.DefaultReader = in

				out := new(bytes.Buffer)
				errOut := new(bytes.Buffer)
				ui := cui.New(in, out, errOut)

				args := strings.Split(c.args, " ")
				args = append([]string{"--cli", "--port", srv.port}, args...)
				if c.useTLS {
					args = append([]string{"--cacert", "testdata/cert/rootCA.pem"}, args...)
				}
				code := newCommand(ui).Run(args)
				require.Equal(t, c.code, code, errOut.String())

				if c.code == 0 {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()), errOut.String())
				}
			})
		}
	})

	t.Run("from file", func(t *testing.T) {
		cases := []struct {
			args          string
			code          int
			useReflection bool
			useWeb        bool
			useTLS        bool
		}{
			{args: "--file testdata/in.json", code: 1},
			{args: "--file testdata/in.json testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld --service Greeter testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld --call SayHello testdata/helloworld.proto", code: 1},
			{args: "--file testdata/in.json --package helloworld --service Greeter --call SayHello", code: 1},
			{args: "--file testdata/in.json --package helloworld --service Greeter --call SayHello testdata/helloworld.proto"},

			{args: "--reflection --file testdata/in.json", code: 1, useReflection: true},
			{args: "--reflection --file testdata/in.json --service Greeter", code: 1, useReflection: true},
			{args: "--reflection --file testdata/in.json --call SayHello", code: 1, useReflection: true},
			{args: "--reflection --file testdata/in.json --service Greeter --call SayHello", code: 0, useReflection: true},

			{args: "--web --file testdata/in.json --package helloworld --service Greeter --call SayHello testdata/helloworld.proto", useWeb: true},

			{args: "--tls --host localhost --file testdata/in.json --package helloworld --service Greeter --call SayHello testdata/helloworld.proto", useTLS: true},
		}

		for _, c := range cases {
			t.Run(c.args, func(t *testing.T) {
				srv := newServer(t, c.useReflection, c.useTLS)
				defer srv.start(c.useWeb).stop()
				defer cleanup()

				in := strings.NewReader(`{ "name": "maho" }`)
				cli.DefaultReader = in

				out, eout := new(bytes.Buffer), new(bytes.Buffer)
				ui := cui.New(in, out, eout)

				args := append([]string{"--cli", "--port", srv.port}, strings.Split(c.args, " ")...)
				if c.useTLS {
					args = append([]string{"--cacert", "testdata/cert/rootCA.pem"}, args...)
				}
				code := newCommand(ui).Run(args)
				require.Equalf(t, c.code, code, "expected %d, but got %d. out = '%s', errout = '%s'", c.code, code, flatten(out.String()), flatten(eout.String()))

				if c.code == 0 {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
				}
			})
		}
	})
}
