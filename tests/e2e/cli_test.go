// +build e2e

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ktr0731/evans/adapter/cli"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/di"
	"github.com/ktr0731/grpc-test/server"
	"github.com/phayes/freeport"
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

	cases := []struct {
		args string
		code int

		// Server config
		useReflection bool
		useWeb        bool
		useTLS        bool
		specifyCA     bool
	}{
		{args: "", code: 1},
		{args: "testdata/api.proto", code: 1},
		{args: "--package api testdata/api.proto", code: 1},
		{args: "--package api --service Example testdata/api.proto", code: 1},
		{args: "--package api --service Example --call Unary", code: 1},
		{args: "--package api --service Example --call Unary testdata/api.proto"},

		{args: "--reflection", code: 1, useReflection: true},
		{args: "--reflection --package api --service Example", code: 1, useReflection: true},
		{args: "--reflection --service Example", code: 1, useReflection: true},     // Package api is inferred.
		{args: "--reflection --service api.Example", code: 1, useReflection: true}, // Specify package by --service flag.
		{args: "--reflection --call Unary", useReflection: true},                   // Package api and service Example are inferred.
		{args: "--reflection --service Example --call Unary", useReflection: true},

		{args: "--web --package api --service Example --call Unary testdata/api.proto", useWeb: true},

		{args: "--web --reflection", useReflection: true, useWeb: true, code: 1},
		{args: "--web --reflection --service bar", useReflection: true, useWeb: true, code: 1},
		{args: "--web --reflection --service Example", useReflection: true, useWeb: true, code: 1},
		{args: "--web --reflection --service Example --call Unary", useReflection: true, useWeb: true},

		{args: "--tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true},
		{args: "--tls --cert testdata/cert/localhost.pem --certkey testdata/cert/localhost-key.pem -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true},
		// If both of --tls and --insecure are provided, --insecure is ignored.
		{args: "--tls --insecure -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true},
		{args: "--tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, code: 1},

		{args: "--tls --web -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, code: 1},

		{args: "--file testdata/in.json", code: 1},
		{args: "--file testdata/in.json testdata/api.proto", code: 1},
		{args: "--file testdata/in.json --package api testdata/api.proto", code: 1},
		{args: "--file testdata/in.json --package api --service Example testdata/api.proto", code: 1},
		{args: "--file testdata/in.json --package api --service Example --call Unary", code: 1},
		{args: "--file testdata/in.json --package api --service Example --call Unary testdata/api.proto"},

		{args: "--reflection --file testdata/in.json", code: 1, useReflection: true},
		{args: "--reflection --file testdata/in.json --service Example", code: 1, useReflection: true},
		{args: "--reflection --file testdata/in.json --service Example --call Unary", code: 0, useReflection: true},

		{args: "--web --file testdata/in.json --package api --service Example --call Unary testdata/api.proto", useWeb: true},

		{args: "--tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true},
		{args: "--tls --cert testdata/cert/localhost.pem --certkey testdata/cert/localhost-key.pem  -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true},
		{args: "--tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, code: 1},
	}

	for _, c := range cases {
		c := c
		t.Run(c.args, func(t *testing.T) {
			port, err := freeport.GetFreePort()
			require.NoError(t, err, "failed to get a free port for gRPC test server")

			addr := fmt.Sprintf(":%d", port)
			opts := []server.Option{server.WithAddr(addr)}
			if c.useReflection {
				opts = append(opts, server.WithReflection())
			}
			if c.useTLS {
				opts = append(opts, server.WithTLS())
			}
			if c.useWeb {
				opts = append(opts, server.WithProtocol(server.ProtocolImprobableGRPCWeb))
			}

			defer server.New(opts...).Serve().Stop()
			defer cleanup()

			in := strings.NewReader(`{ "name": "maho" }`)
			cli.DefaultReader = in

			out := new(bytes.Buffer)
			errOut := new(bytes.Buffer)
			ui := cui.New(in, out, errOut)

			args := strings.Split(c.args, " ")
			args = append([]string{"--cli", "--port", strconv.Itoa(port)}, args...)
			if c.useTLS && c.specifyCA {
				args = append([]string{"--cacert", "testdata/cert/rootCA.pem"}, args...)
			}
			code := newCommand(ui).Run(args)
			require.Equal(t, c.code, code, errOut.String())

			if c.code == 0 {
				assert.Equal(t, `{ "message": "hello, maho" }`, flatten(out.String()), errOut.String())
			}
		})
	}
}
