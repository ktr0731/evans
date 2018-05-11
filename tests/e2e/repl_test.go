package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ktr0731/evans/adapter/controller"
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
			{args: "--silent", code: 1},
			{args: "--silent testdata/helloworld.proto", code: 1},
			{args: "--silent --package helloworld testdata/helloworld.proto", code: 1},
			{args: "--silent --package helloworld --service Greeter testdata/helloworld.proto", code: 1},
			{args: "--silent --package helloworld --call SayHello testdata/helloworld.proto", code: 1},
			{args: "--silent --package helloworld --service Greeter --call SayHello", code: 1},
			{args: "--silent --repl --package helloworld --service Greeter testdata/helloworld.proto"},
		}

		for i, c := range cases {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				out := new(bytes.Buffer)
				controller.DefaultREPLUI = &controller.REPLUI{
					UI: controller.NewUI(os.Stdin, out, ioutil.Discard),
				}

				p := helper.NewMockPrompt([]string{
					"call SayHello",
					"maho",
					"exit",
				}, []string{})
				cleanup := SetPrompt(p)
				defer cleanup()

				code := newCLI(controller.NewBasicUI()).Run(strings.Split(c.args, " "))
				require.Equal(t, c.code, code)

				if c.code == 0 {
					assert.Equal(t, `{ "message": "Hello, maho!" }`, flatten(out.String()))
				}
			})
		}
	})
}
