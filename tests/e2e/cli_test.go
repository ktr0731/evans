// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
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

const (
	normalIn           = `{ "name": "maho" }`
	normalOut          = `{ "message": "hello, maho" }`
	clientStreamingIn  = `{ "name": "ash" } { "name": "eiji" }`
	clientStreamingOut = `{ "message": "you sent requests 2 times (ash, eiji)." }`
	bidiStreamingIn    = clientStreamingIn
)

func TestCLI(t *testing.T) {
	cleanup := di.Reset

	defer func() {
		cli.DefaultReader = os.Stdin
	}()

	cases := []struct {
		// in specifies the stdin input. Ignored if --file is specified.
		in   string
		args string
		code int

		// Server config
		useReflection bool
		useWeb        bool
		useTLS        bool
		specifyCA     bool

		// Expected output. Ignored if it is empty.
		out string
		// assertOut is called if it isn't nil.
		assertOut func(t *testing.T, out string)
	}{
		{in: normalIn, args: "", code: 1},
		{in: normalIn, args: "testdata/api.proto", code: 1},
		{in: normalIn, args: "--package api testdata/api.proto", code: 1},
		{in: normalIn, args: "--package api --service Example testdata/api.proto", code: 1},
		{in: normalIn, args: "--package api --service Example --call Unary", code: 1},
		{in: normalIn, args: "--package api --service Example --call Unary testdata/api.proto", out: normalOut},
		{in: clientStreamingIn, args: "--package api --service Example --call ClientStreaming testdata/api.proto", out: clientStreamingOut},
		{
			in:   normalIn,
			args: "--package api --service Example --call ServerStreaming testdata/api.proto",
			assertOut: func(t *testing.T, out string) {
				// Transform to a JSON-formed text.
				in := fmt.Sprintf(`[ %s ]`, strings.Replace(out, "} ", "}, ", -1))
				s := []struct {
					Message string `json:"message"`
				}{}
				err := json.Unmarshal([]byte(in), &s)
				require.NoError(t, err, "json.Unmarshal must parse the response message")

				assert.NotZero(t, s, "the response message must have one or more JSON texts, but missing")
				for i, f := range s {
					assert.Equal(t, f.Message, fmt.Sprintf("hello maho, I greet %d times.", i+1), "each message must contain text such that 'hello maho, I greet {n} times.'")
				}
			},
		},
		{
			in:   bidiStreamingIn,
			args: "--package api --service Example --call BidiStreaming testdata/api.proto",
			assertOut: func(t *testing.T, out string) {
				// Transform to a JSON-formed text.
				in := fmt.Sprintf(`[ %s ]`, strings.Replace(out, "} ", "}, ", -1))
				s := []struct {
					Message string `json:"message"`
				}{}
				err := json.Unmarshal([]byte(in), &s)
				require.NoError(t, err, "json.Unmarshal must parse the response message")

				assert.NotZero(t, s, "the response message must have one or more JSON texts, but missing")
				// First, the server greets for "ash" at least one times.
				// After that, the server also greets for "eiji".
				name := "ash"
				var i int
				for _, f := range s {
					if name != "eiji" && strings.HasPrefix(f.Message, "hello eiji, ") {
						name = "eiji"
						i = 0
					}
					assert.Equal(t, f.Message, fmt.Sprintf("hello %s, I greet %d times.", name, i+1), "each message must contain text such that 'hello (ash|eiji), I greet {n} times.'")
					i++
				}
			},
		},

		{in: normalIn, args: "--reflection", code: 1, useReflection: true},
		{in: normalIn, args: "--reflection --package api --service Example", code: 1, useReflection: true},
		{in: normalIn, args: "--reflection --service Example", code: 1, useReflection: true},     // Package api is inferred.
		{in: normalIn, args: "--reflection --service api.Example", code: 1, useReflection: true}, // Specify package by --service flag.
		{in: normalIn, args: "--reflection --call Unary", useReflection: true, out: normalOut},   // Package api and service Example are inferred.
		{in: normalIn, args: "--reflection --service Example --call Unary", useReflection: true, out: normalOut},

		{in: normalIn, args: "--web --package api --service Example --call Unary testdata/api.proto", useWeb: true, out: normalOut},

		{in: normalIn, args: "--web --reflection", useReflection: true, useWeb: true, code: 1},
		{in: normalIn, args: "--web --reflection --service bar", useReflection: true, useWeb: true, code: 1},
		{in: normalIn, args: "--web --reflection --service Example", useReflection: true, useWeb: true, code: 1},
		{in: normalIn, args: "--web --reflection --service Example --call Unary", useReflection: true, useWeb: true, out: normalOut},
		{in: clientStreamingIn, args: "--web --reflection --service Example --call ClientStreaming testdata/api.proto", useReflection: true, useWeb: true, out: clientStreamingOut},
		{
			in:            normalIn,
			args:          "--web --reflection --service Example --call ServerStreaming testdata/api.proto",
			useReflection: true,
			useWeb:        true,
			assertOut: func(t *testing.T, out string) {
				s := toStruct(t, out)
				for i, f := range s {
					assert.Equal(t, f.Message, fmt.Sprintf("hello maho, I greet %d times.", i+1), "each message must contain text such that 'hello maho, I greet {n} times.'")
				}
			},
		},
		{
			in:            bidiStreamingIn,
			args:          "--web --reflection --service Example --call BidiStreaming testdata/api.proto",
			useReflection: true,
			useWeb:        true,
			assertOut: func(t *testing.T, out string) {
				s := toStruct(t, out)
				// First, the server greets for "ash" at least one times.
				// After that, the server also greets for "eiji".
				name := "ash"
				var i int
				for _, f := range s {
					if name != "eiji" && strings.HasPrefix(f.Message, "hello eiji, ") {
						name = "eiji"
						i = 0
					}
					assert.Equal(t, f.Message, fmt.Sprintf("hello %s, I greet %d times.", name, i+1), "each message must contain text such that 'hello (ash|eiji), I greet {n} times.'")
					i++
				}
			},
		},

		{in: normalIn, args: "--tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, out: normalOut},
		{in: normalIn, args: "--tls --cert testdata/cert/localhost.pem --certkey testdata/cert/localhost-key.pem -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, out: normalOut},
		// If both of --tls and --insecure are provided, --insecure is ignored.
		{in: normalIn, args: "--tls --insecure -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, out: normalOut},
		{in: normalIn, args: "--tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, code: 1},

		{in: normalIn, args: "--tls --web -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, code: 1},

		{args: "--file testdata/in.json", code: 1},
		{args: "--file testdata/in.json testdata/api.proto", code: 1},
		{args: "--file testdata/in.json --package api testdata/api.proto", code: 1},
		{args: "--file testdata/in.json --package api --service Example testdata/api.proto", code: 1},
		{args: "--file testdata/in.json --package api --service Example --call Unary", code: 1},
		{args: "--file testdata/in.json --package api --service Example --call Unary testdata/api.proto", out: normalOut},

		{args: "--reflection --file testdata/in.json", code: 1, useReflection: true},
		{args: "--reflection --file testdata/in.json --service Example", code: 1, useReflection: true},
		{args: "--reflection --file testdata/in.json --service Example --call Unary", useReflection: true, out: normalOut},

		{args: "--web --file testdata/in.json --package api --service Example --call Unary testdata/api.proto", useWeb: true, out: normalOut},

		{args: "--file testdata/in.json --tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, out: normalOut},
		{args: "--file testdata/in.json --tls --cert testdata/cert/localhost.pem --certkey testdata/cert/localhost-key.pem  -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, specifyCA: true, out: normalOut},
		{args: "--file testdata/in.json --tls -r --host localhost --service Example --call Unary", useReflection: true, useTLS: true, code: 1},
	}

	for _, c := range cases {
		c := c
		t.Run(c.args, func(t *testing.T) {
			srv, port := newServer(t, c.useReflection, c.useTLS, c.useWeb)
			defer srv.Serve().Stop()
			defer cleanup()

			in := strings.NewReader(c.in)
			cli.DefaultReader = in

			out := new(bytes.Buffer)
			errOut := new(bytes.Buffer)
			ui := cui.New(in, out, errOut)

			args := strings.Split(c.args, " ")
			args = append([]string{"--cli", "--port", port}, args...)
			if c.useTLS && c.specifyCA {
				args = append([]string{"--cacert", "testdata/cert/rootCA.pem"}, args...)
			}
			code := newCommand(ui).Run(args)
			require.Equal(t, c.code, code, errOut.String())

			if c.code == 0 {
				out := flatten(out.String())
				if len(c.out) != 0 {
					assert.Equal(t, c.out, out, errOut.String())
				}
				if c.assertOut != nil {
					c.assertOut(t, out)
				}
			}
		})
	}
}

// toStruct converts a response string to a structured one.
func toStruct(t *testing.T, out string) []struct {
	Message string `json:"message"`
} {
	// Transform to a JSON-formed text.
	in := fmt.Sprintf(`[ %s ]`, strings.Replace(out, "} ", "}, ", -1))
	s := []struct {
		Message string `json:"message"`
	}{}
	err := json.Unmarshal([]byte(in), &s)
	require.NoError(t, err, "json.Unmarshal must parse the response message")
	assert.NotZero(t, s, "the response message must have one or more JSON texts, but missing")
	return s
}
